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
	authClient repositories.AuthServiceClient
}

func NewJobApplicationService(repo repositories.JobApplicationRepository, authClient repositories.AuthServiceClient) *JobApplicationService {
	return &JobApplicationService{
		repo:       repo,
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

	application.Status = "Pendiente"
	application.ApplicationDate = time.Now()

	return s.repo.Create(ctx, application)
}
