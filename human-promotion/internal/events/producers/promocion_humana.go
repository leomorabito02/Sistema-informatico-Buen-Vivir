// Package producers arma los eventos que publica Promoción Humana y los
// deja en el outbox. Los payloads llevan SOLO ids y metadatos: nunca la
// situación personal, los avances ni la descripción de un hito.
package producers

import (
	"context"

	"cloud.google.com/go/firestore"

	"buenvivir/human-promotion/internal/events"
	"buenvivir/human-promotion/internal/outbox"
)

// PromocionHumana publica los eventos del servicio. Un *PromocionHumana
// nil (o sin writer) no publica nada.
type PromocionHumana struct {
	writer *outbox.Writer
	topic  string
}

func NewPromocionHumana(writer *outbox.Writer, topic string) *PromocionHumana {
	return &PromocionHumana{writer: writer, topic: topic}
}

func (p *PromocionHumana) Enabled() bool { return p != nil && p.writer != nil }

// Notify avisa al publicador después del commit.
func (p *PromocionHumana) Notify() {
	if p.Enabled() {
		p.writer.Notify()
	}
}

func (p *PromocionHumana) emit(ctx context.Context, tx *firestore.Transaction, eventType string, payload any, participanteID string) error {
	if !p.Enabled() {
		return nil
	}
	return p.writer.AddTx(ctx, tx, outbox.Event{Topic: p.topic, Type: eventType, Payload: payload, OrderingKey: participanteID})
}

func (p *PromocionHumana) FichaCreada(ctx context.Context, tx *firestore.Transaction, fichaID, participanteID string) error {
	return p.emit(ctx, tx, events.FichaCreada, events.FichaPayload{FichaID: fichaID, ParticipanteID: participanteID}, participanteID)
}

func (p *PromocionHumana) FichaActualizada(ctx context.Context, tx *firestore.Transaction, fichaID, participanteID string) error {
	return p.emit(ctx, tx, events.FichaActualizada, events.FichaPayload{FichaID: fichaID, ParticipanteID: participanteID}, participanteID)
}

// Hito publica hito.registrado (nuevo = true) o hito.actualizado.
func (p *PromocionHumana) Hito(ctx context.Context, tx *firestore.Transaction, nuevo bool, h events.HitoPayload) error {
	eventType := events.HitoActualizado
	if nuevo {
		eventType = events.HitoRegistrado
	}
	return p.emit(ctx, tx, eventType, h, h.ParticipanteID)
}
