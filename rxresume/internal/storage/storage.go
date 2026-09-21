package storage

import (
	"context"
	"fmt"
	"io"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

type ObjectStorageService interface {
	UploadFile(ctx context.Context, objectName string, content io.Reader, contentType string) (string, error)
	DeleteFile(ctx context.Context, objectName string) error
}

type gcsStorageService struct {
	client     *storage.Client
	bucketName string
	emulator   string
}

// NewObjectStorageService initializes a GCP Cloud Storage client with local emulator fallback support.
func NewObjectStorageService(ctx context.Context, projectID, bucketName, emulatorHost, credsPath string) (ObjectStorageService, error) {
	var opts []option.ClientOption

	if emulatorHost != "" {
		opts = append(opts, option.WithEndpoint(emulatorHost), option.WithoutAuthentication())
	} else if credsPath != "" {
		opts = append(opts, option.WithCredentialsFile(credsPath))
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
