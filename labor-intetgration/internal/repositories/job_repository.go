package repositories

import (
	"context"

	"buenvivir/labor-integration/internal/models"
)

type JobRepository interface {
	FindJobs(ctx context.Context) ([]*models.JobOffer, error)
}

type jobRepository struct{}

func NewJobRepository() JobRepository {
	return &jobRepository{}
}

func (r *jobRepository) FindJobs(ctx context.Context) ([]*models.JobOffer, error) {
	return []*models.JobOffer{
		{ID: "job-1", Title: "Auxiliar Administrativo", Company: "Cooperativa Buen Vivir", Location: "Córdoba", Type: "Full-time"},
	}, nil
}
