package storage

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

type ObjectStorageService interface {
	UploadFile(ctx context.Context, objectName string, content io.Reader, contentType string) (string, error)
	DownloadFile(ctx context.Context, objectName string) (io.ReadCloser, error)
	DeleteFile(ctx context.Context, objectName string) error
}

type gcsStorageService struct {
	client     *storage.Client
	bucketName string
	emulator   string
}

// NewObjectStorageService initializes a GCP Cloud Storage client with local emulator fallback support.
func NewObjectStorageService(ctx context.Context, projectID, bucketName, emulatorHost, credsPath string) (ObjectStorageService, error) {
	opts, err := resolveStorageClientOptions(emulatorHost, credsPath)
	if err != nil {
		return nil, err
	}

	client, err := storage.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage client: %w", err)
	}

	return &gcsStorageService{
		client:     client,
		bucketName: bucketName,
		emulator:   emulatorHost,
	}, nil
}

// resolveStorageClientOptions builds Google Storage client options from emulator settings,
// explicit credentials path/JSON, or base64-encoded service account environment variables.
func resolveStorageClientOptions(emulatorHost, credsPath string) ([]option.ClientOption, error) {
	if emulatorHost != "" {
		return []option.ClientOption{
			option.WithEndpoint(emulatorHost),
			option.WithoutAuthentication(),
		}, nil
	}

	creds := strings.TrimSpace(credsPath)
	if creds == "" {
		for _, envKey := range []string{"GOOGLE_APPLICATION_CREDENTIALS_BASE64", "FIREBASE_CREDENTIALS_BASE64"} {
			if b64 := strings.TrimSpace(os.Getenv(envKey)); b64 != "" {
				decoded, err := base64.StdEncoding.DecodeString(b64)
				if err != nil {
					return nil, fmt.Errorf("error decoding %s: %w", envKey, err)
				}
				creds = strings.TrimSpace(string(decoded))
				break
			}
		}
	}

	if creds == "" {
		return nil, nil
	}
	if strings.HasPrefix(creds, "{") {
		return []option.ClientOption{option.WithCredentialsJSON([]byte(creds))}, nil //nolint:staticcheck // mismo criterio que firebase_service.go: credenciales inline
	}
	return []option.ClientOption{option.WithCredentialsFile(creds)}, nil //nolint:staticcheck // mismo criterio que firebase_service.go: archivo de credenciales
}

// UploadFile uploads a reader stream to GCP Object Storage and returns the public or emulated file URL.
func (s *gcsStorageService) UploadFile(ctx context.Context, objectName string, content io.Reader, contentType string) (string, error) {
	bucket := s.client.Bucket(s.bucketName)
	obj := bucket.Object(objectName)

	wc := obj.NewWriter(ctx)
	wc.ContentType = contentType

	if _, err := io.Copy(wc, content); err != nil {
		_ = wc.Close()
		return "", fmt.Errorf("failed to write object bytes: %w", err)
	}

	if err := wc.Close(); err != nil {
		return "", fmt.Errorf("failed to close object writer: %w", err)
	}

	if s.emulator != "" {
		return fmt.Sprintf("%s/%s/%s", s.emulator, s.bucketName, objectName), nil
	}

	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", s.bucketName, objectName), nil
}

// DeleteFile removes an object from GCP Object Storage.
func (s *gcsStorageService) DeleteFile(ctx context.Context, objectName string) error {
	bucket := s.client.Bucket(s.bucketName)
	return bucket.Object(objectName).Delete(ctx)
}

// DownloadFile gets a reader stream from GCP Object Storage.
func (s *gcsStorageService) DownloadFile(ctx context.Context, objectName string) (io.ReadCloser, error) {
	bucket := s.client.Bucket(s.bucketName)
	return bucket.Object(objectName).NewReader(ctx)
}
