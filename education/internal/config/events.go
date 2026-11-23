package config

import (
	"os"
	"strings"
	"time"
)

// EventsConfig es la configuración de Pub/Sub. Nada es obligatorio: sin
// PUBSUB_PROJECT_ID el servicio no publica eventos.
type EventsConfig struct {
	// ProjectID del proyecto de GCP donde están los topics. Vacío =
	// publicación desactivada.
	ProjectID string

	TopicEducacion string
	TopicContenido string

	// PushAuth = "disabled" desactiva la verificación OIDC de los push
	// (solo local con el emulador; en producción no arranca).
	PushAuth string
	// PushAudience es la audience configurada en las suscripciones push
	// de este servicio (su URL de Cloud Run).
	PushAudience string
	// PushServiceAccounts son las service accounts autorizadas a mandar
	// push a este servicio (una por suscripción/consumidor).
	PushServiceAccounts []string

	// ContentfulWebhookSecret es el valor que Contentful manda en el
	// header X-Webhook-Secret. Vacío = webhook deshabilitado.
	ContentfulWebhookSecret string

	// ContentCacheTTL es cuánto se considera fresca una respuesta de la
	// Content Delivery API. 0 desactiva la caché.
	ContentCacheTTL time.Duration
}

// PublishEnabled dice si el servicio publica eventos.
func (e EventsConfig) PublishEnabled() bool { return e.ProjectID != "" }

// ConsumeEnabled dice si hay configuración para recibir push.
func (e EventsConfig) ConsumeEnabled() bool {
	return e.PushAuth == "disabled" || (e.PushAudience != "" && len(e.PushServiceAccounts) > 0)
}

func loadEventsConfig() EventsConfig {
	ttl := 5 * time.Minute
	if raw := strings.TrimSpace(os.Getenv("CONTENT_CACHE_TTL")); raw != "" {
		if d, err := time.ParseDuration(raw); err == nil && d >= 0 {
			ttl = d
		}
	}
	return EventsConfig{
		ProjectID:               strings.TrimSpace(os.Getenv("PUBSUB_PROJECT_ID")),
		TopicEducacion:          envOr("PUBSUB_TOPIC_EDUCACION", "educacion-eventos"),
		TopicContenido:          envOr("PUBSUB_TOPIC_CONTENIDO", "contenido-eventos"),
		PushAuth:                strings.TrimSpace(os.Getenv("PUBSUB_PUSH_AUTH")),
		PushAudience:            strings.TrimSpace(os.Getenv("PUBSUB_PUSH_AUDIENCE")),
		PushServiceAccounts:     splitList(os.Getenv("PUBSUB_PUSH_SERVICE_ACCOUNTS")),
		ContentfulWebhookSecret: os.Getenv("CONTENTFUL_WEBHOOK_SECRET"),
		ContentCacheTTL:         ttl,
	}
}

func splitList(raw string) []string {
	var out []string
	for _, v := range strings.Split(raw, ",") {
		if v = strings.TrimSpace(v); v != "" {
			out = append(out, v)
		}
	}
	return out
}
