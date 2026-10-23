package routes

import (
	"net/http"

	"buenvivir/dashboard/internal/controllers"
	"buenvivir/dashboard/internal/reportes"
)

// Rutas agrupa lo que expone el servicio.
type Rutas struct {
	Dashboard *controllers.DashboardController
	// Reportes: historial de exportaciones (HU-55). nil = responde 503.
	Reportes *reportes.Handlers
	// Eventos: push de Pub/Sub. nil = no se expone.
	Eventos http.Handler
	// Guard protege las rutas de la administración. nil = RequireAdmin
	// (rol puesto por el gateway de Docker).
	Guard func(http.Handler) http.Handler
}

// RegisterRoutes: el resumen ya calculado (HU-52 a HU-54), las áreas
// formativas para el filtro y el historial de reportes (HU-55). Todo es
// solo para la administración. Los datos crudos de cada servicio ya no se
// exponen: el cálculo se hace acá y el frontend solo muestra.
func RegisterRoutes(mux *http.ServeMux, r Rutas) {
	ctrl := r.Dashboard
	mux.HandleFunc("GET /health", ctrl.HealthCheck)

	// Suscripciones push de Pub/Sub (usuarios, educación e inserción
	// laboral): invalidan la caché de indicadores. Se autentican con el
	// token OIDC de Google, no con el rol del gateway.
	if r.Eventos != nil {
		mux.Handle("POST /eventos", r.Eventos)
	}

	guard := r.Guard
	if guard == nil {
		guard = controllers.RequireAdmin
	}
	admin := func(h http.HandlerFunc) http.Handler { return guard(h) }
	mux.Handle("GET /api/v1/dashboard/resumen", admin(ctrl.Resumen))
	mux.Handle("GET /api/v1/dashboard/areas", admin(ctrl.Areas))

	rep := r.Reportes
	mux.Handle("POST /api/v1/dashboard/reportes", admin(rep.Registrar))
	mux.Handle("GET /api/v1/dashboard/reportes", admin(rep.Listar))
	mux.Handle("GET /api/v1/dashboard/reportes/{id}/pdf", admin(rep.Descargar))
}
