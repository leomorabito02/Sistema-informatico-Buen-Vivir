// Package consumers tiene el consumidor de eventos de Dashboard: cada
// evento que cambia un indicador (HU-52) invalida la caché de esa fuente.
package consumers

import (
	"context"
	"log/slog"

	"buenvivir/dashboard/internal/events"
	"buenvivir/dashboard/internal/services"
)

// ConsumerDashboard es el nombre del consumidor (las tres suscripciones
// de Dashboard llegan al mismo endpoint).
const ConsumerDashboard = "dashboard"

// Invalidator vence la caché de las fuentes indicadas.
type Invalidator interface {
	Invalidate(fuentes ...services.Fuente)
}

// fuentesPorEvento dice qué indicador cambia con cada evento.
var fuentesPorEvento = map[string]services.Fuente{
	events.ParticipanteCreado:          services.FuenteParticipantes,
	events.ParticipanteEliminado:       services.FuenteParticipantes,
	events.CursoCreado:                 services.FuenteCursos,
	events.CursoEliminado:              services.FuenteCursos,
	events.CursoCompletadoParticipante: services.FuenteCursos,
	events.HabilidadOtorgada:           services.FuenteHabilidades,
	events.PostulacionSeleccionada:     services.FuentePostulaciones,
}

// NewDashboard consume usuarios-eventos, educacion-eventos e
// insercion-laboral-eventos. Invalidar es idempotente por naturaleza
// (hacerlo dos veces no cambia nada), así que alcanza con descartar
// duplicados en memoria: Dashboard no tiene base de datos.
func NewDashboard(verifier events.Verifier, store events.ProcessedStore, cache Invalidator, logger *slog.Logger) *events.Consumer {
	c := events.NewConsumer(ConsumerDashboard, verifier, store, logger)
	for eventType, fuente := range fuentesPorEvento {
		fuente := fuente
		c.Handle(eventType, 1, func(context.Context, events.Envelope) error {
			cache.Invalidate(fuente)
			return nil
		})
	}
	return c
}
