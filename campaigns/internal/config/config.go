package config

import (
	"os"
	"strings"
)

// DefaultPort es el puerto del servicio si no se define PORT (el mismo
// que declaraban el Dockerfile y el .env.example del esqueleto).
const DefaultPort = "8080"

// Config agrupa la configuración del servicio de Campañas.
type Config struct {
	Port string

	// CORSAllowedOrigins: orígenes del frontend separados por coma. Vacío
	// = no se habilita CORS (el acceso normal es por el API Gateway).
	CORSAllowedOrigins string

	// NotificationsServiceURL: URL base del servicio de notificaciones.
	NotificationsServiceURL string

	// InternalHMACSecret: Secreto compartido para autenticar eventos con notifications-service.
	InternalHMACSecret string

	// Events es la configuración para publicar eventos en Pub/Sub vía outbox.
	Events EventsConfig
}

type EventsConfig struct {
	ProjectID     string
	TopicCampanas string
}

func (e EventsConfig) PublishEnabled() bool {
	return e.ProjectID != ""
}

// LoadConfig lee la configuración del entorno. JWT_SECRET y las
// credenciales de Firestore las leen directamente JWTService y
// NewFirestoreClient, con el mismo criterio que human-promotion.
func LoadConfig() *Config {
	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = DefaultPort
	}
	notifURL := strings.TrimSpace(os.Getenv("NOTIFICATIONS_SERVICE_URL"))
	topic := strings.TrimSpace(os.Getenv("PUBSUB_TOPIC_CAMPANAS"))
	if topic == "" {
		topic = "campanas-eventos"
	}
	return &Config{
		Port:                    port,
		CORSAllowedOrigins:      os.Getenv("CORS_ALLOWED_ORIGINS"),
		NotificationsServiceURL: notifURL,
		InternalHMACSecret:      os.Getenv("INTERNAL_HMAC_SECRET"),
		Events: EventsConfig{
			ProjectID:     strings.TrimSpace(os.Getenv("PUBSUB_PROJECT_ID")),
			TopicCampanas: topic,
		},
	}
}
