package config

import (
	"os"
	"strings"
	"time"
)

// Config del servicio de Dashboard. Los indicadores se calculan en vivo
// con lo que devuelven los demás servicios; lo único que guarda es el
// historial de reportes exportados (HU-55).
type Config struct {
	Port         string
	AuthURL      string
	EducationURL string
	LaborURL     string
	// HumanPromotionURL: fichas e hitos para los indicadores de
	// Promoción Humana. Vacía = ese indicador responde "no configurado".
	HumanPromotionURL string

	// JWTSecret: con él el Dashboard valida el token por su cuenta
	// (obligatorio en Google Cloud). Vacío = confía en el X-User-Role que
	// pone el gateway de Docker.
	JWTSecret string
	// UpstreamIDToken: pedir un token de identidad de Google para llamar
	// a los otros servicios (Cloud Run con --no-allow-unauthenticated).
	UpstreamIDToken bool
	// ReportesBucket: bucket de Cloud Storage para los PDF exportados.
	// Vacío = el historial no se guarda (responde 503).
	ReportesBucket string
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
		Port:              env("PORT", "8090"),
		AuthURL:           env("AUTH_SERVICE_URL", "http://auth-service:8080"),
		EducationURL:      env("EDUCATION_SERVICE_URL", "http://education-service:8210"),
		LaborURL:          env("LABOR_SERVICE_URL", "http://labor-integration:8080"),
		HumanPromotionURL: env("HUMAN_PROMOTION_SERVICE_URL", "http://human-promotion:8085"),
		Timeout:           15 * time.Second,

		JWTSecret:       os.Getenv("JWT_SECRET"),
		UpstreamIDToken: strings.EqualFold(strings.TrimSpace(os.Getenv("UPSTREAM_ID_TOKEN")), "true"),
		ReportesBucket:  strings.TrimSpace(os.Getenv("REPORTES_BUCKET")),

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

// env devuelve la variable si está definida, aunque esté vacía: en Google
// Cloud LABOR_SERVICE_URL="" significa "no desplegado", no "usar la de
// Docker".
func env(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return strings.TrimSpace(v)
	}
	return def
}
