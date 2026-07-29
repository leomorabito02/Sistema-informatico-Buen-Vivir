package routes

import (
	"net/http"

	"buenvivir/campaigns/internal/controllers"
)

func RegisterRoutes(mux *http.ServeMux, ctrl *controllers.CampaignController) {
	mux.HandleFunc("/health", ctrl.HealthCheck)
	mux.HandleFunc("/api/v1/campaigns", ctrl.GetCampaigns)
}
