package routes

import (
	"net/http"

	"buenvivir/dashboard/internal/controllers"
)

// RegisterRoutes: una ruta por fuente de datos del dashboard (HU-52 a
// HU-55). Todas son de lectura y solo para la administración.
func RegisterRoutes(mux *http.ServeMux, ctrl *controllers.DashboardController, events http.Handler) {
	mux.HandleFunc("GET /health", ctrl.HealthCheck)

	// Suscripciones push de Pub/Sub (usuarios, educación e inserción
	// laboral): invalidan la caché de indicadores. Se autentican con el
	// token OIDC de Google, no con el rol del gateway. nil = no se expone.
	if events != nil {
		mux.Handle("POST /eventos", events)
	}

	admin := func(h http.HandlerFunc) http.Handler { return controllers.RequireAdmin(h) }
	mux.Handle("GET /api/v1/dashboard/participantes", admin(ctrl.Participantes))
	mux.Handle("GET /api/v1/dashboard/cursos", admin(ctrl.Cursos))
	mux.Handle("GET /api/v1/dashboard/habilidades", admin(ctrl.Habilidades))
	mux.Handle("GET /api/v1/dashboard/postulaciones", admin(ctrl.Postulaciones))
	mux.Handle("GET /api/v1/dashboard/areas", admin(ctrl.Areas))
}
