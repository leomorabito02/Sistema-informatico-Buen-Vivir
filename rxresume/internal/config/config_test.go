package config

import (
	"encoding/base64"
	"testing"
)

// El CD de Cloud Run y el .env local pasan DATABASE_URL: si no hay
// DATABASE_URL_RXRESUME se usa esa.
func TestLoadConfig_BaseDeDatos(t *testing.T) {
	t.Setenv("DATABASE_URL_RXRESUME", "")
	t.Setenv("DATABASE_URL", "postgres://neon/rx")
	if got := LoadConfig().DatabaseURL; got != "postgres://neon/rx" {
		t.Fatalf("DATABASE_URL no se usó: %q", got)
	}

	t.Setenv("DATABASE_URL_RXRESUME", "postgres://propia/rx")
	if got := LoadConfig().DatabaseURL; got != "postgres://propia/rx" {
		t.Fatalf("DATABASE_URL_RXRESUME debería tener prioridad: %q", got)
	}

	t.Setenv("DATABASE_URL_RXRESUME", "")
	t.Setenv("DATABASE_URL", "")
	if got := LoadConfig().DatabaseURL; got == "" {
		t.Fatal("sin variables debería quedar el default local")
	}
}

// Comillas o espacios alrededor del secret no deben romper la conexión.
func TestLoadConfig_BaseDeDatosConComillas(t *testing.T) {
	t.Setenv("DATABASE_URL_RXRESUME", "")
	for _, v := range []string{"'postgresql://u:p@h/db'", "\"postgresql://u:p@h/db\"", "  postgresql://u:p@h/db \n"} {
		t.Setenv("DATABASE_URL", v)
		if got := LoadConfig().DatabaseURL; got != "postgresql://u:p@h/db" {
			t.Fatalf("%q quedó como %q", v, got)
		}
	}
}

func TestLoadConfig_GoogleCredentialsBase64(t *testing.T) {
	// Arrange
	rawJSON := `{"type":"service_account","project_id":"ceibuenvivir"}`
	encoded := base64.StdEncoding.EncodeToString([]byte("  " + rawJSON + "\n"))
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS_BASE64", encoded)
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/tmp/fallback.json")

	// Act
	cfg := LoadConfig()

	// Assert
	if cfg.ApplicationCreds != rawJSON {
		t.Fatalf("expected decoded JSON %q, got %q", rawJSON, cfg.ApplicationCreds)
	}
}

func TestLoadConfig_GoogleCredentialsFallback(t *testing.T) {
	// Arrange
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS_BASE64", "invalid-base64!!!")
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", " /tmp/creds.json ")

	// Act
	cfg := LoadConfig()

	// Assert
	if cfg.ApplicationCreds != "/tmp/creds.json" {
		t.Fatalf("expected fallback path %q, got %q", "/tmp/creds.json", cfg.ApplicationCreds)
	}
}
