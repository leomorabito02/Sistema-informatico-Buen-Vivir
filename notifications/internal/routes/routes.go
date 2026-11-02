package routes

import (
	"net/http"

	"buenvivir/notifications/internal/controllers"
	"buenvivir/notifications/internal/middlewares"
)

func RegisterRoutes(mux *http.ServeMux, ctrl *controllers.NotificationController, internalHMACSecret ...string) {
	// Healthcheck endpoint (public)
	mux.HandleFunc("GET /health", ctrl.HealthCheck)

	// Protected notification endpoints
	mux.Handle("GET /api/v1/notifications", middlewares.AuthHeaders(http.HandlerFunc(ctrl.GetNotifications)))
	mux.Handle("GET /api/v1/notifications/unread-count", middlewares.AuthHeaders(http.HandlerFunc(ctrl.GetUnreadCount)))
	mux.Handle("GET /api/v1/notifications/{id}", middlewares.AuthHeaders(http.HandlerFunc(ctrl.GetNotificationDetail)))
	mux.Handle("PATCH /api/v1/notifications/{id}/read", middlewares.AuthHeaders(http.HandlerFunc(ctrl.MarkAsRead)))
	mux.Handle("PATCH /api/v1/notifications/read-all", middlewares.AuthHeaders(http.HandlerFunc(ctrl.MarkAllAsRead)))
	mux.Handle("DELETE /api/v1/notifications/{id}", middlewares.AuthHeaders(http.HandlerFunc(ctrl.DeleteNotification)))

	// User notification preferences
	mux.Handle("GET /api/v1/notifications/preferences", middlewares.AuthHeaders(http.HandlerFunc(ctrl.GetPreferences)))
	mux.Handle("GET /api/v1/notifications/preferences/{category}", middlewares.AuthHeaders(http.HandlerFunc(ctrl.GetPreferenceByCategory)))
	mux.Handle("POST /api/v1/notifications/preferences", middlewares.AuthHeaders(http.HandlerFunc(ctrl.SavePreferences)))
	mux.Handle("POST /api/v1/notifications/preferences/init", middlewares.AuthHeaders(http.HandlerFunc(ctrl.InitDefaultPreferences)))
	mux.Handle("PATCH /api/v1/notifications/preferences", middlewares.AuthHeaders(http.HandlerFunc(ctrl.PatchPreference)))

	// Internal notification ingestion endpoint (protected via HMAC-SHA256 signature)
	secret := ""
	if len(internalHMACSecret) > 0 {
		secret = internalHMACSecret[0]
	}
	mux.Handle("POST /api/v1/notifications/send", middlewares.RequireInternalHMAC(secret)(http.HandlerFunc(ctrl.IngestNotification)))
}
