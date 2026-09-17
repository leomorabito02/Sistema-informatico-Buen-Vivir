package controllers

import (
	"encoding/json"
	"net/http"

	"buenvivir/human-promotion/internal/services"
)

type PromotionController struct {
	service services.PromotionService
}

func NewPromotionController(service services.PromotionService) *PromotionController {
	return &PromotionController{service: service}
}

func (c *PromotionController) GetPrograms(w http.ResponseWriter, r *http.Request) {
	programs, err := c.service.GetPrograms(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(programs)
}

func (c *PromotionController) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "UP", "service": "human-promotion-service"})
}
