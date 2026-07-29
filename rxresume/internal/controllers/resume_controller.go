package controllers

import (
	"encoding/json"
	"net/http"

	"buenvivir/rxresume/internal/services"
)

type ResumeController struct {
	service services.ResumeService
}

func NewResumeController(service services.ResumeService) *ResumeController {
	return &ResumeController{service: service}
}

func (c *ResumeController) GetResume(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		userID = "usr-1"
	}

	resume, err := c.service.GetResume(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resume)
}

func (c *ResumeController) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "UP", "service": "rxresume-service"})
}
