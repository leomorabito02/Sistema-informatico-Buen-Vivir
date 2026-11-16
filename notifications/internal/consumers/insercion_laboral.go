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

const ConsumerInsercionLaboral = "notificaciones-insercion-laboral"

// NewInsercionLaboralConsumer sets up the push consumer for labor integration domain events.
func NewInsercionLaboralConsumer(
	verifier events.Verifier,
	store events.ProcessedStore,
	registry *rules.Registry,
	resolver audience.Resolver,
	repo repositories.NotificationRepository,
	worker *delivery.Worker,
	logger *slog.Logger,
) *events.Consumer {
	c := events.NewConsumer(ConsumerInsercionLaboral, verifier, store, logger)

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
		events.OfertaCreada,
		events.OfertaRevisada,
		events.PostulacionCreada,
		events.PostulacionDecidida,
		events.OfertaEliminada,
		events.OfertaCompatible,
	}

	for _, et := range eventTypes {
		c.Handle(et, 1, handleEvent)
	}

	return c
}

