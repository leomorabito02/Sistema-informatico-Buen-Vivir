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

type JobOfferService struct {
	repo       repositories.JobOfferRepository
	appRepo    repositories.JobApplicationRepository
	authClient repositories.AuthServiceClient
	producer   producers.InsercionLaboralProducer
}

type JobOfferServiceOption func(*JobOfferService)

// ErrOfferForbidden: el rol no puede hacer esa operación sobre la oferta.
// Antes editar y borrar no controlaban rol ni dueño (auditoría
// 2026-11-11, Problema 4).
var ErrOfferForbidden = errors.New("forbidden: no tenés permisos sobre esta oferta")

// Roles de gestión de ofertas. role_dev pasa como admin.
func esAdminOferta(role string) bool { return role == "role_admin" || role == "role_dev" }
func esStaffOferta(role string) bool  { return esAdminOferta(role) || role == "role_volunteer" }

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

	return nil
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

// isOfferVisibleForRole: pendientes, rechazadas y eliminadas solo las ve
// la gestión (admin, voluntario) y la empresa dueña (el controller ya le
// filtra las suyas). El resto ve solo las publicadas (HU-40). Antes
// cualquier rol que no fuera participante (docente, por ejemplo) veía
// también las pendientes de todas las empresas.
func (s *JobOfferService) isOfferVisibleForRole(offer *models.JobOffer, role string) bool {
	if esStaffOferta(role) || role == "role_company" {
		return true
	}
	return offer.Status == "Activa" || offer.Status == "Vencida"
}

func (s *JobOfferService) UpdateOffer(ctx context.Context, id string, offer *models.JobOffer, role, userID string) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if !puedeGestionarOferta(existing, role, userID) {
		return ErrOfferForbidden
	}
	offer.Status = estadoPermitido(existing.Status, offer.Status, role)

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

// estadoPermitido decide el estado después de editar. Solo el admin lo
// cambia libremente (aprobar es suyo, HU-40/41). Empresa y voluntario
// solo pueden renovar una oferta vencida ("Vencida" -> "Activa"); otro
// cambio se ignora. Un estado vacío deja el que estaba (antes lo borraba).
func estadoPermitido(actual, pedido, role string) string {
	if pedido == "" {
		return actual
	}
	if esAdminOferta(role) {
		return pedido
	}
	if actual == "Vencida" && pedido == "Activa" {
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
		return s.repo.UpdateStatus(ctx, id, "Eliminada") // Soft delete
	}
	return s.repo.Delete(ctx, id) // Hard delete
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
