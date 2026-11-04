package config

import "os"

type Config struct {
	Port                    string
	NotificationsServiceURL string
	InternalHMACSecret      string
}

func LoadConfig() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	notifURL := os.Getenv("NOTIFICATIONS_SERVICE_URL")
	hmacSecret := os.Getenv("INTERNAL_HMAC_SECRET")

	return &Config{
		Port:                    port,
		NotificationsServiceURL: notifURL,
		InternalHMACSecret:      hmacSecret,
	}
}
