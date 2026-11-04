package producers

import (
	"context"

	"buenvivir/campaigns/internal/events"
	"buenvivir/campaigns/internal/outbox"
)

// CampanasProducer defines methods to publish events from campaigns service.
type CampanasProducer interface {
	CampanaCreada(ctx context.Context, campanaID, creadoPor string) error
	Enabled() bool
	Notify()
}

type campanasProducer struct {
	writer *outbox.Writer
	topic  string
}

// NewCampanasProducer creates a new CampanasProducer.
func NewCampanasProducer(writer *outbox.Writer, topic string) CampanasProducer {
	if topic == "" {
		topic = events.TopicCampanas
	}
	return &campanasProducer{writer: writer, topic: topic}
}

func (p *campanasProducer) Enabled() bool {
	return p != nil && p.writer != nil
}

func (p *campanasProducer) Notify() {
	if p.Enabled() {
		p.writer.Notify()
	}
}

func (p *campanasProducer) CampanaCreada(ctx context.Context, campanaID, creadoPor string) error {
	if !p.Enabled() {
		return nil
	}
	payload := events.CampanaCreadaPayload{
		CampanaID: campanaID,
		CreadoPor: creadoPor,
	}
	ev := outbox.Event{
		Topic:       p.topic,
		Type:        events.CampanaCreada,
		Payload:     payload,
		OrderingKey: campanaID,
	}
	return p.writer.Add(ctx, ev)
}

