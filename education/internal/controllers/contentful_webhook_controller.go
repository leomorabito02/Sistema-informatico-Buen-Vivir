package controllers

import (
	"context"
	"crypto/subtle"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"buenvivir/education/internal/events"
	"buenvivir/education/internal/events/producers"
)

// HeaderWebhookSecret es el header que se configura en el webhook de
// Contentful (Settings → Webhooks → Headers, marcado como "secret").
// #nosec G101 -- header name constant, not a credential
const HeaderWebhookSecret = "X-Webhook-Secret"

// maxWebhookBody: el cuerpo de un webhook de entry es de unos pocos KB.
const maxWebhookBody = 1 << 20

// ContenidoPublisher deja el evento de contenido en el outbox.
type ContenidoPublisher interface {
	Contenido(ctx context.Context, eventID, eventType string, payload events.ContenidoPayload) error
}

// ContentfulWebhookController recibe los webhooks de Contentful y los encola como
// eventos de contenido.
type ContentfulWebhookController struct {
	secret    []byte
	publisher ContenidoPublisher
}

func NewContentfulWebhookController(secret string, publisher ContenidoPublisher) *ContentfulWebhookController {
	return &ContentfulWebhookController{secret: []byte(secret), publisher: publisher}
}

func (c *ContentfulWebhookController) Receive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	got := []byte(r.Header.Get(HeaderWebhookSecret))
	if len(c.secret) == 0 || subtle.ConstantTimeCompare(got, c.secret) != 1 {
		slog.WarnContext(r.Context(), "webhook de Contentful rechazado: secreto inválido")
		writeErrorJSON(w, http.StatusUnauthorized, "no autorizado")
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, maxWebhookBody))
	if err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "cuerpo ilegible")
		return
	}
	topic := r.Header.Get("X-Contentful-Topic")
	hook, ok, err := producers.ParseContentfulWebhook(topic, body)
	if errors.Is(err, producers.ErrWebhookInvalido) {
		writeErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	if !ok {
		writeJSON(w, http.StatusOK, map[string]any{"ignorado": true, "topic": topic})
		return
	}

	if err := c.publisher.Contenido(r.Context(), hook.EventID, hook.EventType, hook.Payload); err != nil {
		// 5xx: Contentful reintenta el webhook más tarde.
		slog.ErrorContext(r.Context(), "no se pudo encolar el evento de contenido",
			slog.String("event_id", hook.EventID), slog.String("event_type", hook.EventType),
			slog.String("correlation_id", events.CorrelationIDFrom(r.Context())),
			slog.String("resultado", "error"), slog.String("error", err.Error()))
		writeErrorJSON(w, http.StatusServiceUnavailable, "no se pudo registrar el evento")
		return
	}
	slog.InfoContext(r.Context(), "webhook de Contentful encolado",
		slog.String("event_id", hook.EventID), slog.String("event_type", hook.EventType),
		slog.String("entry_id", hook.Payload.EntryID),
		slog.String("correlation_id", events.CorrelationIDFrom(r.Context())),
		slog.String("resultado", "encolado"))
	writeJSON(w, http.StatusAccepted, map[string]string{"event_id": hook.EventID})
}
