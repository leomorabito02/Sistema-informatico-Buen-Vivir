package config

import (
	"fmt"
	"os"
	"strings"
)

// AppEnv identifica el ambiente de despliegue de ESTE microservicio.
type AppEnv string

const (
	AppEnvTesting    AppEnv = "testing"
	AppEnvProduction AppEnv = "production"
)

// ContentfulEnv identifica el "Environment" de Contentful (dentro del
// mismo Space) contra el que habla este microservicio.
type ContentfulEnv string

const (
	ContentfulEnvTesting ContentfulEnv = "testing"
	ContentfulEnvMaster  ContentfulEnv = "master"
)

type Config struct {
	Port     string
	GRPCPort string

	AppEnv AppEnv

	ContentfulSpaceID     string
	ContentfulEnvironment ContentfulEnv
	ContentfulCDAToken    string // Content Delivery API (lectura pública)
	ContentfulCMAToken    string // Content Management API (escritura)
	ContentfulLocale      string

	CORSAllowedOrigins string

	// AuthServiceURL es la base de auth-service (ej. http://auth-service:8080).
	// Obligatoria en producción.
	AuthServiceURL        string
	AuthServiceGRPCTarget string
	// UpstreamIDToken activa la generación de un token de identidad de Google
	// en X-Serverless-Authorization para llamadas a servicios Cloud Run privados.
	UpstreamIDToken bool

	// EntregasBucket es el bucket privado de Cloud Storage para los adjuntos de las
	// entregas. Si está vacío, las entregas se aceptan solo con texto.
	EntregasBucket string

	// SkillsAI configura la sugerencia de habilidades con IA (HU-21 / HU-58). Es opcional.
	SkillsAI SkillsAIConfig

	// Events es la mensajería asíncrona con Pub/Sub (ver
	// docs/eventos/README.md). Desactivada si falta PUBSUB_PROJECT_ID.
	Events EventsConfig
}

// LoadConfig lee la configuración del entorno y del archivo .env (el entorno gana).
// Falla si APP_ENV es inválido o si la combinación con CONTENTFUL_ENVIRONMENT es peligrosa.
func LoadConfig() (*Config, error) {
	if err := loadEnvFile(".env"); err != nil {
		return nil, fmt.Errorf("no se pudo leer .env: %w", err)
	}

	appEnvRaw := os.Getenv("APP_ENV")
	if appEnvRaw == "" {
		appEnvRaw = string(AppEnvTesting)
	}
	appEnv := AppEnv(appEnvRaw)

	if appEnv != AppEnvTesting && appEnv != AppEnvProduction {
		return nil, fmt.Errorf("APP_ENV inválido: %q (valores permitidos: %q, %q)", appEnvRaw, AppEnvTesting, AppEnvProduction)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8210"
	}
	grpcPort := strings.TrimSpace(os.Getenv("GRPC_PORT"))
	if grpcPort == "" {
		grpcPort = "50053"
	}

	cfg := &Config{
		Port:                  port,
		GRPCPort:              grpcPort,
		AppEnv:                appEnv,
		ContentfulSpaceID:     os.Getenv("CONTENTFUL_SPACE_ID"),
		ContentfulEnvironment: ContentfulEnv(os.Getenv("CONTENTFUL_ENVIRONMENT")),
		ContentfulCDAToken:    os.Getenv("CONTENTFUL_CDA_TOKEN"),
		ContentfulCMAToken:    os.Getenv("CONTENTFUL_CMA_TOKEN"),
		ContentfulLocale:      os.Getenv("CONTENTFUL_LOCALE"),
		CORSAllowedOrigins:    os.Getenv("CORS_ALLOWED_ORIGINS"),
		AuthServiceURL:        strings.TrimRight(os.Getenv("AUTH_SERVICE_URL"), "/"),
		AuthServiceGRPCTarget: strings.TrimSpace(os.Getenv("AUTH_SERVICE_GRPC_TARGET")),
		UpstreamIDToken:       strings.EqualFold(strings.TrimSpace(os.Getenv("UPSTREAM_ID_TOKEN")), "true"),
		EntregasBucket:        strings.TrimSpace(os.Getenv("ENTREGAS_BUCKET")),
	}
	cfg.SkillsAI = loadSkillsAIConfig()
	cfg.Events = loadEventsConfig()

	if cfg.ContentfulLocale == "" {
		cfg.ContentfulLocale = "en-US"
	}

	if err := cfg.validateEnvironmentSafety(); err != nil {
		return nil, err
	}

	if cfg.ContentfulSpaceID == "" || cfg.ContentfulCDAToken == "" || cfg.ContentfulCMAToken == "" {
		return nil, fmt.Errorf("faltan variables de Contentful: CONTENTFUL_SPACE_ID, CONTENTFUL_CDA_TOKEN y CONTENTFUL_CMA_TOKEN son obligatorias")
	}

	if cfg.AppEnv == AppEnvProduction && cfg.AuthServiceURL == "" && cfg.AuthServiceGRPCTarget == "" {
		return nil, fmt.Errorf("falta la variable AUTH_SERVICE_URL: obligatoria en producción")
	}

	return cfg, nil
}

// validateEnvironmentSafety evita que testing escriba en el Environment "master" de
// Contentful o que producción corra contra "testing".
func (c *Config) validateEnvironmentSafety() error {
	if c.ContentfulEnvironment == "" {
		return fmt.Errorf("falta la variable CONTENTFUL_ENVIRONMENT")
	}

	switch {
	case c.AppEnv == AppEnvProduction && c.ContentfulEnvironment != ContentfulEnvMaster:
		return fmt.Errorf(
			"combinación de ambientes peligrosa: APP_ENV=%s requiere CONTENTFUL_ENVIRONMENT=%s, pero se encontró %q",
			AppEnvProduction, ContentfulEnvMaster, c.ContentfulEnvironment,
		)
	case c.AppEnv == AppEnvTesting && c.ContentfulEnvironment == ContentfulEnvMaster:
		return fmt.Errorf(
			"combinación de ambientes peligrosa: APP_ENV=%s no puede apuntar a CONTENTFUL_ENVIRONMENT=%s (produccion)",
			AppEnvTesting, ContentfulEnvMaster,
		)
	}

	return nil
}
