package repositories

import (
	"context"
	"encoding/base64"
	"os"
	"strings"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
)

// NewFirestoreClient crea el cliente de Firestore de Campañas. Acepta las credenciales en
// GOOGLE_APPLICATION_CREDENTIALS_BASE64, como JSON crudo o como path (o ADC). Con
// credenciales inline se limpia la variable para que las librerías de Google no la abran
// como path.
func NewFirestoreClient(ctx context.Context) (*firestore.Client, error) {
	credentialsJSON, err := googleCredentialsJSON()
	if err != nil {
		return nil, err
	}
	var opts []option.ClientOption
	if credentialsJSON != nil {
		_ = os.Unsetenv("GOOGLE_APPLICATION_CREDENTIALS")
		opts = append(opts, option.WithCredentialsJSON(credentialsJSON)) //nolint:staticcheck // option.WithCredentialsJSON es necesario para credenciales inline (mismo criterio que human-promotion)
	}
	if dbID := resolveFirestoreDatabase(); dbID != "" {
		return firestore.NewClientWithDatabase(ctx, firestore.DetectProjectID, dbID, opts...)
	}
	return firestore.NewClient(ctx, firestore.DetectProjectID, opts...)
}

// resolveFirestoreDatabase returns the configured Firestore database ID, or empty string for "(default)".
func resolveFirestoreDatabase() string {
	dbID := strings.TrimSpace(os.Getenv("FIRESTORE_DATABASE"))
	if dbID == "" || dbID == "(default)" || strings.EqualFold(dbID, "null") {
		return ""
	}
	return dbID
}

// GoogleCredentialsJSON exposes inline credentials for other Google Cloud clients (Pub/Sub).
// Must be called before NewFirestoreClient, which unsets GOOGLE_APPLICATION_CREDENTIALS.
func GoogleCredentialsJSON() ([]byte, error) {
	return googleCredentialsJSON()
}

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
