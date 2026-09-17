package controllers

import (
	"encoding/json"
	"net/http"

	"buenvivir/labor-integration/internal/services"
)

type JobController struct {
	service services.JobService
}

func NewJobController(service services.JobService) *JobController {
	return &JobController{service: service}
}

func (c *JobController) GetJobs(w http.ResponseWriter, r *http.Request) {
	jobs, err := c.service.GetJobs(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(jobs)
}

func (c *JobController) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "UP", "service": "labor-integration-service"})
}
