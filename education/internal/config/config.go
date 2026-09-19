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

// Config agrupa toda la configuración del servicio de Educación.
type Config struct {
	Port string

	AppEnv AppEnv

	ContentfulSpaceID     string
	ContentfulEnvironment ContentfulEnv
	ContentfulCDAToken    string // Content Delivery API (lectura pública)
	ContentfulCMAToken    string // Content Management API (escritura)
	ContentfulLocale      string

	CORSAllowedOrigins string

	// AuthServiceURL es la base (sin path) de auth-service, ej.
	// "http://auth-service:8080" en Docker o
	// "https://auth-testing-xxxx.run.app" en Cloud Run. La usa
	// AuthServiceClient para validar participantes al inscribir/registrar
	// asistencia y para resolver el listado de participantes en la
	// asistencia general, reenviando el JWT del usuario que llamó a
	// education. Si queda vacía, esas funcionalidades fallan en tiempo de
	// request (no se aborta el arranque del servicio: el resto de
	// education no depende de auth-service).
	AuthServiceURL string
}

// LoadConfig arma la configuración del servicio. El archivo que se
// carga (.env.testing o .env.production) depende de APP_ENV, y las
// variables ya presentes en el entorno real del proceso siempre ganan
// sobre lo que diga el archivo (ver loadEnvFile). Devuelve error si la
// combinación de APP_ENV / CONTENTFUL_ENVIRONMENT es peligrosa, para
// que el servicio directamente no arranque.
func LoadConfig() (*Config, error) {
	appEnvRaw := os.Getenv("APP_ENV")
	if appEnvRaw == "" {
		appEnvRaw = string(AppEnvTesting)
	}
	appEnv := AppEnv(appEnvRaw)

	var envFile string
	switch appEnv {
	case AppEnvTesting:
		envFile = ".env.testing"
	case AppEnvProduction:
		envFile = ".env.production"
	default:
		return nil, fmt.Errorf("APP_ENV inválido: %q (valores permitidos: %q, %q)", appEnvRaw, AppEnvTesting, AppEnvProduction)
	}

	if err := loadEnvFile(envFile); err != nil {
		return nil, fmt.Errorf("no se pudo leer %s: %w", envFile, err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8210"
	}

	cfg := &Config{
		Port:                  port,
		AppEnv:                appEnv,
		ContentfulSpaceID:     os.Getenv("CONTENTFUL_SPACE_ID"),
		ContentfulEnvironment: ContentfulEnv(os.Getenv("CONTENTFUL_ENVIRONMENT")),
		ContentfulCDAToken:    os.Getenv("CONTENTFUL_CDA_TOKEN"),
		ContentfulCMAToken:    os.Getenv("CONTENTFUL_CMA_TOKEN"),
		ContentfulLocale:      os.Getenv("CONTENTFUL_LOCALE"),
		CORSAllowedOrigins:    os.Getenv("CORS_ALLOWED_ORIGINS"),
		AuthServiceURL:        strings.TrimRight(os.Getenv("AUTH_SERVICE_URL"), "/"),
	}

	if cfg.ContentfulLocale == "" {
		cfg.ContentfulLocale = "en-US"
	}

	if err := cfg.validateEnvironmentSafety(); err != nil {
		return nil, err
	}

	if cfg.ContentfulSpaceID == "" || cfg.ContentfulCDAToken == "" || cfg.ContentfulCMAToken == "" {
		return nil, fmt.Errorf("faltan variables de Contentful: CONTENTFUL_SPACE_ID, CONTENTFUL_CDA_TOKEN y CONTENTFUL_CMA_TOKEN son obligatorias")
	}

	return cfg, nil
}

// validateEnvironmentSafety rechaza combinaciones peligrosas entre el
// ambiente de la app y el Environment de Contentful, para que no sea
// posible que una instancia de testing escriba por error en el
// Environment "master" (producción) de Contentful, ni que producción
// quede corriendo contra "testing".
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
