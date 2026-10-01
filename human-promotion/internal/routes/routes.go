package routes

import (
	"github.com/go-chi/chi/v5"

	"buenvivir/human-promotion/internal/controllers"
	"buenvivir/human-promotion/internal/middlewares"
	"buenvivir/human-promotion/internal/services"
)

// Controllers agrupa los controladores HTTP del servicio.
type Controllers struct {
	Ficha        *controllers.FichaController
	Hito         *controllers.HitoController
	Participante *controllers.ParticipanteController
}

// RegisterRoutes registra los endpoints de Promoción Humana.
//
//   - Fichas, hitos, búsqueda y PDF son confidenciales: solo role_admin y
//     role_volunteer (RequireRole deja pasar también a role_dev, igual
//     que en el resto de los servicios).
//   - /mi-linea-de-tiempo es del participante y devuelve solo sus hitos
//     visibles, filtrados en el backend (HU-34).
//
// Un rol no autorizado recibe 403 y el intento queda registrado en el log.
func RegisterRoutes(r chi.Router, jwtService services.JWTService, ctrls Controllers) {
	r.Get("/health", controllers.HealthCheck)

	r.Route("/api/v1/human-promotion", func(r chi.Router) {
		r.Use(middlewares.AuthMiddleware(jwtService))

		r.Group(func(r chi.Router) {
			r.Use(middlewares.RequireRole("role_admin", "role_volunteer"))

			r.Get("/participantes", ctrls.Participante.Buscar)

			r.Post("/fichas", ctrls.Ficha.Create)
			r.Get("/fichas", ctrls.Ficha.GetByParticipante)
			r.Get("/fichas/{id}", ctrls.Ficha.Get)
			r.Put("/fichas/{id}", ctrls.Ficha.Update)
			r.Get("/fichas/{id}/pdf", ctrls.Ficha.ExportPDF)

			r.Post("/fichas/{id}/hitos", ctrls.Hito.Create)
			r.Get("/fichas/{id}/hitos", ctrls.Hito.List)
			r.Put("/fichas/{id}/hitos/{hitoId}", ctrls.Hito.Update)
		})

		r.Group(func(r chi.Router) {
			r.Use(middlewares.RequireRole("role_participant"))

			r.Get("/mi-linea-de-tiempo", ctrls.Participante.MiLineaDeTiempo)
		})
	})
}
