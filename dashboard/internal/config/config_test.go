package config_test

import (
	"testing"

	"buenvivir/dashboard/internal/config"
)

func TestLoadConfig_DefaultsYVariables(t *testing.T) {
	cfg := config.LoadConfig()
	if cfg.Port != "8090" || cfg.EducationURL != "http://education-service:8210" || cfg.Timeout <= 0 {
		t.Fatalf("valores por defecto: %+v", cfg)
	}

	t.Setenv("PORT", "9999")
	t.Setenv("LABOR_SERVICE_URL", "http://labor:1")
	cfg = config.LoadConfig()
	if cfg.Port != "9999" || cfg.LaborURL != "http://labor:1" || cfg.AuthURL != "http://auth-service:8080" {
		t.Fatalf("variables de entorno: %+v", cfg)
	}
}
