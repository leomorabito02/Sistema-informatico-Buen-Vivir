package services

import (
	"context"

	"buenvivir/labor-integration/internal/models"
	"buenvivir/labor-integration/internal/repositories"
)

type JobOfferService struct {
	repo repositories.JobOfferRepository
}

func NewJobOfferService(repo repositories.JobOfferRepository) *JobOfferService {
	return &JobOfferService{repo: repo}
}

func (s *JobOfferService) CreateOffer(ctx context.Context, offer *models.JobOffer) error {
	return s.repo.Create(ctx, offer)
}

func (s *JobOfferService) ListOffers(ctx context.Context) ([]*models.JobOffer, error) {
	return s.repo.List(ctx)
}
