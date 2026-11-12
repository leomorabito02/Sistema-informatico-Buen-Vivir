package controllers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"buenvivir/labor-integration/internal/dto"
	"buenvivir/labor-integration/internal/models"
	"buenvivir/labor-integration/internal/repositories"
	"buenvivir/labor-integration/internal/services"

	"github.com/go-chi/chi/v5"
)

const headerUserID = "X-User-ID"

type JobOfferController struct {
	service *services.JobOfferService
}

func NewJobOfferController(service *services.JobOfferService) *JobOfferController {
	return &JobOfferController{service: service}
}

func (c *JobOfferController) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateJobOfferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	role := r.Header.Get(headerUserRole)
	userID := r.Header.Get(headerUserID)

	offer := models.JobOffer{
		CreatorName:    req.CreatorName,
		Title:          req.Title,
		Description:    req.Description,
		Location:       req.Location,
		Modality:       req.Modality,
		JobType:        req.JobType,
		Salary:         req.Salary,
		Schedule:       req.Schedule,
		Requirements:   req.Requirements,
		Deadline:       req.Deadline,
	}

	if role == "role_company" {
		offer.CompanyID = userID
	}

	if err := c.service.CreateOffer(r.Context(), &offer, role, userID); err != nil {
		if errors.Is(err, services.ErrOfferForbidden) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(offer)
}

func (c *JobOfferController) List(w http.ResponseWriter, r *http.Request) {
	role := r.Header.Get(headerUserRole)
	companyID := ""
	if role == "role_company" {
		companyID = r.Header.Get(headerUserID)
	}

	// ?q= busca en título, empresa y ubicación; ?empresa= deja solo las
	// ofertas de esa empresa (HU-45, HU-46).
	q := r.URL.Query().Get("q")
	empresa := r.URL.Query().Get("empresa")
	if len([]rune(q)) > services.MaxBusquedaOfertas || len([]rune(empresa)) > services.MaxBusquedaOfertas {
		http.Error(w, "search too long", http.StatusBadRequest)
		return
	}

	offers, err := c.service.ListOffers(r.Context(), companyID, role)
	if err != nil {
		http.Error(w, "failed to list offers", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(services.FiltrarOfertas(offers, q, empresa))
}

func (c *JobOfferController) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req dto.UpdateJobOfferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	offer := models.JobOffer{
		Title:          req.Title,
		Description:    req.Description,
		Location:       req.Location,
		Modality:       req.Modality,
		JobType:        req.JobType,
		Salary:         req.Salary,
		Schedule:       req.Schedule,
		Requirements:   req.Requirements,
		Deadline:       req.Deadline,
		Status:         req.Status,
	}


	if err := c.service.UpdateOffer(r.Context(), id, &offer, r.Header.Get(headerUserRole), r.Header.Get(headerUserID)); err != nil {
		writeOfferError(w, err, "failed to update offer")
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(offer)
}

func (c *JobOfferController) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := c.service.DeleteOffer(r.Context(), id, r.Header.Get(headerUserRole), r.Header.Get(headerUserID)); err != nil {
		writeOfferError(w, err, "failed to delete offer")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *JobOfferController) Approve(w http.ResponseWriter, r *http.Request) {
	// Aprobar y rechazar es solo del administrador (HU-41); el voluntario
	// consulta ofertas sin poder aprobarlas (HU-06).
	role := r.Header.Get(headerUserRole)
	if role != "role_admin" && role != "role_dev" {
		http.Error(w, "unauthorized", http.StatusForbidden)
		return
	}
	
	id := chi.URLParam(r, "id")
	userID := r.Header.Get(headerUserID)
	
	if err := c.service.ApproveOffer(r.Context(), id, userID); err != nil {
		http.Error(w, "failed to approve offer", http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
}

func (c *JobOfferController) Reject(w http.ResponseWriter, r *http.Request) {
	role := r.Header.Get(headerUserRole)
	if role != "role_admin" && role != "role_dev" {
		http.Error(w, "unauthorized", http.StatusForbidden)
		return
	}
	
	id := chi.URLParam(r, "id")
	userID := r.Header.Get(headerUserID)
	
	if err := c.service.RejectOffer(r.Context(), id, userID); err != nil {
		http.Error(w, "failed to reject offer", http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
}

// writeOfferError: 403 sin permisos, 404 si la oferta no existe y 500 sin
// exponer el detalle para lo demás.
func writeOfferError(w http.ResponseWriter, err error, msg string) {
	switch {
	case errors.Is(err, services.ErrOfferForbidden):
		http.Error(w, "forbidden", http.StatusForbidden)
	case errors.Is(err, repositories.ErrOfferNotFound):
		http.Error(w, "offer not found", http.StatusNotFound)
	default:
		slog.Error(msg, slog.String("error", err.Error()))
		http.Error(w, msg, http.StatusInternalServerError)
	}
}
