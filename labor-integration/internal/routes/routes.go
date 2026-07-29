package routes

import (
	"net/http"

	"buenvivir/labor-integration/internal/controllers"
)

func RegisterRoutes(mux *http.ServeMux, ctrl *controllers.JobController) {
	mux.HandleFunc("/health", ctrl.HealthCheck)
	mux.HandleFunc("/api/v1/labor-integration/jobs", ctrl.GetJobs)
}
