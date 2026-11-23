package services

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"buenvivir/labor-integration/internal/events/producers"
	"buenvivir/labor-integration/internal/models"
	"buenvivir/labor-integration/internal/repositories"
)

type JobApplicationService struct {
	repo       repositories.JobApplicationRepository
	offerRepo  repositories.JobOfferRepository
	authClient repositories.AuthServiceClient
	producer   producers.InsercionLaboralProducer
}

type JobApplicationServiceOption func(*JobApplicationService)

// WithApplicationProducer sets an InsercionLaboralProducer for JobApplicationService.
func WithApplicationProducer(producer producers.InsercionLaboralProducer) JobApplicationServiceOption {
	return func(s *JobApplicationService) {
		s.producer = producer
	}
}

func NewJobApplicationService(repo repositories.JobApplicationRepository, offerRepo repositories.JobOfferRepository, authClient repositories.AuthServiceClient, opts ...JobApplicationServiceOption) *JobApplicationService {
	s := &JobApplicationService{
		repo:       repo,
		offerRepo:  offerRepo,
		authClient: authClient,
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
	if participant == nil || !isApprovedStatus(participant.ApprovalStatus) {
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

	if s.producer != nil {
		partID := application.ParticipantID
		if partID == "" {
			partID = application.UserID
		}
		empresaID := ""
		if s.offerRepo != nil {
			if offer, err := s.offerRepo.GetByID(ctx, application.OfferID); err == nil && offer != nil {
				empresaID = offer.CompanyID
			}
		}
		if err := s.producer.PostulacionCreada(ctx, application.ID, application.OfferID, partID, empresaID); err != nil {
			slog.Warn("failed to emit postulacion.creada", slog.String("error", err.Error()), slog.String("application_id", application.ID))
		} else {
			s.producer.Notify()
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
	if app == nil {
		return repositories.ErrApplicationNotFound
	}

	offer, err := s.offerRepo.GetByID(ctx, app.OfferID)
	if err != nil {
		return err
	}
	if offer == nil {
		return repositories.ErrOfferNotFound
	}

	if offer.CompanyID != companyID {
		return errors.New("unauthorized: you do not own the offer for this application")
	}

	previousStatus := app.Status
	if err := s.repo.UpdateStatus(ctx, applicationID, status); err != nil {
		return err
	}

	if s.producer != nil && isStatusSelected(status) && !isStatusSelected(previousStatus) {
		partID := app.ParticipantID
		if partID == "" {
			partID = app.UserID
		}
		if err := s.producer.PostulacionSeleccionada(ctx, app.ID, app.OfferID, partID, offer.CompanyID); err != nil {
			slog.Warn("failed to emit postulacion.seleccionada", slog.String("error", err.Error()), slog.String("application_id", app.ID))
		} else {
			s.producer.Notify()
		}
	}

	if s.producer != nil && previousStatus != status {
		partID := app.ParticipantID
		if partID == "" {
			partID = app.UserID
		}
		if err := s.producer.PostulacionDecidida(ctx, app.ID, app.OfferID, partID, offer.CompanyID, status); err != nil {
			slog.Warn("failed to emit postulacion.decidida", slog.String("error", err.Error()), slog.String("application_id", app.ID))
		} else {
			s.producer.Notify()
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
	apps, err := s.repo.ListByOffer(ctx, offerID)
	if err != nil {
		return nil, err
	}
	s.fillApplicantNames(ctx, apps)
	return apps, nil
}

// fillApplicantNames completa el nombre de cada postulante con auth
// (best effort: si auth no responde, la lista sale igual sin nombres).
func (s *JobApplicationService) fillApplicantNames(ctx context.Context, apps []*models.JobApplication) {
	if s.authClient == nil || len(apps) == 0 {
		return
	}
	ids := make([]string, 0, len(apps))
	seen := map[string]bool{}
	for _, a := range apps {
		id := applicantUserID(a)
		if id != "" && !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return
	}
	names, _ := s.authClient.GetUserNames(ctx, ids)
	for _, a := range apps {
		if n := names[applicantUserID(a)]; n != "" {
			a.Nombre = n
		}
	}
}

// applicantUserID es el id de usuario (usr-…) del postulante; las
// postulaciones viejas lo tenían en participante_id.
func applicantUserID(a *models.JobApplication) string {
	if a == nil {
		return ""
	}
	if strings.HasPrefix(a.UserID, "usr-") {
		return a.UserID
	}
	if strings.HasPrefix(a.ParticipantID, "usr-") {
		return a.ParticipantID
	}
	return ""
}

func (s *JobApplicationService) DeleteApplication(ctx context.Context, appID string, role string, userID string) error {
	app, err := s.repo.GetByID(ctx, appID)
	if err != nil {
		return err
	}
	if app == nil {
		return repositories.ErrApplicationNotFound
	}

	if role != "role_admin" && role != "role_volunteer" {
		if app.UserID != userID && app.ParticipantID != userID {
			return errors.New("unauthorized: you do not have permission to delete this application")
		}
	}

	return s.repo.Delete(ctx, appID)
}

func isStatusSelected(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "aceptado", "aceptada", "accepted", "aprobada", "aprobado", "seleccionado", "contratado":
		return true
	default:
		return false
	}
}
