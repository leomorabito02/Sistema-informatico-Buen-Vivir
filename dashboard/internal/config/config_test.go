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

func TestLoadConfig_CacheYPush(t *testing.T) {
	t.Setenv("DASHBOARD_CACHE_TTL", "")
	t.Setenv("PUBSUB_PUSH_AUTH", "")
	t.Setenv("PUBSUB_PUSH_AUDIENCE", "")
	t.Setenv("PUBSUB_PUSH_SERVICE_ACCOUNTS", "")
	cfg := config.LoadConfig()
	if cfg.CacheTTL.Minutes() != 5 || cfg.ConsumeEnabled() {
		t.Fatalf("por defecto: %+v", cfg)
	}

	t.Setenv("APP_ENV", "production")
	t.Setenv("DASHBOARD_CACHE_TTL", "30s")
	t.Setenv("PUBSUB_PUSH_AUDIENCE", "https://dashboard.run.app/eventos")
	t.Setenv("PUBSUB_PUSH_SERVICE_ACCOUNTS", "push-dashboard@p.iam.gserviceaccount.com, ")
	cfg = config.LoadConfig()
	if cfg.CacheTTL.Seconds() != 30 || !cfg.Production || !cfg.ConsumeEnabled() || len(cfg.PushServiceAccounts) != 1 {
		t.Fatalf("con variables: %+v", cfg)
	}

	t.Setenv("PUBSUB_PUSH_AUDIENCE", "")
	t.Setenv("PUBSUB_PUSH_AUTH", "disabled")
	if !config.LoadConfig().ConsumeEnabled() {
		t.Fatal("local con el emulador")
	}
}
