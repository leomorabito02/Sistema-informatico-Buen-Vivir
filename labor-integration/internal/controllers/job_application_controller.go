package controllers

import (
	"encoding/json"
	"net/http"

	"buenvivir/labor-integration/internal/models"
	"buenvivir/labor-integration/internal/services"

	"github.com/go-chi/chi/v5"
)

type JobApplicationController struct {
	service *services.JobApplicationService
}

func NewJobApplicationController(service *services.JobApplicationService) *JobApplicationController {
	return &JobApplicationController{service: service}
}

func (c *JobApplicationController) Apply(w http.ResponseWriter, r *http.Request) {
	var application models.JobApplication
	if err := json.NewDecoder(r.Body).Decode(&application); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Seguridad: Sobrescribir el ParticipantID con el ID autenticado desde el Gateway
	application.ParticipantID = r.Header.Get("X-User-ID")

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
	applications, err := c.service.ListByOffer(r.Context(), offerID)
	if err != nil {
		http.Error(w, "failed to list applications", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(applications)
}

func (c *JobApplicationController) ListMine(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	applications, err := c.service.ListByParticipant(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(applications)
}

func (c *JobApplicationController) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "id")
	companyID := r.Header.Get("X-User-ID")
	
	if companyID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var payload struct {
		Estado string `json:"estado"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := c.service.UpdateStatus(r.Context(), appID, companyID, payload.Estado); err != nil {
		if err.Error() == "unauthorized: you do not own the offer for this application" {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"status updated"}`))
}

