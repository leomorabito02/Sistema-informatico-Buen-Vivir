package routes

import (
	"net/http"

	"buenvivir/education/internal/controllers"
)

// withAuth es el punto de enganche del futuro middleware de
// JWT + RBAC (roles "administrador" y "voluntario" únicamente).
//
// TODO(auth): hoy es un pass-through (no valida nada) porque depende
// de cómo termine resuelto el módulo de Autenticación (auth/), que ya
// tiene JWTService + AuthMiddleware + RequireRole en
// auth/internal/middlewares. Cuando esté listo (probablemente
// reutilizando ese JWTService via un cliente HTTP a auth, o
// validando el JWT localmente si comparten el secreto/JWKS), envolver
// acá así:
//
//	func withAuth(next http.HandlerFunc) http.HandlerFunc {
//	    return middlewares.AuthMiddleware(jwtService)(
//	        middlewares.RequireRole("administrador", "voluntario")(next),
//	    ).ServeHTTP
//	}
//
// Todos los endpoints de escritura (POST/PUT/DELETE y
// publicar/despublicar) ya están marcados abajo con withAuth() para
// que ese cambio sea de una sola línea por endpoint.
func withAuth(next http.HandlerFunc) http.HandlerFunc {
	return next
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
