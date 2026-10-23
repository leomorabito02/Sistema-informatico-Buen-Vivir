package storage

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
)

// NewFirestoreClient initializes a Firestore client using configured environment credentials.
func NewFirestoreClient(ctx context.Context, projectID string) (*firestore.Client, error) {
	if projectID == "" {
		projectID = os.Getenv("FIREBASE_PROJECT_ID")
		if projectID == "" {
			projectID = "ceibuenvivir"
		}
	}

	credsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	var opts []option.ClientOption

	if credsPath != "" {
		if len(credsPath) > 0 && credsPath[0] == '{' {
			opts = append(opts, option.WithCredentialsJSON([]byte(credsPath))) //nolint:staticcheck
		} else {
			opts = append(opts, option.WithCredentialsFile(credsPath)) //nolint:staticcheck
		}
	}

	client, err := firestore.NewClient(ctx, projectID, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create firestore client: %w", err)
	}
	return client, nil
}
