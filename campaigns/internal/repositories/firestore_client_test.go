package repositories

import (
	"encoding/base64"
	"testing"
)

func TestGoogleCredentialsJSON(t *testing.T) {
	const json = `{"type":"service_account"}`

	t.Run("base64 tiene prioridad", func(t *testing.T) {
		t.Setenv("GOOGLE_APPLICATION_CREDENTIALS_BASE64", base64.StdEncoding.EncodeToString([]byte(json)))
		t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/ruta/a/otra.json")
		got, err := googleCredentialsJSON()
		if err != nil || string(got) != json {
			t.Fatalf("got %q, %v", got, err)
		}
	})
	t.Run("JSON crudo en la variable", func(t *testing.T) {
		t.Setenv("GOOGLE_APPLICATION_CREDENTIALS_BASE64", "")
		t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "  "+json)
		got, err := googleCredentialsJSON()
		if err != nil || string(got) != json {
			t.Fatalf("got %q, %v", got, err)
		}
	})
	t.Run("ruta a un archivo: credenciales estándar", func(t *testing.T) {
		t.Setenv("GOOGLE_APPLICATION_CREDENTIALS_BASE64", "")
		t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/ruta/sa.json")
		got, err := googleCredentialsJSON()
		if err != nil || got != nil {
			t.Fatalf("got %q, %v", got, err)
		}
	})
	t.Run("base64 inválido", func(t *testing.T) {
		t.Setenv("GOOGLE_APPLICATION_CREDENTIALS_BASE64", "esto-no-es-base64!!")
		if _, err := googleCredentialsJSON(); err == nil {
			t.Fatal("esperaba error")
		}
	})
}
