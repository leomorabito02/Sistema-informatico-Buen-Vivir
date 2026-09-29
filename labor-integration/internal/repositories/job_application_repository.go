package repositories

import (
	"context"

	"cloud.google.com/go/firestore"

	"buenvivir/labor-integration/internal/models"
)

type JobApplicationRepository interface {
	Create(ctx context.Context, application *models.JobApplication) error
}

type jobApplicationRepository struct {
	client *firestore.Client
}

func NewJobApplicationRepository(client *firestore.Client) JobApplicationRepository {
	return &jobApplicationRepository{client: client}
}

func (r *jobApplicationRepository) Create(ctx context.Context, application *models.JobApplication) error {
	ref := r.client.Collection("job_applications").NewDoc()
	application.ID = ref.ID
	_, err := ref.Set(ctx, application)
	return err
}
