package config

import (
	"os"
	"strings"
)

// Config agrupa la configuración del servicio de Promoción Humana.
type Config struct {
	Port string

	CORSAllowedOrigins string

	// EncryptionKeySecret es el secreto de Secret Manager con la clave AES-256 (base64)
	// que cifra situacion_actual, avances y la descripción de los hitos. Es la forma de Cloud Run.
	EncryptionKeySecret string

	// EncryptionKeyBase64 es la alternativa local: la misma clave en base64 en la
	// variable. EncryptionKeySecret tiene prioridad.
	EncryptionKeyBase64 string

	// AuthServiceURL es la base HTTP de auth-service (ej. "https://auth-prod-xxx.run.app").
	// Si está configurada, tiene prioridad sobre AuthServiceGRPCTarget.
	AuthServiceURL string

	// UpstreamIDToken activa la generación de un token de identidad de Google
	// en X-Serverless-Authorization para llamadas entre servicios Cloud Run.
	UpstreamIDToken bool

	// AuthServiceGRPCTarget es la dirección gRPC de auth-service (ej. "auth-service:50051").
	// De ahí salen los datos y la búsqueda de participantes.
	AuthServiceGRPCTarget string

	// EducationServiceURL es la base de education-service, de donde salen las habilidades
	// del participante. Opcional: vacía, la ficha se muestra sin habilidades.
	EducationServiceURL        string
	EducationServiceGRPCTarget string

	// Events es la mensajería asíncrona con Pub/Sub (ver
	// docs/eventos/README.md). Desactivada si falta PUBSUB_PROJECT_ID.
	Events EventsConfig
}

// EventsConfig: misma convención de variables que education.
type EventsConfig struct {
	// Production es APP_ENV=production: ahí no se puede desactivar la
	// verificación de los push.
	Production          bool
	ProjectID           string
	TopicPromocion      string
	PushAuth            string
	PushAudience        string
	PushServiceAccounts []string
}

// PublishEnabled dice si el servicio publica eventos.
func (e EventsConfig) PublishEnabled() bool { return e.ProjectID != "" }

// ConsumeEnabled dice si hay configuración para recibir push.
func (e EventsConfig) ConsumeEnabled() bool {
	return e.PushAuth == "disabled" || (e.PushAudience != "" && len(e.PushServiceAccounts) > 0)
}

func loadEventsConfig() EventsConfig {
	topic := strings.TrimSpace(os.Getenv("PUBSUB_TOPIC_PROMOCION_HUMANA"))
	if topic == "" {
		topic = "promocion-humana-eventos"
	}
	var accounts []string
	for _, a := range strings.Split(os.Getenv("PUBSUB_PUSH_SERVICE_ACCOUNTS"), ",") {
		if a = strings.TrimSpace(a); a != "" {
			accounts = append(accounts, a)
		}
	}
	return EventsConfig{
		Production:          strings.TrimSpace(os.Getenv("APP_ENV")) == "production",
		ProjectID:           strings.TrimSpace(os.Getenv("PUBSUB_PROJECT_ID")),
		TopicPromocion:      topic,
		PushAuth:            strings.TrimSpace(os.Getenv("PUBSUB_PUSH_AUTH")),
		PushAudience:        strings.TrimSpace(os.Getenv("PUBSUB_PUSH_AUDIENCE")),
		PushServiceAccounts: accounts,
	}
}

// LoadConfig arma la configuración a partir de variables de entorno.
// La validez de la clave de cifrado no se chequea acá sino en
// crypto.LoadKey, que es quien sabe resolverla (Secret Manager o env).
func LoadConfig() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8085"
	}

	return &Config{
		Port:                port,
		CORSAllowedOrigins:  os.Getenv("CORS_ALLOWED_ORIGINS"),
		EncryptionKeySecret: strings.TrimSpace(os.Getenv("ENCRYPTION_KEY_SECRET")),
		EncryptionKeyBase64: strings.TrimSpace(os.Getenv("ENCRYPTION_KEY")),

		AuthServiceURL:             strings.TrimRight(strings.TrimSpace(os.Getenv("AUTH_SERVICE_URL")), "/"),
		AuthServiceGRPCTarget:      strings.TrimSpace(os.Getenv("AUTH_SERVICE_GRPC_TARGET")),
		EducationServiceURL:        strings.TrimSpace(os.Getenv("EDUCATION_SERVICE_URL")),
		EducationServiceGRPCTarget: strings.TrimSpace(os.Getenv("EDUCATION_SERVICE_GRPC_TARGET")),
		UpstreamIDToken:            strings.EqualFold(strings.TrimSpace(os.Getenv("UPSTREAM_ID_TOKEN")), "true"),
		Events:                     loadEventsConfig(),
	}
}
