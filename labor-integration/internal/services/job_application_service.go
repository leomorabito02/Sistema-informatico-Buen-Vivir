package services

import (
	"context"
	"errors"
	"time"

	"buenvivir/labor-integration/internal/models"
	"buenvivir/labor-integration/internal/repositories"
)

type JobApplicationService struct {
	repo         repositories.JobApplicationRepository
	offerRepo    repositories.JobOfferRepository
	authClient   repositories.AuthServiceClient
	dispatcher   NotificationDispatcher
	syncDispatch bool
}

type JobApplicationServiceOption func(*JobApplicationService)

// WithApplicationDispatcher sets a NotificationDispatcher for JobApplicationService.
func WithApplicationDispatcher(dispatcher NotificationDispatcher) JobApplicationServiceOption {
	return func(s *JobApplicationService) {
		s.dispatcher = dispatcher
	}
}

// WithApplicationSyncDispatch sets synchronous dispatching for tests.
func WithApplicationSyncDispatch(sync bool) JobApplicationServiceOption {
	return func(s *JobApplicationService) {
		s.syncDispatch = sync
	}
}

func NewJobApplicationService(repo repositories.JobApplicationRepository, offerRepo repositories.JobOfferRepository, authClient repositories.AuthServiceClient, opts ...JobApplicationServiceOption) *JobApplicationService {
	s := &JobApplicationService{
		repo:       repo,
		offerRepo:  offerRepo,
		authClient: authClient,
		dispatcher: NewNoopNotificationDispatcher(),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
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
	// Auth acepta el id de participante o el de usuario: se guardan los
	// dos ya resueltos, así el resto de los módulos cruza por participante.
	application.ParticipantID = participant.Id
	application.UserID = participant.UserId

	// Prevent duplicate applications (las viejas tienen el id de usuario)
	applied, err := s.repo.HasApplied(ctx, application.OfferID, []string{participant.Id, participant.UserId})
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

	if err := s.repo.Create(ctx, application); err != nil {
		return err
	}

	if s.dispatcher != nil && s.offerRepo != nil {
		offer, err := s.offerRepo.GetByID(ctx, application.OfferID)
		if err == nil && offer != nil {
			if s.syncDispatch {
				s.dispatcher.ApplicationReceived(ctx, application, offer, application.UserID)
			} else {
				go s.dispatcher.ApplicationReceived(context.Background(), application, offer, application.UserID)
			}
		}
	}

	return nil
}

// ListByParticipant recibe el id de usuario del token. Busca por ese id
// (postulaciones viejas) y por el id de participante que le corresponde.
func (s *JobApplicationService) ListByParticipant(ctx context.Context, userID string) ([]*models.JobApplication, error) {
	ids := []string{userID}
	if participant, err := s.authClient.GetParticipant(ctx, userID); err == nil && participant != nil && participant.Id != "" {
		ids = append(ids, participant.Id)
	}
	return s.repo.ListByParticipant(ctx, ids)
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

	previousStatus := app.Status
	if err := s.repo.UpdateStatus(ctx, applicationID, status); err != nil {
		return err
	}

	if s.dispatcher != nil && previousStatus != status {
		app.Status = status
		if s.syncDispatch {
			s.dispatcher.ApplicationDecided(ctx, app, offer, status)
		} else {
			go s.dispatcher.ApplicationDecided(context.Background(), app, offer, status)
		}
	}
	
	return nil
}


func (s *JobApplicationService) ListByOffer(ctx context.Context, offerID string, role string, userID string) ([]*models.JobApplication, error) {
	if role == "role_company" {
		offer, err := s.offerRepo.GetByID(ctx, offerID)
		if err != nil {
			return nil, err
		}
		if offer.CompanyID != userID {
			return nil, errors.New("unauthorized: you do not own this offer")
		}
	}
	// admin and volunteer can view all applications, participant maybe not, but we allow them or we can restrict.
	// We'll let admin, volunteer and the owner company view it.
	return s.repo.ListByOffer(ctx, offerID)
}

func (s *JobApplicationService) DeleteApplication(ctx context.Context, appID string, role string, userID string) error {
	app, err := s.repo.GetByID(ctx, appID)
	if err != nil {
		return err
	}

	if role != "role_admin" && role != "role_volunteer" {
		if app.UserID != userID && app.ParticipantID != userID {
			return errors.New("unauthorized: you do not have permission to delete this application")
		}
	}

	return s.repo.Delete(ctx, appID)
}
