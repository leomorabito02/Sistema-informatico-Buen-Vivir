package services

import (
	"context"
	"errors"
	"time"

	"buenvivir/labor-integration/internal/models"
	"buenvivir/labor-integration/internal/repositories"
)

type JobApplicationService struct {
	repo       repositories.JobApplicationRepository
	offerRepo  repositories.JobOfferRepository
	authClient repositories.AuthServiceClient
}

func NewJobApplicationService(repo repositories.JobApplicationRepository, offerRepo repositories.JobOfferRepository, authClient repositories.AuthServiceClient) *JobApplicationService {
	return &JobApplicationService{
		repo:       repo,
		offerRepo:  offerRepo,
		authClient: authClient,
	}
}

func (s *JobApplicationService) ApplyToJob(ctx context.Context, application *models.JobApplication) error {
	// Verify participant exists and is approved
	participant, err := s.authClient.GetParticipant(ctx, application.ParticipantID)
	if err != nil {
		return err
	}
	if participant == nil || participant.ApprovalStatus != "approved" {
		return errors.New("participant not approved or does not exist")
	}

	// Prevent duplicate applications
	applied, err := s.repo.HasApplied(ctx, application.OfferID, application.ParticipantID)
	if err != nil {
		return err
	}
	if applied {
		return errors.New("already applied")
	}

	loc, err := time.LoadLocation("America/Argentina/Buenos_Aires")
	if err != nil {
		loc = time.FixedZone("UTC-3", -3*60*60)
	}
	now := time.Now().In(loc)

	application.Status = "Pendiente"
	application.ApplicationDate = now
	application.CreatedAt = now
	application.UpdatedAt = now

	return s.repo.Create(ctx, application)
}

func (s *JobApplicationService) ListByParticipant(ctx context.Context, participantID string) ([]*models.JobApplication, error) {
	return s.repo.ListByParticipant(ctx, participantID)
}

func (s *JobApplicationService) UpdateStatus(ctx context.Context, applicationID string, companyID string, status string) error {
	app, err := s.repo.GetByID(ctx, applicationID)
	if err != nil {
		return err
	}
	
	offer, err := s.offerRepo.GetByID(ctx, app.OfferID)
	if err != nil {
		return err
	}
	
	if offer.CompanyID != companyID {
		return errors.New("unauthorized: you do not own the offer for this application")
	}
	
	return s.repo.UpdateStatus(ctx, applicationID, status)
}


func (s *JobApplicationService) ListByOffer(ctx context.Context, offerID string) ([]*models.JobApplication, error) {
	return s.repo.ListByOffer(ctx, offerID)
}
