package config

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
)

// NewFirestoreClient initializes and returns a new Firestore client.
func NewFirestoreClient(ctx context.Context) (*firestore.Client, error) {
	projectID := strings.TrimSpace(os.Getenv("GOOGLE_CLOUD_PROJECT"))
	if projectID == "" {
		projectID = firestore.DetectProjectID
	}

	dbID := strings.TrimSpace(os.Getenv("FIRESTORE_DATABASE"))

	credentialsJSON, err := googleCredentialsJSON()
	if err != nil {
		return nil, fmt.Errorf("invalid Google credentials: %w", err)
	}

	var opts []option.ClientOption
	if credentialsJSON != nil {
		_ = os.Unsetenv("GOOGLE_APPLICATION_CREDENTIALS")
		opts = append(opts, option.WithCredentialsJSON(credentialsJSON)) //nolint:staticcheck // inline credentials injected in Cloud Run CD
	}

	var client *firestore.Client
	if dbID != "" && dbID != "(default)" {
		client, err = firestore.NewClientWithDatabase(ctx, projectID, dbID, opts...)
	} else {
		client, err = firestore.NewClient(ctx, projectID, opts...)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create Firestore client: %w", err)
	}

	return client, nil
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
