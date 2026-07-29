package routes

import (
	"net/http"

	"buenvivir/dashboard/internal/controllers"
)

func RegisterRoutes(mux *http.ServeMux, ctrl *controllers.DashboardController) {
	mux.HandleFunc("/health", ctrl.HealthCheck)
	mux.HandleFunc("/api/v1/dashboard/metrics", ctrl.GetMetrics)
}
