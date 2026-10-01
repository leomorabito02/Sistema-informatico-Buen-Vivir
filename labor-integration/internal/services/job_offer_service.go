package services

import (
	"context"
	"time"

	"buenvivir/labor-integration/internal/models"
	"buenvivir/labor-integration/internal/repositories"
)

type JobOfferService struct {
	repo    repositories.JobOfferRepository
	appRepo repositories.JobApplicationRepository
}

func NewJobOfferService(repo repositories.JobOfferRepository, appRepo repositories.JobApplicationRepository) *JobOfferService {
	return &JobOfferService{repo: repo, appRepo: appRepo}
}

func (s *JobOfferService) CreateOffer(ctx context.Context, offer *models.JobOffer) error {
	now := time.Now()
	offer.CreatedAt = now
	offer.UpdatedAt = now
	return s.repo.Create(ctx, offer)
}

func (s *JobOfferService) ListOffers(ctx context.Context, companyID string) ([]*models.JobOffer, error) {
	offers, err := s.repo.List(ctx, companyID)
	if err != nil {
		return nil, err
	}

	for _, offer := range offers {
		if offer.Status == "Activa" && !offer.Deadline.IsZero() && time.Now().After(offer.Deadline) {
			offer.Status = "Vencida"
			_ = s.repo.UpdateStatus(ctx, offer.ID, "Vencida")
		}
	}

	return offers, nil
}

func (s *JobOfferService) UpdateOffer(ctx context.Context, id string, offer *models.JobOffer) error {
	offer.UpdatedAt = time.Now()
	return s.repo.Update(ctx, id, offer)
}

func (s *JobOfferService) DeleteOffer(ctx context.Context, id string) error {
	hasApps, err := s.appRepo.HasApplicationsForOffer(ctx, id)
	if err != nil {
		return err
	}
	if hasApps {
		return s.repo.UpdateStatus(ctx, id, "Eliminada") // Soft delete
	}
	return s.repo.Delete(ctx, id) // Hard delete
}
