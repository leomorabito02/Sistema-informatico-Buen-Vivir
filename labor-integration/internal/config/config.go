package config

import (
	"os"
	"strings"
)

type Config struct {
	Port                    string
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
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	notifURL := os.Getenv("NOTIFICATIONS_SERVICE_URL")
	hmacSecret := os.Getenv("INTERNAL_HMAC_SECRET")

	topic := strings.TrimSpace(os.Getenv("PUBSUB_TOPIC_INSERCION_LABORAL"))
	if topic == "" {
		topic = "insercion-laboral-eventos"
	}

	return &Config{
		Port:                    port,
		NotificationsServiceURL: notifURL,
		InternalHMACSecret:      hmacSecret,
		Events: EventsConfig{
			ProjectID:             strings.TrimSpace(os.Getenv("PUBSUB_PROJECT_ID")),
			TopicInsercionLaboral: topic,
		},
	}
}
