package routes

import (
	"github.com/go-chi/chi/v5"

	"buenvivir/campaigns/internal/controllers"
	"buenvivir/campaigns/internal/middlewares"
)

// RegisterRoutes registra las rutas de Campañas (HU-47 a HU-51).
//
//   - Públicas, sin token: el listado de campañas publicadas, con búsqueda
//     por título (HU-50, HU-51).
//   - Gestión: crear, editar, publicar, dar de baja y consultar, para administración y
//     voluntariado (HU-06, HU-47 a HU-49). role_admin y role_dev pasan
//     por el bypass global de RequireRole.
func RegisterRoutes(r chi.Router, jwtService middlewares.JWTService, ctrl *controllers.CampaignController) {
	r.Get("/health", ctrl.HealthCheck)

	r.Route("/api/v1/campaigns", func(r chi.Router) {
		r.Get("/publicas", ctrl.ListPublished)

		r.Group(func(r chi.Router) {
			r.Use(middlewares.AuthMiddleware(jwtService))
			r.Use(middlewares.RequireRole("role_volunteer"))

			r.Get("/", ctrl.List)
			// Auditoría: solo administración. Ruta estática: chi la resuelve antes que /{id}.
			r.With(middlewares.RequireRole("role_admin")).Get("/auditoria", ctrl.Auditoria)
			r.Post("/", ctrl.Create)
			r.Get("/{id}", ctrl.Get)
			r.Put("/{id}", ctrl.Update)
			r.Post("/{id}/publicar", ctrl.Publish)
			// baja lógica
			r.Delete("/{id}", ctrl.Delete)
		})
	})
}
