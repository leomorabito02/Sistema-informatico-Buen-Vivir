// Package consumers tiene los consumidores de eventos de Promoción
// Humana (suscripciones push de Pub/Sub). Cada handler es idempotente.
package consumers

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"buenvivir/human-promotion/internal/events"
	"buenvivir/human-promotion/internal/models"
	"buenvivir/human-promotion/internal/repositories"
)

const (
	ConsumerEducacion = "promocion-humana-educacion"
	ConsumerUsuarios  = "promocion-humana-usuarios"
)

// LineaTiempo agrega entradas a la línea de tiempo del participante.
type LineaTiempo interface {
	Add(ctx context.Context, eventID string, e repositories.EntradaLineaTiempo) error
}

// FichaDeactivator da de baja la ficha de un participante.
type FichaDeactivator interface {
	DeactivateByParticipante(ctx context.Context, participanteID string) (bool, error)
}

func fecha(t time.Time) string { return t.UTC().Format(models.FechaLayout) }

// NewEducacion consume educacion-eventos: inscripciones, cursos
// completados y asistencia al centro pasan a la línea de tiempo del
// participante (HU-33, HU-34, HU-19bis). Los textos no llevan números
// ni porcentajes.
func NewEducacion(verifier events.Verifier, store events.ProcessedStore, linea LineaTiempo, logger *slog.Logger) *events.Consumer {
	c := events.NewConsumer(ConsumerEducacion, verifier, store, logger)

	c.Handle(events.InscripcionCreada, 1, func(ctx context.Context, env events.Envelope) error {
		p, err := events.DecodePayload[events.InscripcionCreadaPayload](env)
		if err != nil {
			return err
		}
		return add(ctx, linea, env, p.ParticipanteID, repositories.EntradaLineaTiempo{
			Tipo: repositories.LineaTiempoInscripcion, Descripcion: "Se inscribió en un curso",
			Fecha: fecha(p.FechaInscripcion), CursoID: p.CursoID,
		})
	})

	c.Handle(events.CursoCompletadoParticipante, 1, func(ctx context.Context, env events.Envelope) error {
		p, err := events.DecodePayload[events.CursoCompletadoPayload](env)
		if err != nil {
			return err
		}
		return add(ctx, linea, env, p.ParticipanteID, repositories.EntradaLineaTiempo{
			Tipo: repositories.LineaTiempoCursoCompletado, Descripcion: "Completó un curso",
			Fecha: fecha(p.CompletadoEn), CursoID: p.CursoID,
		})
	})

	// Solo la presencia suma a la trayectoria: una ausencia no se anota
	// en la línea de tiempo que ve el participante.
	c.Handle(events.AsistenciaGeneralRegistrada, 1, func(ctx context.Context, env events.Envelope) error {
		p, err := events.DecodePayload[events.AsistenciaGeneralPayload](env)
		if err != nil {
			return err
		}
		if !p.Presente {
			return nil
		}
		return add(ctx, linea, env, p.ParticipanteID, repositories.EntradaLineaTiempo{
			Tipo: repositories.LineaTiempoAsistenciaGeneral, Descripcion: "Asistió al centro", Fecha: p.Fecha,
		})
	})
	return c
}

func add(ctx context.Context, linea LineaTiempo, env events.Envelope, participanteID string, e repositories.EntradaLineaTiempo) error {
	if participanteID == "" {
		return fmt.Errorf("%s sin participante_id", env.EventType)
	}
	e.ParticipanteID = participanteID
	e.EventType = env.EventType
	e.CreadoEn = time.Now().UTC()
	return linea.Add(ctx, env.EventID, e)
}

// NewUsuarios consume usuarios-eventos (Autenticación): la baja de un
// participante da de baja (lógica) su ficha, sin borrar nada.
func NewUsuarios(verifier events.Verifier, store events.ProcessedStore, fichas FichaDeactivator, logger *slog.Logger) *events.Consumer {
	c := events.NewConsumer(ConsumerUsuarios, verifier, store, logger)
	c.Handle(events.ParticipanteEliminado, 1, func(ctx context.Context, env events.Envelope) error {
		p, err := events.DecodePayload[events.ParticipantePayload](env)
		if err != nil {
			return err
		}
		if p.ParticipanteID == "" {
			return fmt.Errorf("%s sin participante_id", env.EventType)
		}
		changed, err := fichas.DeactivateByParticipante(ctx, p.ParticipanteID)
		if err != nil {
			return err
		}
		slog.InfoContext(ctx, "ficha dada de baja", slog.String("event_id", env.EventID),
			slog.String("participante_id", p.ParticipanteID), slog.Bool("cambio", changed))
		return nil
	})
	return c
}
