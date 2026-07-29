package routes

import (
	"net/http"

	"buenvivir/notifications/internal/controllers"
)

func RegisterRoutes(mux *http.ServeMux, ctrl *controllers.NotificationController) {
	mux.HandleFunc("/health", ctrl.HealthCheck)
	mux.HandleFunc("/api/v1/notifications", ctrl.GetNotifications)
}
