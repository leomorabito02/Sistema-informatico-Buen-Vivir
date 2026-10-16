package config

import (
	"os"
	"strings"
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

	// CacheTTL es cuánto se reusa lo que devolvió cada fuente (los
	// eventos de Pub/Sub la invalidan antes). 0 desactiva la caché.
	CacheTTL time.Duration

	// Push de Pub/Sub (ver docs/eventos/README.md). Sin audience ni
	// service accounts el endpoint /eventos no se expone, salvo
	// PUBSUB_PUSH_AUTH=disabled (solo local, con el emulador).
	Production          bool
	PushAuth            string
	PushAudience        string
	PushServiceAccounts []string
}

// ConsumeEnabled dice si hay configuración para recibir push.
func (c *Config) ConsumeEnabled() bool {
	return c.PushAuth == "disabled" || (c.PushAudience != "" && len(c.PushServiceAccounts) > 0)
}

func LoadConfig() *Config {
	return &Config{
		Port:         env("PORT", "8090"),
		AuthURL:      env("AUTH_SERVICE_URL", "http://auth-service:8080"),
		EducationURL: env("EDUCATION_SERVICE_URL", "http://education-service:8210"),
		LaborURL:     env("LABOR_SERVICE_URL", "http://labor-integration:8080"),
		Timeout:      15 * time.Second,

		CacheTTL:            duration("DASHBOARD_CACHE_TTL", 5*time.Minute),
		Production:          os.Getenv("APP_ENV") == "production",
		PushAuth:            strings.TrimSpace(os.Getenv("PUBSUB_PUSH_AUTH")),
		PushAudience:        strings.TrimSpace(os.Getenv("PUBSUB_PUSH_AUDIENCE")),
		PushServiceAccounts: list(os.Getenv("PUBSUB_PUSH_SERVICE_ACCOUNTS")),
	}
}

func duration(key string, def time.Duration) time.Duration {
	if d, err := time.ParseDuration(strings.TrimSpace(os.Getenv(key))); err == nil && d >= 0 {
		return d
	}
	return def
}

func list(raw string) []string {
	var out []string
	for _, v := range strings.Split(raw, ",") {
		if v = strings.TrimSpace(v); v != "" {
			out = append(out, v)
		}
	}
	return out
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
