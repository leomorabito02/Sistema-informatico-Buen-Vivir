package config

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
)

// NewFirestoreClient initializes and returns a new Firestore client.
func NewFirestoreClient(ctx context.Context) (*firestore.Client, error) {
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		return nil, fmt.Errorf("GOOGLE_CLOUD_PROJECT environment variable is not set")
	}

	dbID := os.Getenv("FIRESTORE_DATABASE")

	var opts []option.ClientOption
	if creds := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); creds != "" {
		// If it looks like JSON, use it directly instead of treating it as a file path
		if len(creds) > 0 && creds[0] == '{' {
			opts = append(opts, option.WithCredentialsJSON([]byte(creds)))
		}
	}

	client, err := firestore.NewClientWithDatabase(ctx, projectID, dbID, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Firestore client: %w", err)
	}

	return client, nil
}
