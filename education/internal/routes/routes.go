package routes

import (
	"net/http"

	"buenvivir/education/internal/controllers"
	"buenvivir/education/internal/middlewares"
)

// RequireRole verifica que el rol en el contexto sea uno de los permitidos.
// Para esto, el middleware AuthHeaders ya tuvo que haber leído X-User-Role.
func withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roleID, ok := r.Context().Value(middlewares.RoleIDKey).(string)
		if !ok || roleID == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Implementamos un RequireRole simple para "administrador" y "voluntario"
		// Como auth-service usa prefijos, puede ser "role_admin", "role_volunteer"
		// o el dev "role_dev".
		if roleID != "role_admin" && roleID != "role_volunteer" && roleID != "role_dev" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	}
}

// RegisterRoutes registra los endpoints HTTP del servicio de
// Educación. Los de lectura (GET) exponen contenido publicado de
// Contentful (Content Delivery API) y son públicos. Los de escritura
// hablan con la Content Management API y, en el futuro, van a requerir
// autenticación (ver withAuth arriba).
func RegisterRoutes(
	mux *http.ServeMux,
	courseCtrl *controllers.CourseController,
	trainingAreaCtrl *controllers.TrainingAreaController,
	materialCtrl *controllers.MaterialController,
	healthCtrl *controllers.HealthController,
) {
	mux.HandleFunc("GET /api/salud", healthCtrl.Health)

	// -- course --
	mux.HandleFunc("GET /api/cursos", courseCtrl.List)
	mux.HandleFunc("GET /api/cursos/{id}", courseCtrl.Get)
	mux.HandleFunc("POST /api/cursos", withAuth(courseCtrl.Create))
	mux.HandleFunc("PUT /api/cursos/{id}", withAuth(courseCtrl.Update))
	mux.HandleFunc("DELETE /api/cursos/{id}", withAuth(courseCtrl.Delete))
	mux.HandleFunc("POST /api/cursos/{id}/publicar", withAuth(courseCtrl.Publish))
	mux.HandleFunc("POST /api/cursos/{id}/despublicar", withAuth(courseCtrl.Unpublish))

	// -- trainingArea --
	mux.HandleFunc("GET /api/areas-formativas", trainingAreaCtrl.List)
	mux.HandleFunc("POST /api/areas-formativas", withAuth(trainingAreaCtrl.Create))
	mux.HandleFunc("PUT /api/areas-formativas/{id}", withAuth(trainingAreaCtrl.Update))
	mux.HandleFunc("DELETE /api/areas-formativas/{id}", withAuth(trainingAreaCtrl.Delete))

	// -- material --
	mux.HandleFunc("GET /api/materiales", materialCtrl.List)
	mux.HandleFunc("POST /api/materiales", withAuth(materialCtrl.Create))
	mux.HandleFunc("PUT /api/materiales/{id}", withAuth(materialCtrl.Update))
	mux.HandleFunc("DELETE /api/materiales/{id}", withAuth(materialCtrl.Delete))
}
