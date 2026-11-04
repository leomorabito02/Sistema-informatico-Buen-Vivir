package producers

import (
	"context"

	"buenvivir/labor-integration/internal/events"
	"buenvivir/labor-integration/internal/outbox"
)

// InsercionLaboralProducer defines methods to publish events from labor-integration.
type InsercionLaboralProducer interface {
	OfertaCreada(ctx context.Context, ofertaID, empresaID, estado string) error
	OfertaRevisada(ctx context.Context, ofertaID, empresaID string, aprobada bool) error
	PostulacionCreada(ctx context.Context, postulacionID, ofertaID, participanteID, empresaID string) error
	PostulacionSeleccionada(ctx context.Context, postulacionID, ofertaID, participanteID, empresaID string) error
	PostulacionDecidida(ctx context.Context, postulacionID, ofertaID, participanteID, empresaID, resultado string) error
	Enabled() bool
	Notify()
}

type insercionLaboralProducer struct {
	writer *outbox.Writer
	topic  string
}

// NewInsercionLaboral creates a new InsercionLaboralProducer.
func NewInsercionLaboral(writer *outbox.Writer, topic string) InsercionLaboralProducer {
	if topic == "" {
		topic = events.TopicInsercionLaboral
	}
	return &insercionLaboralProducer{writer: writer, topic: topic}
}

func (p *insercionLaboralProducer) Enabled() bool {
	return p != nil && p.writer != nil
}

func (p *insercionLaboralProducer) Notify() {
	if p.Enabled() {
		p.writer.Notify()
	}
}

func (p *insercionLaboralProducer) OfertaCreada(ctx context.Context, ofertaID, empresaID, estado string) error {
	if !p.Enabled() {
		return nil
	}
	payload := events.OfertaCreadaPayload{
		OfertaID:  ofertaID,
		EmpresaID: empresaID,
		Estado:    estado,
	}
	ev := outbox.Event{
		Topic:       p.topic,
		Type:        events.OfertaCreada,
		Payload:     payload,
		OrderingKey: empresaID,
	}
	return p.writer.Add(ctx, ev)
}

func (p *insercionLaboralProducer) OfertaRevisada(ctx context.Context, ofertaID, empresaID string, aprobada bool) error {
	if !p.Enabled() {
		return nil
	}
	payload := events.OfertaRevisadaPayload{
		OfertaID:  ofertaID,
		EmpresaID: empresaID,
		Aprobada:  aprobada,
	}
	ev := outbox.Event{
		Topic:       p.topic,
		Type:        events.OfertaRevisada,
		Payload:     payload,
		OrderingKey: empresaID,
	}
	return p.writer.Add(ctx, ev)
}

func (p *insercionLaboralProducer) PostulacionCreada(ctx context.Context, postulacionID, ofertaID, participanteID, empresaID string) error {
	if !p.Enabled() {
		return nil
	}
	payload := events.PostulacionCreadaPayload{
		PostulacionID:  postulacionID,
		OfertaID:       ofertaID,
		ParticipanteID: participanteID,
		EmpresaID:      empresaID,
	}
	ev := outbox.Event{
		Topic:       p.topic,
		Type:        events.PostulacionCreada,
		Payload:     payload,
		OrderingKey: participanteID,
	}
	return p.writer.Add(ctx, ev)
}

func (p *insercionLaboralProducer) PostulacionDecidida(ctx context.Context, postulacionID, ofertaID, participanteID, empresaID, resultado string) error {
	if !p.Enabled() {
		return nil
	}
	payload := events.PostulacionDecididaPayload{
		PostulacionID:  postulacionID,
		OfertaID:       ofertaID,
		ParticipanteID: participanteID,
		EmpresaID:      empresaID,
		Resultado:      resultado,
	}
	ev := outbox.Event{
		Topic:       p.topic,
		Type:        events.PostulacionDecidida,
		Payload:     payload,
		OrderingKey: participanteID,
	}
	return p.writer.Add(ctx, ev)
}

func (p *insercionLaboralProducer) PostulacionSeleccionada(ctx context.Context, postulacionID, ofertaID, participanteID, empresaID string) error {
	if !p.Enabled() {
		return nil
	}
	payload := events.PostulacionSeleccionadaPayload{
		PostulacionID:  postulacionID,
		OfertaID:       ofertaID,
		ParticipanteID: participanteID,
		EmpresaID:      empresaID,
	}
	ev := outbox.Event{
		Topic:       p.topic,
		Type:        events.PostulacionSeleccionada,
		Payload:     payload,
		OrderingKey: participanteID,
	}
	return p.writer.Add(ctx, ev)
}

