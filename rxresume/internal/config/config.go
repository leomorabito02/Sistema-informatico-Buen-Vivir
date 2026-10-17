package config

import "os"

// Config contains configuration parameters for rxresume microservice.
type Config struct {
	Port             string
	DatabaseURL      string
	RendererURL      string
	GCPProjectID     string
	GCPBucketName    string
	ApplicationCreds string
}

// LoadConfig reads configuration settings from environment variables with safe defaults.
func LoadConfig() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
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
		Port:             port,
		DatabaseURL:      dbURL,
		RendererURL:      rendererURL,
		GCPProjectID:     gcpProjectID,
		GCPBucketName:    gcpBucketName,
		ApplicationCreds: appCreds,
	}
}
