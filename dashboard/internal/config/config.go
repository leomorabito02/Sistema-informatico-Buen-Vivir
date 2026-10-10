package config

import (
	"os"
	"time"
)

// Config del servicio de Dashboard. No tiene base de datos: solo las
// direcciones de los servicios que consulta (dentro de la red de Docker).
type Config struct {
	Port         string
	AuthURL      string
	EducationURL string
	LaborURL     string
	// Tiempo máximo de cada consulta a un servicio de origen.
	Timeout time.Duration
}

func LoadConfig() *Config {
	return &Config{
		Port:         env("PORT", "8090"),
		AuthURL:      env("AUTH_SERVICE_URL", "http://auth-service:8080"),
		EducationURL: env("EDUCATION_SERVICE_URL", "http://education-service:8210"),
		LaborURL:     env("LABOR_SERVICE_URL", "http://labor-integration:8080"),
		Timeout:      15 * time.Second,
	}
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
