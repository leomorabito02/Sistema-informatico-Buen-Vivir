package producers

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"buenvivir/education/internal/events"
)

// ContentfulWebhook es un webhook de Contentful ya traducido a evento.
type ContentfulWebhook struct {
	EventID   string
	EventType string
	Payload   events.ContenidoPayload
}

// ErrWebhookInvalido: el cuerpo no trae lo mínimo (sys.id).
var ErrWebhookInvalido = errors.New("webhook de Contentful inválido")

// contentfulActions mapea la acción del header X-Contentful-Topic
// ("ContentManagement.Entry.<acción>") al evento que se publica.
var contentfulActions = map[string]string{
	"publish":   events.ContenidoPublicado,
	"unpublish": events.ContenidoDespublicado,
	"delete":    events.ContenidoDespublicado,
}

type webhookBody struct {
	Sys struct {
		ID          string `json:"id"`
		Revision    int    `json:"revision"`
		Version     int    `json:"version"`
		UpdatedAt   string `json:"updatedAt"`
		DeletedAt   string `json:"deletedAt"`
		ContentType struct {
			Sys struct {
				ID string `json:"id"`
			} `json:"sys"`
		} `json:"contentType"`
		Environment struct {
			Sys struct {
				ID string `json:"id"`
			} `json:"sys"`
		} `json:"environment"`
	} `json:"sys"`
}

// ParseContentfulWebhook traduce un webhook de Contentful (ok = false si el topic no
// interesa). El event_id es determinístico (entry + acción + revisión) para descartar
// los reintentos.
func ParseContentfulWebhook(topic string, body []byte) (ContentfulWebhook, bool, error) {
	parts := strings.Split(topic, ".")
	if len(parts) != 3 || parts[0] != "ContentManagement" || parts[1] != "Entry" {
		return ContentfulWebhook{}, false, nil
	}
	accion := parts[2]
	eventType, ok := contentfulActions[accion]
	if !ok {
		return ContentfulWebhook{}, false, nil
	}

	var b webhookBody
	if err := json.Unmarshal(body, &b); err != nil {
		return ContentfulWebhook{}, false, fmt.Errorf("%w: %v", ErrWebhookInvalido, err)
	}
	if strings.TrimSpace(b.Sys.ID) == "" {
		return ContentfulWebhook{}, false, fmt.Errorf("%w: falta sys.id", ErrWebhookInvalido)
	}

	revision := b.Sys.Revision
	marca := fmt.Sprint(revision)
	if revision == 0 {
		// unpublish/delete no siempre traen revision: se usa la versión o
		// la fecha del cambio para distinguir un hecho de otro.
		marca = fmt.Sprintf("v%d|%s|%s", b.Sys.Version, b.Sys.UpdatedAt, b.Sys.DeletedAt)
	}
	return ContentfulWebhook{
		EventID:   events.NameUUID("contentful|" + b.Sys.ID + "|" + accion + "|" + marca),
		EventType: eventType,
		Payload: events.ContenidoPayload{
			EntryID:     b.Sys.ID,
			ContentType: b.Sys.ContentType.Sys.ID,
			Environment: b.Sys.Environment.Sys.ID,
			Accion:      accion,
			Revision:    revision,
		},
	}, true, nil
}
