package config

import (
	"os"
	"strings"
)

// Config contains configuration parameters for rxresume microservice.
type Config struct {
	Port             string
	DatabaseURL      string
	RendererURL      string
	GCPProjectID     string
	GCPBucketName    string
	ApplicationCreds string
	// JWTSecret: con valor, la identidad sale del JWT verificado y no de
	// los headers X-User-* (ver routes/security.go).
	JWTSecret string
	// CORSAllowedOrigins: orígenes extra, separados por comas.
	CORSAllowedOrigins string
	AppEnv             string
}

// LoadConfig reads configuration settings from environment variables with safe defaults.
func LoadConfig() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// DATABASE_URL_RXRESUME tiene prioridad; si no está se usa
	// DATABASE_URL, que es la que pasan el CD de Cloud Run y el .env local.
	// Antes solo se leía la primera: en testing y en Docker el servicio
	// caía al default localhost:5432 y "Mi CV" respondía 500 (auditoría
	// 17/11/2026).
	dbURL := cleanDSN(os.Getenv("DATABASE_URL_RXRESUME"))
	if dbURL == "" {
		dbURL = cleanDSN(os.Getenv("DATABASE_URL"))
	}
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/buenvivir_rxresume?sslmode=disable"
	}

	rendererURL := os.Getenv("RENDERER_URL")
	if rendererURL == "" {
		rendererURL = "http://cv-renderer:5005"
	}

	gcpProjectID := os.Getenv("GCP_PROJECT_ID")
	if gcpProjectID == "" {
		gcpProjectID = "ceibuenvivir"
	}

	gcpBucketName := os.Getenv("GCP_BUCKET_NAME")
	if gcpBucketName == "" {
		gcpBucketName = "buenvivir-resumes-bucket"
	}

	appCreds := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")

	return &Config{
		Port:               port,
		DatabaseURL:        dbURL,
		RendererURL:        rendererURL,
		GCPProjectID:       gcpProjectID,
		GCPBucketName:      gcpBucketName,
		ApplicationCreds:   appCreds,
		JWTSecret:          os.Getenv("JWT_SECRET"),
		CORSAllowedOrigins: os.Getenv("CORS_ALLOWED_ORIGINS"),
		AppEnv:             os.Getenv("APP_ENV"),
	}
}

// cleanDSN saca espacios y comillas que envuelvan la cadena de conexión.
// En testing el secret se cargó como 'postgresql://...' (copiado del
// comando psql de Neon) y el driver no reconocía la URL: intentaba
// conectarse por socket local con el usuario del contenedor (17/11/2026).
func cleanDSN(v string) string {
	v = strings.TrimSpace(v)
	for len(v) >= 2 && (v[0] == '\'' && v[len(v)-1] == '\'' || v[0] == '"' && v[len(v)-1] == '"') {
		v = strings.TrimSpace(v[1 : len(v)-1])
	}
	return v
}
