package consumers

import (
	"context"
	"log/slog"

	"buenvivir/notifications/internal/audience"
	"buenvivir/notifications/internal/delivery"
	"buenvivir/notifications/internal/events"
	"buenvivir/notifications/internal/repositories"
	"buenvivir/notifications/internal/rules"
)

const ConsumerEducacion = "notificaciones-educacion"

// NewEducacionConsumer sets up the push consumer for education domain events.
func NewEducacionConsumer(
	verifier events.Verifier,
	store events.ProcessedStore,
	registry *rules.Registry,
	resolver audience.Resolver,
	repo repositories.NotificationRepository,
	worker *delivery.Worker,
	logger *slog.Logger,
) *events.Consumer {
	c := events.NewConsumer(ConsumerEducacion, verifier, store, logger)

	handleEvent := func(ctx context.Context, env events.Envelope) error {
		rule, ok := registry.Get(env.EventType)
		if !ok {
			return nil
		}
		plans, err := rule.Build(ctx, env)
		if err != nil {
			return err
		}
		return PersistPlans(ctx, plans, resolver, repo, worker)
	}

	eventTypes := []string{
		events.InscripcionCreada,
		events.ContenidoDocenteSubido,
		events.ContenidoDocenteModerado,
		events.CursoCompletadoParticipante,
		events.HabilidadOtorgada,
		events.CursoEliminado,
		events.CursoCreado,
		events.ActividadRecordatorio,
	}

	for _, et := range eventTypes {
		c.Handle(et, 1, handleEvent)
	}

	return c
}

