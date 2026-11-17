package routes

import (
	"net/http"

	"buenvivir/notifications/internal/controllers"
	"buenvivir/notifications/internal/middlewares"
)

func RegisterRoutes(mux *http.ServeMux, ctrl *controllers.NotificationController) {
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
}

// AsyncEndpoints groups Pub/Sub push subscription and background flush handlers.
type AsyncEndpoints struct {
	Educacion        http.Handler
	PromocionHumana  http.Handler
	Usuarios         http.Handler
	InsercionLaboral http.Handler
	Campanas         http.Handler
	DeliveryFlush    http.Handler
}

// RegisterAsyncRoutes binds unauthenticated/OIDC-verified push and flush endpoints.
func RegisterAsyncRoutes(mux *http.ServeMux, a AsyncEndpoints) {
	if a.Educacion != nil {
		mux.Handle("POST /eventos/educacion", a.Educacion)
	}
	if a.PromocionHumana != nil {
		mux.Handle("POST /eventos/promocion-humana", a.PromocionHumana)
	}
	if a.Usuarios != nil {
		mux.Handle("POST /eventos/usuarios", a.Usuarios)
	}
	if a.InsercionLaboral != nil {
		mux.Handle("POST /eventos/insercion-laboral", a.InsercionLaboral)
	}
	if a.Campanas != nil {
		mux.Handle("POST /eventos/campanas", a.Campanas)
	}
	if a.DeliveryFlush != nil {
		mux.Handle("POST /internal/delivery/flush", a.DeliveryFlush)
		mux.Handle("POST /internal/outbox/flush", a.DeliveryFlush)
	}
}

// RegisterContactRoute registra el formulario público de contacto (sin
// sesión: lo usan visitantes de la landing y empresas).
func RegisterContactRoute(mux *http.ServeMux, ctrl *controllers.ContactController) {
	if ctrl != nil {
		mux.HandleFunc("POST /api/v1/notifications/contacto", ctrl.Enviar)
	}
}
