package repositories

import (
	"context"
	"encoding/base64"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFirestoreClient_RawJSONCredentials_UnsetsEnvVar(t *testing.T) {
	if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" {
		t.Skip("FIRESTORE_EMULATOR_HOST no seteado")
	}

	// GOOGLE_APPLICATION_CREDENTIALS con el JSON crudo (no un path):
	// NewFirestoreClient tiene que detectarlo por el prefijo '{', usarlo
	// con option.WithCredentialsJSON y limpiar la variable para que el
	// transporte gRPC no intente abrirla como archivo. Contra el emulador
	// las credenciales se ignoran.
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", fakeCredentialsJSON)

	client, err := NewFirestoreClient(context.Background())
	require.NoError(t, err)
	defer func() { _ = client.Close() }()

	assert.Empty(t, os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"),
		"la variable con el JSON crudo se limpia tras construir el cliente")
}

func TestNewFirestoreClient_Base64Credentials_TakesPriority(t *testing.T) {
	if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" {
		t.Skip("FIRESTORE_EMULATOR_HOST no seteado")
	}

	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS_BASE64", base64.StdEncoding.EncodeToString([]byte(fakeCredentialsJSON)))
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/no/deberia/usarse.json")

	client, err := NewFirestoreClient(context.Background())
	require.NoError(t, err)
	defer func() { _ = client.Close() }()

	assert.Empty(t, os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"),
		"cuando hay _BASE64, se limpia el path de GOOGLE_APPLICATION_CREDENTIALS")
}

func TestNewFirestoreClient_Base64Credentials_InvalidReturnsError(t *testing.T) {
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS_BASE64", "no-es-base64-válido!!!")

	client, err := NewFirestoreClient(context.Background())
	require.Error(t, err)
	assert.Nil(t, client)
}

func TestNewFirestoreClient_PathCredentials_LeavesEnvVarUntouched(t *testing.T) {
	if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" {
		t.Skip("FIRESTORE_EMULATOR_HOST no seteado")
	}

	// Valor que NO empieza con '{': se trata como path a un archivo de
	// credenciales (ADC estándar) y la variable no se toca. Con el
	// emulador levantado el cliente se crea igual sin resolver el path.
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/ruta/inexistente/sa.json")

	client, err := NewFirestoreClient(context.Background())
	require.NoError(t, err)
	defer func() { _ = client.Close() }()

	assert.Equal(t, "/ruta/inexistente/sa.json", os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
}

// TestGoogleCredentialsJSON_EquivalenceClasses tests credentials resolution using Equivalence Classes without requiring Firestore emulator.
func TestGoogleCredentialsJSON_EquivalenceClasses(t *testing.T) {
	t.Run("Class 1: Valid Base64 credentials", func(t *testing.T) {
		// Arrange
		expectedJSON := `{"type":"service_account","project_id":"p1"}`
		b64 := base64.StdEncoding.EncodeToString([]byte(expectedJSON))
		t.Setenv("GOOGLE_APPLICATION_CREDENTIALS_BASE64", b64)

		// Act
		raw, err := GoogleCredentialsJSON()

		// Assert
		require.NoError(t, err)
		assert.Equal(t, []byte(expectedJSON), raw)
	})

	t.Run("Class 2: Invalid Base64 credentials returns decoding error", func(t *testing.T) {
		// Arrange
		t.Setenv("GOOGLE_APPLICATION_CREDENTIALS_BASE64", "not-valid-base64???")

		// Act
		raw, err := GoogleCredentialsJSON()

		// Assert
		assert.Error(t, err)
		assert.Empty(t, raw)
	})

	t.Run("Class 3: Raw JSON in GOOGLE_APPLICATION_CREDENTIALS (starts with '{')", func(t *testing.T) {
		// Arrange
		t.Setenv("GOOGLE_APPLICATION_CREDENTIALS_BASE64", "")
		expectedJSON := `{"project_id":"inline-sa"}`
		t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", expectedJSON)

		// Act
		raw, err := GoogleCredentialsJSON()

		// Assert
		require.NoError(t, err)
		assert.Equal(t, []byte(expectedJSON), raw)
	})

	t.Run("Class 4: File path in GOOGLE_APPLICATION_CREDENTIALS returns nil", func(t *testing.T) {
		// Arrange
		t.Setenv("GOOGLE_APPLICATION_CREDENTIALS_BASE64", "")
		t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/etc/secrets/sa.json")

		// Act
		raw, err := GoogleCredentialsJSON()

		// Assert
		require.NoError(t, err)
		assert.Nil(t, raw)
	})

	t.Run("Class 5: Empty/unset env variables returns nil", func(t *testing.T) {
		// Arrange
		t.Setenv("GOOGLE_APPLICATION_CREDENTIALS_BASE64", "   ")
		t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "   ")

		// Act
		raw, err := GoogleCredentialsJSON()

		// Assert
		require.NoError(t, err)
		assert.Nil(t, raw)
	})
}

func TestResolveFirestoreDatabase(t *testing.T) {
	for _, tc := range []struct {
		name string
		val  string
		want string
	}{
		{"empty", "", ""},
		{"default", "(default)", ""},
		{"null", "null", ""},
		{"custom", "buenvivir-prod", "buenvivir-prod"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("FIRESTORE_DATABASE", tc.val)
			assert.Equal(t, tc.want, resolveFirestoreDatabase())
		})
	}
}

