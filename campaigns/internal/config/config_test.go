package config

import "testing"

func TestLoadConfig(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:4200")
	cfg := LoadConfig()
	if cfg.Port != DefaultPort || cfg.CORSAllowedOrigins != "http://localhost:4200" {
		t.Fatalf("config por defecto: %+v", cfg)
	}

	t.Setenv("PORT", " 9090 ")
	if cfg := LoadConfig(); cfg.Port != "9090" {
		t.Fatalf("PORT definido: %+v", cfg)
	}
}
