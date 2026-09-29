package controllers

import (
	"encoding/json"
	"net/http"

	"buenvivir/labor-integration/internal/models"
	"buenvivir/labor-integration/internal/services"
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

	if err := c.service.ApplyToJob(r.Context(), &application); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(application)
}
