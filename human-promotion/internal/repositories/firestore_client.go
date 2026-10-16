package repositories

import (
	"context"
	"encoding/base64"
	"os"
	"strings"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
)

// NewFirestoreClient crea el cliente de Firestore compartido por los
// repositories de fichas e hitos (base propia de Promoción Humana).
//
// Acepta las credenciales igual que education-service:
//   - GOOGLE_APPLICATION_CREDENTIALS_BASE64: el JSON de la service account
//     en base64 (lo que inyecta el pipeline de CD a Cloud Run);
//   - GOOGLE_APPLICATION_CREDENTIALS con el JSON crudo (empieza con '{');
//   - GOOGLE_APPLICATION_CREDENTIALS como path, o sin definir: Application
//     Default Credentials estándar (también cubre FIRESTORE_EMULATOR_HOST).
//
// Con credenciales inline se limpia la variable de entorno para que
// ninguna capa interna de las librerías de Google intente abrirla como path.
func NewFirestoreClient(ctx context.Context) (*firestore.Client, error) {
	credentialsJSON, err := googleCredentialsJSON()
	if err != nil {
		return nil, err
	}
	if credentialsJSON == nil {
		return firestore.NewClient(ctx, firestore.DetectProjectID)
	}
	_ = os.Unsetenv("GOOGLE_APPLICATION_CREDENTIALS")
	return firestore.NewClient(ctx, firestore.DetectProjectID, option.WithCredentialsJSON(credentialsJSON)) //nolint:staticcheck // option.WithCredentialsJSON es necesario para credenciales inline (mismo criterio que education)
}

// GoogleCredentialsJSON expone las credenciales inline para otros
// clientes de Google (Pub/Sub). Hay que llamarla ANTES de
// NewFirestoreClient, que limpia GOOGLE_APPLICATION_CREDENTIALS.
func GoogleCredentialsJSON() ([]byte, error) { return googleCredentialsJSON() }

func googleCredentialsJSON() ([]byte, error) {
	if b64 := strings.TrimSpace(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS_BASE64")); b64 != "" {
		return base64.StdEncoding.DecodeString(b64)
	}
	credentialsEnv := strings.TrimSpace(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	if strings.HasPrefix(credentialsEnv, "{") {
		return []byte(credentialsEnv), nil
	}
	return nil, nil
}
