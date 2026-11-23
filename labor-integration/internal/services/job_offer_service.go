package services

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"buenvivir/labor-integration/api/proto/authpb"
	"buenvivir/labor-integration/internal/events/producers"
	"buenvivir/labor-integration/internal/models"
	"buenvivir/labor-integration/internal/repositories"
)

type JobOfferService struct {
	repo       repositories.JobOfferRepository
	appRepo    repositories.JobApplicationRepository
	authClient repositories.AuthServiceClient
	producer   producers.InsercionLaboralProducer
	skills     repositories.SkillsClient
}

type JobOfferServiceOption func(*JobOfferService)

// ErrOfferForbidden: el rol no puede hacer esa operación sobre la oferta.
var ErrOfferForbidden = errors.New("forbidden: no tenés permisos sobre esta oferta")

// Roles de gestión de ofertas. role_dev pasa como admin.
func esAdminOferta(role string) bool { return role == "role_admin" || role == "role_dev" }
func esStaffOferta(role string) bool { return esAdminOferta(role) || role == "role_volunteer" }

// puedeGestionarOferta: admin y voluntario sobre cualquiera (así lo ofrece
// hoy la pantalla de ofertas); la empresa solo sobre las suyas; el resto,
// ninguna.
func puedeGestionarOferta(offer *models.JobOffer, role, userID string) bool {
	if esStaffOferta(role) {
		return true
	}
	return role == "role_company" && userID != "" && offer.CompanyID == userID
}

// WithOfferProducer sets an InsercionLaboralProducer for JobOfferService.
func WithOfferProducer(producer producers.InsercionLaboralProducer) JobOfferServiceOption {
	return func(s *JobOfferService) {
		s.producer = producer
	}
}

