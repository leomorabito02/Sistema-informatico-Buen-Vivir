// Package consumers tiene los handlers de las suscripciones push de Pub/Sub de
// Educación. Cada handler es idempotente.
package consumers

import (
	"context"
	"fmt"
	"log/slog"

	"buenvivir/education/internal/events"
	"buenvivir/education/internal/models"
)

// Nombres de los consumidores (y de sus suscripciones, ver
// infra/pubsub/setup.sh).
const (
	ConsumerUsuarios  = "educacion-usuarios"
	ConsumerContenido = "educacion-contenido"
)

// EnrollmentDeactivator da de baja las inscripciones de un participante.
type EnrollmentDeactivator interface {
	DeactivateByParticipant(ctx context.Context, participanteID, motivo string) (int, error)
}

// CacheInvalidator vence la caché de contenido de Contentful.
type CacheInvalidator interface {
	InvalidateCache()
}

// NewUsuarios consume usuarios-eventos (Autenticación). Cuando un
// participante se da de baja, sus inscripciones activas pasan a "baja"
// conservando todo el historial (HU-09, HU-16).
func NewUsuarios(verifier events.Verifier, store events.ProcessedStore, enrollments EnrollmentDeactivator, logger *slog.Logger) *events.Consumer {
	c := events.NewConsumer(ConsumerUsuarios, verifier, store, logger)
	deactivate := func(ctx context.Context, env events.Envelope, participanteID string) error {
		if participanteID == "" {
			return fmt.Errorf("%s sin participante_id", env.EventType)
		}
		n, err := enrollments.DeactivateByParticipant(ctx, participanteID, models.MotivoBajaParticipanteEliminado)
		if err != nil {
			return err
		}
		slog.InfoContext(ctx, "inscripciones dadas de baja",
			slog.String("event_id", env.EventID), slog.String("participante_id", participanteID), slog.Int("cantidad", n))
		return nil
	}

	c.Handle(events.ParticipanteEliminado, 1, func(ctx context.Context, env events.Envelope) error {
		p, err := events.DecodePayload[events.ParticipantePayload](env)
		if err != nil {
			return err
		}
		return deactivate(ctx, env, p.ParticipanteID)
	})

	// usuario.actualizado solo importa si un participante dejó de estar activo. Si la
	// cuenta se reactiva, las inscripciones no se reactivan solas.
	c.Handle(events.UsuarioActualizado, 1, func(ctx context.Context, env events.Envelope) error {
		p, err := events.DecodePayload[events.UsuarioActualizadoPayload](env)
		if err != nil {
			return err
		}
		if p.ParticipanteID == "" || p.Status == "" || p.Status == "active" {
			return nil
		}
		return deactivate(ctx, env, p.ParticipanteID)
	})
	return c
}

// NewContenido consume contenido-eventos (webhooks de Contentful): un
// cambio publicado o despublicado vence la caché de contenido.
func NewContenido(verifier events.Verifier, store events.ProcessedStore, cache CacheInvalidator, logger *slog.Logger) *events.Consumer {
	c := events.NewConsumer(ConsumerContenido, verifier, store, logger)
	invalidate := func(context.Context, events.Envelope) error {
		cache.InvalidateCache()
		return nil
	}
	c.Handle(events.ContenidoPublicado, 1, invalidate)
	c.Handle(events.ContenidoDespublicado, 1, invalidate)
	return c
}
