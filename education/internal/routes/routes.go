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
// antes de operar (ver DocenteController). El grupo de inscripción /
// asistencia por actividad / asistencia general es para role_admin y
// role_volunteer (HU-18, HU-19, HU-19bis): a diferencia de
// /api/docente, no tienen scope de "mis cursos" — admin y voluntario
// operan sobre cualquier curso o participante del centro.
type Controllers struct {
	Course           *controllers.CourseController
	TrainingArea     *controllers.TrainingAreaController
	Material         *controllers.MaterialController
	Health           *controllers.HealthController
	Docente          *controllers.DocenteController
	Enrollment       *controllers.EnrollmentController
	Attendance       *controllers.AttendanceController
	GeneralAttendance *controllers.GeneralAttendanceController
}

func RegisterRoutes(
	r chi.Router,
	jwtService services.JWTService,
	ctrls Controllers,
) {
	const apiCursosID = "/api/cursos/{id}"
	r.Get("/api/salud", ctrls.Health.Health)

	r.Group(func(r chi.Router) {
		r.Use(middlewares.AuthMiddleware(jwtService))

		// -- lectura: cualquier rol autenticado --
		r.Get("/api/cursos", ctrls.Course.List)
		r.Get(apiCursosID, ctrls.Course.Get)
		r.Get("/api/areas-formativas", ctrls.TrainingArea.List)
		r.Get("/api/materiales", ctrls.Material.List)

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
			r.Get("/api/cursos/borradores", ctrls.Course.ListDrafts)

			r.Post("/api/cursos", ctrls.Course.Create)
			r.Put(apiCursosID, ctrls.Course.Update)
			r.Delete(apiCursosID, ctrls.Course.Delete)
			r.Post("/api/cursos/{id}/publicar", ctrls.Course.Publish)
			r.Post("/api/cursos/{id}/despublicar", ctrls.Course.Unpublish)

			r.Post("/api/areas-formativas", ctrls.TrainingArea.Create)
			r.Put("/api/areas-formativas/{id}", ctrls.TrainingArea.Update)
			r.Delete("/api/areas-formativas/{id}", ctrls.TrainingArea.Delete)

			// borrar material: solo admin (no hay modelo de ownership de
			// materiales, un material puede estar linkeado a varios cursos).
			r.Delete("/api/materiales/{id}", ctrls.Material.Delete)
		})

		// -- materiales: crear/subir/editar los puede hacer admin O
		//    docente interno (para adjuntar/actualizar archivos de sus
		//    propios cursos, ver vista docente de las HU). El borrado
		//    queda arriba, solo admin.
		r.Group(func(r chi.Router) {
			r.Use(middlewares.RequireRole("role_admin", "role_teacher"))
			r.Use(middlewares.RequireInternalTeacher)

			r.Post("/api/materiales", ctrls.Material.Create)
			r.Post("/api/materiales/upload", ctrls.Material.Upload)
			r.Put("/api/materiales/{id}", ctrls.Material.Update)
		})

		// -- inscripción, asistencia por actividad y asistencia general:
		//    role_admin o role_volunteer (HU-18, HU-19, HU-19bis). Rutas
		//    sueltas (no r.Route) por la misma razón que el bloque de
		//    admin de más arriba: evitar el Mount de chi.
		r.Group(func(r chi.Router) {
			r.Use(middlewares.RequireRole("role_admin", "role_volunteer"))

			r.Post("/api/cursos/{id}/inscripciones", ctrls.Enrollment.Enroll)
			r.Get("/api/cursos/{id}/inscripciones", ctrls.Enrollment.List)

			r.Post("/api/actividades/{actividadId}/asistencia", ctrls.Attendance.Register)
			r.Get("/api/actividades/{actividadId}/asistencia", ctrls.Attendance.List)

			r.Get("/api/asistencia-general", ctrls.GeneralAttendance.List)
			r.Post("/api/asistencia-general", ctrls.GeneralAttendance.Register)

			r.Get("/api/participants/{id}/inscripciones", ctrls.Enrollment.ListByParticipant)
			r.Get("/api/docentes/{id}/cursos", ctrls.Course.ListByTeacherAdmin)
			
			r.Post("/api/bulk-cursos-usuarios", ctrls.Enrollment.GetBulkUserCourses)
		})

		// -- docente responsable: role_teacher interno --
		r.Route("/api/docente", func(r chi.Router) {
			r.Use(middlewares.RequireRole("role_teacher"))
			r.Use(middlewares.RequireInternalTeacher)

			r.Get("/cursos", ctrls.Docente.ListMyCourses)
			r.Get("/cursos/borradores", ctrls.Docente.ListMyDraftCourses)
			r.Post("/cursos", ctrls.Docente.CreateMyCourse)
			r.Put("/cursos/{id}", ctrls.Docente.UpdateMyCourse)
			r.Get("/cursos/{cursoId}/actividades", ctrls.Docente.ListActivities)
			r.Post("/cursos/{cursoId}/actividades", ctrls.Docente.CreateActivity)
			r.Post("/actividades/{actividadId}/asistencia", ctrls.Docente.RegisterAttendance)
			r.Get("/actividades/{actividadId}/asistencia", ctrls.Docente.ListAttendance)
			r.Get("/cursos/{cursoId}/inscriptos", ctrls.Docente.ListEnrollments)
			r.Get("/cursos/{cursoId}/progreso", ctrls.Docente.GetProgress)
		})
	})
}