func NewJobOfferService(repo repositories.JobOfferRepository, appRepo repositories.JobApplicationRepository, authClient repositories.AuthServiceClient, opts ...JobOfferServiceOption) *JobOfferService {
	s := &JobOfferService{
		repo:       repo,
		appRepo:    appRepo,
		authClient: authClient,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *JobOfferService) CreateOffer(ctx context.Context, offer *models.JobOffer, role string, userID string) error {
	// HU-39: publican ofertas las empresas (y la gestión del centro).
	if !esStaffOferta(role) && role != "role_company" {
		return ErrOfferForbidden
	}
	if err := validateOfferFields(offer, true); err != nil {
		return err
	}
	if role == "role_company" {
		company, err := s.authClient.GetCompany(ctx, offer.CompanyID)
		if err != nil {
			return err
		}
		if company == nil {
			return errors.New("company profile not found")
		}

		if company.UserEmail == "" || company.UserPhone == "" || company.UserStatus != "active" ||
			company.BusinessName == "" || company.Cuit == "" || strings.HasPrefix(company.Cuit, "PENDING-") || !isApprovedStatus(company.ApprovalStatus) {
			return errors.New("company profile is incomplete or not approved")
		}
	}

	now := time.Now()
	offer.CreatedAt = now
	offer.UpdatedAt = now

	if esStaffOferta(role) {
		offer.Status = "Activa"
		offer.ApprovedBy = userID
		if offer.CompanyID == "" {
			offer.CompanyID = userID
		}
	} else {
		offer.Status = "Pendiente"
	}

	if err := s.repo.Create(ctx, offer); err != nil {
		return err
	}

	if s.producer != nil {
		if err := s.producer.OfertaCreada(ctx, offer.ID, offer.CompanyID, offer.Status); err != nil {
			slog.Warn("failed to emit oferta.creada", slog.String("error", err.Error()), slog.String("offer_id", offer.ID))
		} else {
			s.producer.Notify()
		}
	}
	// HU-44: si nace publicada (la creó la gestión), se busca a quién le sirve.
	s.notifyCompatibles(ctx, offer)

	return nil
}

func (s *JobOfferService) ListOffers(ctx context.Context, companyID string, role string, userID string) ([]*models.JobOffer, error) {
	offers, err := s.repo.List(ctx, companyID)
	if err != nil {
		return nil, err
	}

	companyCache := make(map[string]*authpb.GetCompanyResponse)
	var appliedOffers map[string]bool
	if !esStaffOferta(role) && role != "role_company" && s.appRepo != nil && userID != "" {
		appliedOffers = s.appliedOfferSet(ctx, userID)
	}

	var filteredOffers []*models.JobOffer
	for _, offer := range offers {
		s.checkOfferExpiration(ctx, offer)
		s.enrichCompanyInfo(ctx, offer, role, companyCache)

		if s.isOfferVisibleForRole(offer, role, appliedOffers) {
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

func (s *JobOfferService) enrichCompanyInfo(ctx context.Context, offer *models.JobOffer, role string, cache map[string]*authpb.GetCompanyResponse) {
	if s.authClient == nil || offer.CompanyID == "" {
		return
	}
	comp, cached := cache[offer.CompanyID]
	if !cached {
		fetched, err := s.authClient.GetCompany(ctx, offer.CompanyID)
		if err == nil {
			comp = fetched
		}
		cache[offer.CompanyID] = comp
	}
	if comp == nil {
		return
	}
	if offer.CreatorName == "" && comp.BusinessName != "" {
		offer.CreatorName = comp.BusinessName
	}
	if esStaffOferta(role) {
		offer.CompanyEmail = comp.UserEmail
		offer.CompanyPhone = comp.UserPhone
	}
}

func (s *JobOfferService) appliedOfferSet(ctx context.Context, userID string) map[string]bool {
	ids := []string{userID}
	if s.authClient != nil {
		if participant, err := s.authClient.GetParticipant(ctx, userID); err == nil && participant != nil && participant.Id != "" {
			ids = append(ids, participant.Id)
		}
	}
	apps, err := s.appRepo.ListByParticipant(ctx, ids)
	if err != nil {
		return nil
	}
	set := make(map[string]bool, len(apps))
	for _, app := range apps {
		if app != nil && app.OfferID != "" {
			set[app.OfferID] = true
		}
	}
	return set
}

// isOfferVisibleForRole restricts pending, rejected, and deleted offers to staff
// and the owning company; participants see active offers plus finalized ones they applied to.
func (s *JobOfferService) isOfferVisibleForRole(offer *models.JobOffer, role string, appliedOffers map[string]bool) bool {
	if esStaffOferta(role) || role == "role_company" {
		return true
	}
	if offer.Status == "Activa" {
		return true
	}
	return (offer.Status == "Vencida" || offer.Status == "Cerrada") && appliedOffers[offer.ID]
}

func (s *JobOfferService) UpdateOffer(ctx context.Context, id string, offer *models.JobOffer, role, userID string) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return repositories.ErrOfferNotFound
	}
	if !puedeGestionarOferta(existing, role, userID) {
		return ErrOfferForbidden
	}
	previousStatus := existing.Status
	offer.Status = estadoPermitido(existing.Status, offer.Status, role)
	reactivating := (previousStatus == "Vencida" || previousStatus == "Cerrada") && offer.Status == "Activa"
	if err := validateOfferFields(offer, reactivating); err != nil {
		return err
	}

	now := time.Now()
	existing.Title = offer.Title
	existing.Description = offer.Description
	existing.Location = offer.Location
	existing.Modality = offer.Modality
	existing.JobType = offer.JobType
	existing.Salary = offer.Salary
	existing.Schedule = offer.Schedule
	existing.Vacancies = offer.Vacancies
	existing.Requirements = offer.Requirements
	existing.Deadline = offer.Deadline
	existing.Status = offer.Status
	existing.UpdatedAt = now

	if offer.Status == "Activa" && previousStatus != "Activa" {
		existing.ClosedAt = nil
		existing.ClosedBy = ""
	} else if offer.Status == "Cerrada" && previousStatus != "Cerrada" {
		existing.ClosedAt = &now
		existing.ClosedBy = userID
	}

	if err := s.repo.Update(ctx, id, existing); err != nil {
		return err
	}
	*offer = *existing
	return nil
}

// CloseOffer closes an active offer before its deadline.
func (s *JobOfferService) CloseOffer(ctx context.Context, id, role, userID string) (*models.JobOffer, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, repositories.ErrOfferNotFound
	}
	if !puedeGestionarOferta(existing, role, userID) {
		return nil, ErrOfferForbidden
	}
	s.checkOfferExpiration(ctx, existing)
	if existing.Status != "Activa" {
		return nil, &offerFieldError{msg: "solo se pueden cerrar ofertas activas"}
	}
	now := time.Now()
	if err := s.repo.Close(ctx, id, userID, now); err != nil {
		return nil, err
	}
	existing.Status = "Cerrada"
	existing.ClosedBy = userID
	existing.ClosedAt = &now
	existing.UpdatedAt = now
	return existing, nil
}

// estadoPermitido decides the resulting status on update. Only admin can change
// status freely; company and volunteer may reactivate Vencida or Cerrada offers.
func estadoPermitido(actual, pedido, role string) string {
	if pedido == "" {
		return actual
	}
	if esAdminOferta(role) {
		return pedido
	}
	if (actual == "Vencida" || actual == "Cerrada") && pedido == "Activa" {
		return pedido
	}
	return actual
}

func (s *JobOfferService) DeleteOffer(ctx context.Context, id, role, userID string) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if !puedeGestionarOferta(existing, role, userID) {
		return ErrOfferForbidden
	}
	hasApps, err := s.appRepo.HasApplicationsForOffer(ctx, id)
	if err != nil {
		return err
	}
	if hasApps {
		err = s.repo.UpdateStatus(ctx, id, "Eliminada") // Soft delete
	} else {
		err = s.repo.Delete(ctx, id) // Hard delete
	}
	if err != nil {
		return err
	}
	// HU-41: si la borró la gestión del centro, se avisa a la empresa.
	if s.producer != nil {
		if err := s.producer.OfertaEliminada(ctx, existing.ID, existing.CompanyID, esStaffOferta(role)); err != nil {
			slog.Warn("failed to emit oferta.eliminada", slog.String("error", err.Error()), slog.String("offer_id", id))
		} else {
			s.producer.Notify()
		}
	}
	return nil
}

