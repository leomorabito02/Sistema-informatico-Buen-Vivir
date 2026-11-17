package config

import "testing"

// El CD de Cloud Run y el .env local pasan DATABASE_URL; antes solo se
// leía DATABASE_URL_RXRESUME y el servicio caía a localhost:5432.
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
