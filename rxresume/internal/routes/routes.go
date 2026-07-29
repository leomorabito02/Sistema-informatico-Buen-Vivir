package routes

import (
	"net/http"

	"buenvivir/rxresume/internal/controllers"
)

func RegisterRoutes(mux *http.ServeMux, ctrl *controllers.ResumeController) {
	mux.HandleFunc("/health", ctrl.HealthCheck)
	mux.HandleFunc("/api/v1/rxresume", ctrl.GetResume)
}
