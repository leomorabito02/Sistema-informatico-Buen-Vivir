package routes

import (
	"github.com/go-chi/chi/v5"

	"buenvivir/education/internal/controllers"
	"buenvivir/education/internal/middlewares"
	"buenvivir/education/internal/services"
)

// RegisterRoutes registra los endpoints HTTP del servicio de
// Educación sobre un router chi. Los GET de curso/área/material
// exponen contenido publicado de Contentful (Content Delivery API) y
// requieren un JWT válido pero ningún rol en particular. Las
// escrituras (POST/PUT/DELETE, publicar/despublicar) requieren
// role_admin. El grupo /api/docente es exclusivo de role_teacher
// interno (is_external=false) y valida ownership sobre cada curso
// antes de operar (ver DocenteController).
func RegisterRoutes(
	r chi.Router,
	jwtService services.JWTService,
	courseCtrl *controllers.CourseController,
	trainingAreaCtrl *controllers.TrainingAreaController,
	materialCtrl *controllers.MaterialController,
	healthCtrl *controllers.HealthController,
	docenteCtrl *controllers.DocenteController,
) {
	r.Get("/api/salud", healthCtrl.Health)

	r.Group(func(r chi.Router) {
		r.Use(middlewares.AuthMiddleware(jwtService))

		// -- lectura: cualquier rol autenticado --
		r.Get("/api/cursos", courseCtrl.List)
		r.Get("/api/cursos/{id}", courseCtrl.Get)
		r.Get("/api/areas-formativas", trainingAreaCtrl.List)
		r.Get("/api/materiales", materialCtrl.List)

		// -- escritura: exclusiva de role_admin --
		//
		// OJO: acá NO se puede usar r.Route(prefix, func(r chi.Router){...})
		// para agrupar estos verbos, porque chi lo resuelve como un Mount
		// (un subrouter aparte montado en ese prefijo para TODOS los
		// métodos), que intercepta y tapa el r.Get(mismo path) ya
		// registrado arriba para cualquier rol autenticado — un GET
		// terminaría cayendo en el subrouter admin-only y devolviendo 403
		// en vez de llegar a courseCtrl.List. Por eso cada verbo se
		// registra suelto, en el mismo router (comparten el mismo árbol
		// de rutas de chi, cada uno atado a su propio método).
		r.Group(func(r chi.Router) {
			r.Use(middlewares.RequireRole("role_admin"))

			// GET estático: chi lo resuelve antes que /api/cursos/{id}
			// (registrado arriba para cualquier rol), así que no lo tapa.
			r.Get("/api/cursos/borradores", courseCtrl.ListDrafts)

			r.Post("/api/cursos", courseCtrl.Create)
			r.Put("/api/cursos/{id}", courseCtrl.Update)
			r.Delete("/api/cursos/{id}", courseCtrl.Delete)
			r.Post("/api/cursos/{id}/publicar", courseCtrl.Publish)
			r.Post("/api/cursos/{id}/despublicar", courseCtrl.Unpublish)

			r.Post("/api/areas-formativas", trainingAreaCtrl.Create)
			r.Put("/api/areas-formativas/{id}", trainingAreaCtrl.Update)
			r.Delete("/api/areas-formativas/{id}", trainingAreaCtrl.Delete)

			// borrar material: solo admin (no hay modelo de ownership de
			// materiales, un material puede estar linkeado a varios cursos).
			r.Delete("/api/materiales/{id}", materialCtrl.Delete)
		})

		// -- materiales: crear/subir/editar los puede hacer admin O
		//    docente interno (para adjuntar/actualizar archivos de sus
		//    propios cursos, ver vista docente de las HU). El borrado
		//    queda arriba, solo admin.
		r.Group(func(r chi.Router) {
			r.Use(middlewares.RequireRole("role_admin", "role_teacher"))
			r.Use(middlewares.RequireInternalTeacher)

			r.Post("/api/materiales", materialCtrl.Create)
			r.Post("/api/materiales/upload", materialCtrl.Upload)
			r.Put("/api/materiales/{id}", materialCtrl.Update)
		})

		// -- docente responsable: role_teacher interno --
		r.Route("/api/docente", func(r chi.Router) {
			r.Use(middlewares.RequireRole("role_teacher"))
			r.Use(middlewares.RequireInternalTeacher)

			r.Get("/cursos", docenteCtrl.ListMyCourses)
			r.Get("/cursos/borradores", docenteCtrl.ListMyDraftCourses)
			r.Post("/cursos", docenteCtrl.CreateMyCourse)
			r.Put("/cursos/{id}", docenteCtrl.UpdateMyCourse)
			r.Get("/cursos/{cursoId}/actividades", docenteCtrl.ListActivities)
			r.Post("/cursos/{cursoId}/actividades", docenteCtrl.CreateActivity)
			r.Post("/actividades/{actividadId}/asistencia", docenteCtrl.RegisterAttendance)
			r.Get("/actividades/{actividadId}/asistencia", docenteCtrl.ListAttendance)
			r.Get("/cursos/{cursoId}/inscriptos", docenteCtrl.ListEnrollments)
			r.Get("/cursos/{cursoId}/progreso", docenteCtrl.GetProgress)
		})
	})
}
