package routes

import (
	"net/http"

	"buenvivir/human-promotion/internal/controllers"
)

func RegisterRoutes(mux *http.ServeMux, ctrl *controllers.PromotionController) {
	mux.HandleFunc("/health", ctrl.HealthCheck)
	mux.HandleFunc("/api/v1/human-promotion/programs", ctrl.GetPrograms)
}
