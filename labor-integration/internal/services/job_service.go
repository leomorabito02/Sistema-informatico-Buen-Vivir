package services

import (
	"context"

	"buenvivir/labor-integration/internal/models"
	"buenvivir/labor-integration/internal/repositories"
)

type JobService interface {
	GetJobs(ctx context.Context) ([]*models.JobOffer, error)
}

type jobService struct {
	repo repositories.JobRepository
}

func NewJobService(repo repositories.JobRepository) JobService {
	return &jobService{repo: repo}
}

func (s *jobService) GetJobs(ctx context.Context) ([]*models.JobOffer, error) {
	return s.repo.FindJobs(ctx)
}
