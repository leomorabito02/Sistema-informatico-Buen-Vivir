package config

import "os"

// Config contains configuration parameters for rxresume microservice.
type Config struct {
	Port                    string
	DatabaseURL             string
	ReactiveResumeURL       string
	ReactiveResumeSecretKey string
	ReactiveResumePublicURL string
	GCPProjectID            string
	GCPBucketName           string
	GCPEmulatorHost         string
	ApplicationCreds        string
	AuthServiceURL          string
	FrontendURL             string
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

	rxURL := os.Getenv("REACTIVE_RESUME_URL")
	if rxURL == "" {
		rxURL = "http://reactive-resume:3000"
	}

	rxSecret := os.Getenv("REACTIVE_RESUME_SECRET_KEY")
	if rxSecret == "" {
		rxSecret = "buenvivir_secret_key_123"
	}

	rxPublicURL := os.Getenv("REACTIVE_RESUME_PUBLIC_URL")
	if rxPublicURL == "" {
		rxPublicURL = "http://localhost:3000"
	}

	gcpProjectID := os.Getenv("GCP_PROJECT_ID")
	if gcpProjectID == "" {
		gcpProjectID = "buenvivir-gcp-project"
	}

	gcpBucketName := os.Getenv("GCP_BUCKET_NAME")
	if gcpBucketName == "" {
		gcpBucketName = "buenvivir-resumes-bucket"
	}

	gcpEmulatorHost := os.Getenv("GCP_EMULATOR_HOST")
	appCreds := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")

	authServiceURL := os.Getenv("AUTH_SERVICE_URL")
	if authServiceURL == "" {
		authServiceURL = "http://auth-service:8080"
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:4200"
	}

	return &Config{
		Port:                    port,
		DatabaseURL:             dbURL,
		ReactiveResumeURL:       rxURL,
		ReactiveResumeSecretKey: rxSecret,
		ReactiveResumePublicURL: rxPublicURL,
		GCPProjectID:            gcpProjectID,
		GCPBucketName:           gcpBucketName,
		GCPEmulatorHost:         gcpEmulatorHost,
		ApplicationCreds:        appCreds,
		AuthServiceURL:          authServiceURL,
		FrontendURL:             frontendURL,
	}
}
