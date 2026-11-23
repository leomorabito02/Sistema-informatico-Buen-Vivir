package config

import "testing"

func TestLoadConfig_FrontendURL(t *testing.T) {
	t.Run("UsesExplicitPublicFrontendURL", func(t *testing.T) {
		// Arrange
		t.Setenv("ENV", "testing")
		t.Setenv("FRONTEND_URL", "https://custom.ceibuenvivir.org/")

		// Act
		cfg := LoadConfig()

		// Assert
		if cfg.FrontendURL != "https://custom.ceibuenvivir.org" {
			t.Errorf("expected explicit public FrontendURL, got %q", cfg.FrontendURL)
		}
	})

	t.Run("IgnoresLocalhostAndDefaultsToTesting", func(t *testing.T) {
		// Arrange
		t.Setenv("ENV", "testing")
		t.Setenv("FRONTEND_URL", "http://localhost:4200")

		// Act
		cfg := LoadConfig()

		// Assert
		if cfg.FrontendURL != "https://testing.ceibuenvivir.pages.dev" {
			t.Errorf("expected testing URL instead of localhost, got %q", cfg.FrontendURL)
		}
	})

	t.Run("DefaultsToProductionWhenProductionEnv", func(t *testing.T) {
		// Arrange
		t.Setenv("ENV", "production")
		t.Setenv("FRONTEND_URL", "")

		// Act
		cfg := LoadConfig()

		// Assert
		if cfg.FrontendURL != "https://ceibuenvivir.org" {
			t.Errorf("expected production URL, got %q", cfg.FrontendURL)
		}
	})
}
