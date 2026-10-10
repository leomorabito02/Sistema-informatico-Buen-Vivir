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
// antes de operar (ver DocenteController). El grupo /api/participante
// es del rol participante y opera solo sobre sus propias inscripciones
// y entregas. El grupo de inscripción /
// actividades / asistencia por actividad / asistencia general es para
// role_admin y role_volunteer (HU-18, HU-19, HU-19bis): a diferencia de
// /api/docente, no tienen scope de "mis cursos" — admin y voluntario
// operan sobre cualquier curso o participante del centro.
type Controllers struct {
	Course            *controllers.CourseController
	TrainingArea      *controllers.TrainingAreaController
	Material          *controllers.MaterialController
	Health            *controllers.HealthController
	Docente           *controllers.DocenteController
	Enrollment        *controllers.EnrollmentController
	Attendance        *controllers.AttendanceController
	GeneralAttendance *controllers.GeneralAttendanceController
	Activity          *controllers.ActivityController
	Submission        *controllers.SubmissionController
	Skill             *controllers.SkillController
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
		r.Get("/api/cursos/{id}/habilidades", ctrls.Skill.CourseSkills)

		// -- adjuntos de una entrega: cualquier rol autenticado llega,
		//    y SubmissionController.DownloadFile decide por rol (admin/
		//    voluntario: todo; docente: sus cursos; participante: lo suyo).
		r.Get("/api/entregas/{entregaId}/archivos/{indice}", ctrls.Submission.DownloadFile)

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

			// Dashboard: todas las habilidades ganadas. Ruta estática, no
			// choca con GET /api/habilidades (catálogo, más abajo).
			r.Get("/api/habilidades/otorgadas", ctrls.Skill.GrantedSkills)

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

		// -- inscripción, actividades (CRUD + entregas), consulta de
		//    asistencia por actividad y asistencia general: role_admin o
		//    role_volunteer (HU-18, HU-19, HU-19bis). La asistencia por
		//    actividad es solo lectura: la generan las entregas. Rutas
		//    sueltas (no r.Route) por la misma razón que el bloque de
		//    admin de más arriba: evitar el Mount de chi.
		r.Group(func(r chi.Router) {
			r.Use(middlewares.RequireRole("role_admin", "role_volunteer"))

			r.Post("/api/cursos/{id}/inscripciones", ctrls.Enrollment.Enroll)
			r.Get("/api/cursos/{id}/inscripciones", ctrls.Enrollment.List)

			r.Get("/api/cursos/{id}/actividades", ctrls.Activity.List)
			r.Post("/api/cursos/{id}/actividades", ctrls.Activity.Create)
			r.Put("/api/actividades/{actividadId}", ctrls.Activity.Update)
			r.Delete("/api/actividades/{actividadId}", ctrls.Activity.Delete)
			r.Get("/api/actividades/{actividadId}/entregas", ctrls.Activity.ListSubmissions)

			r.Get("/api/actividades/{actividadId}/asistencia", ctrls.Attendance.List)

			r.Get("/api/asistencia-general", ctrls.GeneralAttendance.List)
			r.Post("/api/asistencia-general", ctrls.GeneralAttendance.Register)

			r.Get("/api/participants/{id}/inscripciones", ctrls.Enrollment.ListByParticipant)
			r.Get("/api/participants/{id}/habilidades", ctrls.Skill.ParticipantSkills)
			r.Get("/api/docentes/{id}/cursos", ctrls.Course.ListByTeacherAdmin)

			r.Post("/api/bulk-cursos-usuarios", ctrls.Enrollment.GetBulkUserCourses)
		})

		// -- habilidades de las actividades (HU-21 / HU-58): admin,
		//    voluntario y docente interno. El docente solo opera sobre
		//    sus cursos: SkillController.authorizeCourse lo verifica según
		//    el rol (rutas compartidas, no duplicadas bajo /api/docente).
		r.Group(func(r chi.Router) {
			r.Use(middlewares.RequireRole("role_admin", "role_volunteer", "role_teacher"))
			r.Use(middlewares.RequireInternalTeacher)

			r.Get("/api/habilidades", ctrls.Skill.SearchCatalog)
			r.Get("/api/habilidades/categorias", ctrls.Skill.Categories)
			r.Put("/api/actividades/{actividadId}/habilidades", ctrls.Skill.SetActivitySkills)
			// POST estático: no choca con PUT/DELETE /api/actividades/{actividadId}.
			r.Post("/api/actividades/sugerir-habilidades", ctrls.Skill.Suggest)
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
			r.Put("/actividades/{actividadId}", ctrls.Docente.UpdateActivity)
			r.Delete("/actividades/{actividadId}", ctrls.Docente.DeleteActivity)
			r.Get("/actividades/{actividadId}/entregas", ctrls.Docente.ListSubmissions)
			r.Get("/actividades/{actividadId}/asistencia", ctrls.Docente.ListAttendance)
			r.Get("/cursos/{cursoId}/inscriptos", ctrls.Docente.ListEnrollments)
			r.Get("/cursos/{cursoId}/progreso", ctrls.Docente.GetProgress)
		})

		// -- participante: sus cursos inscriptos, el detalle de cada uno
		//    con sus actividades y la entrega de actividades. El
		//    participanteId sale de auth-service con el propio token (ver
		//    SubmissionService.ResolveParticipantID), nunca del request.
		r.Route("/api/participante", func(r chi.Router) {
			r.Use(middlewares.RequireRole("role_participant"))

			r.Get("/cursos", ctrls.Submission.ListMyCourses)
			r.Get("/cursos/{id}", ctrls.Submission.GetMyCourse)
			r.Post("/actividades/{actividadId}/entrega", ctrls.Submission.Submit)
			r.Get("/habilidades", ctrls.Skill.MySkills)
		})
	})
}
