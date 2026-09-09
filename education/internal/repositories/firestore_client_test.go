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
