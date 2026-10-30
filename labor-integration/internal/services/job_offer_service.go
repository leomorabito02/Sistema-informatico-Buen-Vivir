package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"buenvivir/labor-integration/internal/models"
	"buenvivir/labor-integration/internal/repositories"
)

type JobOfferService struct {
	repo       repositories.JobOfferRepository
	appRepo    repositories.JobApplicationRepository
	authClient repositories.AuthServiceClient
}

func NewJobOfferService(repo repositories.JobOfferRepository, appRepo repositories.JobApplicationRepository, authClient repositories.AuthServiceClient) *JobOfferService {
	return &JobOfferService{repo: repo, appRepo: appRepo, authClient: authClient}
}

func (s *JobOfferService) CreateOffer(ctx context.Context, offer *models.JobOffer, role string, userID string) error {
	if role == "role_company" {
		company, err := s.authClient.GetCompany(ctx, offer.CompanyID)
		if err != nil {
			return err
		}
		if company == nil {
			return errors.New("company profile not found")
		}
		
		if company.UserEmail == "" || company.UserPhone == "" || company.UserStatus != "active" ||
			company.BusinessName == "" || company.Cuit == "" || strings.HasPrefix(company.Cuit, "PENDING-") || company.ApprovalStatus != "approved" {
			return errors.New("company profile is incomplete or not approved")
		}
	}

	now := time.Now()
	offer.CreatedAt = now
	offer.UpdatedAt = now
	
	if role == "role_admin" || role == "role_volunteer" {
		offer.Status = "Activa"
		offer.ApprovedBy = userID
	} else {
		offer.Status = "Pendiente"
	}
	
	return s.repo.Create(ctx, offer)
}

func (s *JobOfferService) ListOffers(ctx context.Context, companyID string, role string) ([]*models.JobOffer, error) {
	offers, err := s.repo.List(ctx, companyID)
	if err != nil {
		return nil, err
	}

	var filteredOffers []*models.JobOffer
	for _, offer := range offers {
		s.checkOfferExpiration(ctx, offer)
		s.populateMissingCreatorName(ctx, offer)

		if s.isOfferVisibleForRole(offer, role) {
			filteredOffers = append(filteredOffers, offer)
		}
	}

	return filteredOffers, nil
}

func (s *JobOfferService) checkOfferExpiration(ctx context.Context, offer *models.JobOffer) {
	if offer.Status == "Activa" && !offer.Deadline.IsZero() && time.Now().After(offer.Deadline) {
		offer.Status = "Vencida"
		_ = s.repo.UpdateStatus(ctx, offer.ID, "Vencida")
	}
}

func (s *JobOfferService) populateMissingCreatorName(ctx context.Context, offer *models.JobOffer) {
	if offer.CreatorName == "" && offer.CompanyID != "" {
		comp, err := s.authClient.GetCompany(ctx, offer.CompanyID)
		if err == nil && comp != nil && comp.BusinessName != "" {
			offer.CreatorName = comp.BusinessName
		}
	}
}

func (s *JobOfferService) isOfferVisibleForRole(offer *models.JobOffer, role string) bool {
	if role == "role_participant" {
		return offer.Status == "Activa" || offer.Status == "Vencida"
	}
	return true
}

func (s *JobOfferService) UpdateOffer(ctx context.Context, id string, offer *models.JobOffer) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	existing.Title = offer.Title
	existing.Description = offer.Description
	existing.Location = offer.Location
	existing.Modality = offer.Modality
	existing.JobType = offer.JobType
	existing.Salary = offer.Salary
	existing.Schedule = offer.Schedule
	existing.Requirements = offer.Requirements
	existing.Deadline = offer.Deadline
	existing.Status = offer.Status
	existing.UpdatedAt = time.Now()

	return s.repo.Update(ctx, id, existing)
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

func (s *JobOfferService) ApproveOffer(ctx context.Context, id string, approvedBy string) error {
	return s.repo.UpdateStatusAndApprover(ctx, id, "Activa", approvedBy)
}

func (s *JobOfferService) RejectOffer(ctx context.Context, id string, approvedBy string) error {
	return s.repo.UpdateStatusAndApprover(ctx, id, "Rechazada", approvedBy)
}
