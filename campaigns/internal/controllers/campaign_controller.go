package controllers

import (
	"encoding/json"
	"net/http"

	"buenvivir/campaigns/internal/services"
)

type CampaignController struct {
	service services.CampaignService
}

func NewCampaignController(service services.CampaignService) *CampaignController {
	return &CampaignController{service: service}
}

func (c *CampaignController) GetCampaigns(w http.ResponseWriter, r *http.Request) {
	campaigns, err := c.service.GetCampaigns(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(campaigns)
}

func (c *CampaignController) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "UP", "service": "campaigns-service"})
}
