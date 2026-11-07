package config

import (
	"os"
	"strings"
)

type Config struct {
	Port                    string
	AppEnv                  string
	CORSAllowedOrigins      string
	JWTSecret               string
	AuthServiceURL          string
	AuthServiceGRPCTarget   string
	UpstreamIDToken         bool
	NotificationsServiceURL string
	InternalHMACSecret      string
	Events                  EventsConfig
}

type EventsConfig struct {
	ProjectID             string
	TopicInsercionLaboral string
}

func (e EventsConfig) PublishEnabled() bool {
	return e.ProjectID != ""
}

func LoadConfig() *Config {
	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8080"
	}
	appEnv := strings.TrimSpace(os.Getenv("APP_ENV"))
	authServiceURL := strings.TrimRight(strings.TrimSpace(os.Getenv("AUTH_SERVICE_URL")), "/")
	authGRPCTarget := strings.TrimSpace(os.Getenv("AUTH_SERVICE_GRPC_TARGET"))
	if authGRPCTarget == "" && authServiceURL == "" && appEnv != "production" && appEnv != "testing" {
		authGRPCTarget = "auth-service:50051"
	}

	notifURL := strings.TrimSpace(os.Getenv("NOTIFICATIONS_SERVICE_URL"))
	hmacSecret := os.Getenv("INTERNAL_HMAC_SECRET")

	topic := strings.TrimSpace(os.Getenv("PUBSUB_TOPIC_INSERCION_LABORAL"))
	if topic == "" {
		topic = "insercion-laboral-eventos"
	}

	return &Config{
		Port:                    port,
		AppEnv:                  appEnv,
		CORSAllowedOrigins:      os.Getenv("CORS_ALLOWED_ORIGINS"),
		JWTSecret:               strings.TrimSpace(os.Getenv("JWT_SECRET")),
		AuthServiceURL:          authServiceURL,
		AuthServiceGRPCTarget:   authGRPCTarget,
		UpstreamIDToken:         strings.EqualFold(strings.TrimSpace(os.Getenv("UPSTREAM_ID_TOKEN")), "true"),
		NotificationsServiceURL: notifURL,
		InternalHMACSecret:      hmacSecret,
		Events: EventsConfig{
			ProjectID:             strings.TrimSpace(os.Getenv("PUBSUB_PROJECT_ID")),
			TopicInsercionLaboral: topic,
		},
	}
}
