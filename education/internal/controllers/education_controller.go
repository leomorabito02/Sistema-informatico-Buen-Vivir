package controllers

import (
	"encoding/json"
	"net/http"

	"buenvivir/education/internal/services"
)

type EducationController struct {
	service services.EducationService
}

func NewEducationController(service services.EducationService) *EducationController {
	return &EducationController{service: service}
}

func (c *EducationController) GetCourses(w http.ResponseWriter, r *http.Request) {
	courses, err := c.service.GetCourses(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(courses)
}

func (c *EducationController) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "UP", "service": "education-service"})
}
