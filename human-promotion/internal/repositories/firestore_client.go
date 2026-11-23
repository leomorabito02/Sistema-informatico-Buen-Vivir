package repositories

import (
	"context"
	"encoding/base64"
	"os"
	"strings"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
)

// NewFirestoreClient crea el cliente de Firestore de fichas e hitos. Acepta las
// credenciales en GOOGLE_APPLICATION_CREDENTIALS_BASE64, como JSON crudo o como path
// (o ADC). Con credenciales inline se limpia la variable para que las librerías de
// Google no la abran como path.
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