func (s *JobOfferService) ApproveOffer(ctx context.Context, id string, approvedBy string) error {
	if err := s.repo.UpdateStatusAndApprover(ctx, id, "Activa", approvedBy); err != nil {
		return err
	}
	offer, err := s.repo.GetByID(ctx, id)
	if err == nil && offer != nil && s.producer != nil {
		if err := s.producer.OfertaRevisada(ctx, offer.ID, offer.CompanyID, true); err != nil {
			slog.Warn("failed to emit oferta.revisada", slog.String("error", err.Error()), slog.String("offer_id", offer.ID))
		} else {
			s.producer.Notify()
		}
	}
	if err == nil && offer != nil {
		s.notifyCompatibles(ctx, offer) // HU-44
	}
	return nil
}

func (s *JobOfferService) RejectOffer(ctx context.Context, id string, approvedBy string) error {
	if err := s.repo.UpdateStatusAndApprover(ctx, id, "Rechazada", approvedBy); err != nil {
		return err
	}
	offer, err := s.repo.GetByID(ctx, id)
	if err == nil && offer != nil && s.producer != nil {
		if err := s.producer.OfertaRevisada(ctx, offer.ID, offer.CompanyID, false); err != nil {
			slog.Warn("failed to emit oferta.revisada", slog.String("error", err.Error()), slog.String("offer_id", offer.ID))
		} else {
			s.producer.Notify()
		}
	}
	return nil
}
