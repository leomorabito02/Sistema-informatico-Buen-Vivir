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

const headerUserRole = "X-User-Role"

type JobApplicationController struct {
	service *services.JobApplicationService
}

func NewJobApplicationController(service *services.JobApplicationService) *JobApplicationController {
	return &JobApplicationController{service: service}
}

func (c *JobApplicationController) Apply(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateJobApplicationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	application := models.JobApplication{
		OfferID:       req.OfferID,
		ParticipantID: req.ParticipantID,
	}

	role := r.Header.Get(headerUserRole)
	if role != "role_admin" && role != "role_volunteer" {
		// Seguridad: Sobrescribir el ParticipantID con el ID autenticado desde el Gateway si no es admin/voluntario
		application.ParticipantID = r.Header.Get(headerUserID)
	}

	if err := c.service.ApplyToJob(r.Context(), &application); err != nil {
		if err.Error() == "already applied" {
			http.Error(w, "already applied", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(application)
}

func (c *JobApplicationController) ListByOffer(w http.ResponseWriter, r *http.Request) {
	offerID := chi.URLParam(r, "id")
	role := r.Header.Get(headerUserRole)
	userID := r.Header.Get(headerUserID)
	
	applications, err := c.service.ListByOffer(r.Context(), offerID, role, userID)
	if err != nil {
		if err.Error() == "unauthorized: you do not own this offer" {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, "failed to list applications", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(applications)
}

func (c *JobApplicationController) ListMine(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get(headerUserID)
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	applications, err := c.service.ListByParticipant(r.Context(), userID)
	if err != nil {
		writeApplicationError(w, r, "list my applications", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(applications)
}

func (c *JobApplicationController) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "id")
	companyID := r.Header.Get(headerUserID)
	
	if companyID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var payload dto.UpdateJobApplicationStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := c.service.UpdateStatus(r.Context(), appID, companyID, payload.Status); err != nil {
		if err.Error() == "unauthorized: you do not own the offer for this application" {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		writeApplicationError(w, r, "update application status", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"status updated"}`))
}

func (c *JobApplicationController) Delete(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "id")
	role := r.Header.Get(headerUserRole)
	userID := r.Header.Get(headerUserID)

	if err := c.service.DeleteApplication(r.Context(), appID, role, userID); err != nil {
		if err.Error() == "unauthorized: you do not have permission to delete this application" {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		writeApplicationError(w, r, "delete application", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// writeApplicationError responde 404 si la postulación (o su oferta) no
// existe y, ante cualquier otro error, 500 con un mensaje genérico: el
// detalle queda solo en el log para no exponer información interna
// (alerta 90022 de ZAP, "Application Error Disclosure").
func writeApplicationError(w http.ResponseWriter, r *http.Request, op string, err error) {
	switch {
	case errors.Is(err, repositories.ErrApplicationNotFound):
		http.Error(w, "application not found", http.StatusNotFound)
	case errors.Is(err, repositories.ErrOfferNotFound):
		http.Error(w, "offer not found", http.StatusNotFound)
	default:
		slog.ErrorContext(r.Context(), "job application request failed",
			slog.String("operation", op), slog.String("error", err.Error()))
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

