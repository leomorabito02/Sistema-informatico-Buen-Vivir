package storage

import (
	"context"
	"fmt"
	"os"
	"strings"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
)

// NewFirestoreClient initializes a Firestore client using configured environment credentials.
func NewFirestoreClient(ctx context.Context, projectID string) (*firestore.Client, error) {
	if projectID == "" {
		projectID = strings.TrimSpace(os.Getenv("FIREBASE_PROJECT_ID"))
		if projectID == "" {
			projectID = strings.TrimSpace(os.Getenv("GCP_PROJECT_ID"))
		}
		if projectID == "" {
			projectID = "ceibuenvivir"
		}
	}

	var opts []option.ClientOption
	if strings.TrimSpace(os.Getenv("FIRESTORE_EMULATOR_HOST")) == "" {
		var err error
		opts, err = ResolveGoogleClientOptions(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
		if err != nil {
			return nil, fmt.Errorf("failed to resolve firestore credentials: %w", err)
		}
	}

	if dbID := resolveFirestoreDatabase(); dbID != "" {
		client, err := firestore.NewClientWithDatabase(ctx, projectID, dbID, opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to create firestore client: %w", err)
		}
		return client, nil
	}

	client, err := firestore.NewClient(ctx, projectID, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create firestore client: %w", err)
	}
	return client, nil
}

// resolveFirestoreDatabase returns the configured Firestore database ID, or empty string for "(default)".
func resolveFirestoreDatabase() string {
	dbID := strings.TrimSpace(os.Getenv("FIRESTORE_DATABASE"))
	if dbID == "" || dbID == "(default)" || strings.EqualFold(dbID, "null") {
		return ""
	}
	return dbID
}
