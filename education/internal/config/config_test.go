package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfig_ValidateEnvironmentSafety_DecisionTable tests safety rules using a Decision Table:
// Conditions: AppEnv (Production, Testing) x ContentfulEnvironment (Master, Testing, Empty, Other)
func TestConfig_ValidateEnvironmentSafety_DecisionTable(t *testing.T) {
	decisionTable := []struct {
		name          string
		appEnv        AppEnv
		contentfulEnv ContentfulEnv
		expectError   bool
		errorSubstr   string
	}{
		// Rule 1: Production + Master => Valid
		{
			name:          "Rule 1: Prod with Master is valid",
			appEnv:        AppEnvProduction,
			contentfulEnv: ContentfulEnvMaster,
			expectError:   false,
		},
		// Rule 2: Production + Testing => Invalid (dangerous)
		{
			name:          "Rule 2: Prod with Testing is rejected",
			appEnv:        AppEnvProduction,
			contentfulEnv: ContentfulEnvTesting,
			expectError:   true,
			errorSubstr:   "combinación de ambientes peligrosa",
		},
		// Rule 3: Production + Other => Invalid (dangerous)
		{
			name:          "Rule 3: Prod with non-master is rejected",
			appEnv:        AppEnvProduction,
			contentfulEnv: "staging",
			expectError:   true,
			errorSubstr:   "requiere CONTENTFUL_ENVIRONMENT=master",
		},
		// Rule 4: Testing + Testing => Valid
		{
			name:          "Rule 4: Testing with Testing is valid",
			appEnv:        AppEnvTesting,
			contentfulEnv: ContentfulEnvTesting,
			expectError:   false,
		},
		// Rule 5: Testing + Master => Invalid (dangerous: testing writing to prod)
		{
			name:          "Rule 5: Testing with Master is rejected",
			appEnv:        AppEnvTesting,
			contentfulEnv: ContentfulEnvMaster,
			expectError:   true,
			errorSubstr:   "no puede apuntar a CONTENTFUL_ENVIRONMENT=master",
		},
		// Rule 6: Missing ContentfulEnvironment (boundary: empty string)
		{
			name:          "Rule 6: Empty ContentfulEnvironment is rejected",
			appEnv:        AppEnvTesting,
			contentfulEnv: "",
			expectError:   true,
			errorSubstr:   "falta la variable CONTENTFUL_ENVIRONMENT",
		},
	}

	for _, tt := range decisionTable {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			cfg := &Config{
				AppEnv:                tt.appEnv,
				ContentfulEnvironment: tt.contentfulEnv,
			}

			// Act
			err := cfg.validateEnvironmentSafety()

			// Assert
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorSubstr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestLoadConfig_EquivalenceClassesAndBoundaries validates LoadConfig equivalence partitions and boundary values.
func TestLoadConfig_EquivalenceClassesAndBoundaries(t *testing.T) {
	t.Run("Invalid APP_ENV equivalence class", func(t *testing.T) {
		// Arrange
		t.Setenv("APP_ENV", "invalid-env")

		// Act
		cfg, err := LoadConfig()

		// Assert
		assert.Error(t, err)
		assert.Nil(t, cfg)
		assert.Contains(t, err.Error(), "APP_ENV inválido")
	})

	t.Run("Missing Contentful required variables equivalence class", func(t *testing.T) {
		// Arrange
		t.Setenv("APP_ENV", "testing")
		t.Setenv("CONTENTFUL_ENVIRONMENT", "testing")
		t.Setenv("CONTENTFUL_SPACE_ID", "")
		t.Setenv("CONTENTFUL_CDA_TOKEN", "token-cda")
		t.Setenv("CONTENTFUL_CMA_TOKEN", "token-cma")

		// Act
		cfg, err := LoadConfig()

		// Assert
		assert.Error(t, err)
		assert.Nil(t, cfg)
		assert.Contains(t, err.Error(), "faltan variables de Contentful")
	})

	t.Run("Boundary values: Default port and default locale fallback", func(t *testing.T) {
		// Arrange
		t.Setenv("APP_ENV", "testing")
		t.Setenv("PORT", "")
		t.Setenv("CONTENTFUL_ENVIRONMENT", "testing")
		t.Setenv("CONTENTFUL_SPACE_ID", "space-id")
		t.Setenv("CONTENTFUL_CDA_TOKEN", "cda-token")
		t.Setenv("CONTENTFUL_CMA_TOKEN", "cma-token")
		t.Setenv("CONTENTFUL_LOCALE", "")
		t.Setenv("AUTH_SERVICE_URL", "http://auth:8080/")

		// Act
		cfg, err := LoadConfig()

		// Assert
		require.NoError(t, err)
		require.NotNil(t, cfg)
		assert.Equal(t, "8210", cfg.Port, "Port falls back to default 8210")
		assert.Equal(t, "en-US", cfg.ContentfulLocale, "Locale falls back to default en-US")
		assert.Equal(t, "http://auth:8080", cfg.AuthServiceURL, "Trailing slash is trimmed")
	})

	t.Run("Production APP_ENV loads .env.production", func(t *testing.T) {
		// Arrange
		t.Setenv("APP_ENV", "production")
		t.Setenv("CONTENTFUL_ENVIRONMENT", "master")
		t.Setenv("CONTENTFUL_SPACE_ID", "space-id")
		t.Setenv("CONTENTFUL_CDA_TOKEN", "cda-token")
		t.Setenv("CONTENTFUL_CMA_TOKEN", "cma-token")

		// Act
		cfg, err := LoadConfig()

		// Assert
		require.NoError(t, err)
		assert.Equal(t, AppEnvProduction, cfg.AppEnv)
	})

	t.Run("Empty APP_ENV defaults to testing, custom PORT and LOCALE preserved", func(t *testing.T) {
		// Arrange
		t.Setenv("APP_ENV", "")
		t.Setenv("PORT", "9090")
		t.Setenv("CONTENTFUL_LOCALE", "es-AR")
		t.Setenv("CONTENTFUL_ENVIRONMENT", "testing")
		t.Setenv("CONTENTFUL_SPACE_ID", "space-id")
		t.Setenv("CONTENTFUL_CDA_TOKEN", "cda-token")
		t.Setenv("CONTENTFUL_CMA_TOKEN", "cma-token")

		// Act
		cfg, err := LoadConfig()

		// Assert
		require.NoError(t, err)
		assert.Equal(t, AppEnvTesting, cfg.AppEnv)
		assert.Equal(t, "9090", cfg.Port)
		assert.Equal(t, "es-AR", cfg.ContentfulLocale)
	})

	t.Run("Dangerous environment configuration rejected in LoadConfig", func(t *testing.T) {
		// Arrange
		t.Setenv("APP_ENV", "testing")
		t.Setenv("CONTENTFUL_ENVIRONMENT", "master")

		// Act
		cfg, err := LoadConfig()

		// Assert
		assert.Error(t, err)
		assert.Nil(t, cfg)
		assert.Contains(t, err.Error(), "combinación de ambientes peligrosa")
	})
}

// TestLoadEnvFile_EquivalenceClasses tests .env parsing with comments, quotes and boundaries.
func TestLoadEnvFile_EquivalenceClasses(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env.test")
	content := `
# Comment line
EMPTY_LINE=

KEY_PLAIN=value_plain
KEY_DOUBLE="value with spaces"
KEY_SINGLE='single quoted'
INVALID_LINE_WITHOUT_EQUALS
=KEY_EMPTY
`
	err := os.WriteFile(envPath, []byte(content), 0600)
	require.NoError(t, err)

	// Act
	err = loadEnvFile(envPath)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "value_plain", os.Getenv("KEY_PLAIN"))
	assert.Equal(t, "value with spaces", os.Getenv("KEY_DOUBLE"))
	assert.Equal(t, "single quoted", os.Getenv("KEY_SINGLE"))

	// Non-existent file boundary
	// Act
	errNonExistent := loadEnvFile(filepath.Join(tmpDir, "non_existent.env"))
	// Assert
	assert.NoError(t, errNonExistent, "non-existent file should be treated as no-op")

	// Pre-existing environment variable preserved
	t.Run("Pre-existing environment variable is not overwritten", func(t *testing.T) {
		// Arrange
		subDir := t.TempDir()
		path := filepath.Join(subDir, ".env.existing")
		require.NoError(t, os.WriteFile(path, []byte("PRE_EXISTING_KEY=new_file_val\n"), 0600))
		t.Setenv("PRE_EXISTING_KEY", "original_env_val")

		// Act
		loadErr := loadEnvFile(path)

		// Assert
		require.NoError(t, loadErr)
		assert.Equal(t, "original_env_val", os.Getenv("PRE_EXISTING_KEY"))
	})
}
