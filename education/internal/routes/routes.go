package routes

import (
	"github.com/go-chi/chi/v5"

	"buenvivir/education/internal/controllers"
	"buenvivir/education/internal/middlewares"
	"buenvivir/education/internal/services"
)

// Controllers agrupa los controllers del servicio. Los GET de contenido requieren JWT;
// las escrituras, role_admin; /api/docente es del docente interno (con ownership);
// /api/participante opera sobre lo propio; inscripción y asistencia son de admin y
// voluntario (HU-18, HU-19, HU-19bis).
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
	Comment           *controllers.CommentController
	TeacherContent    *controllers.TeacherContentController
	Participante      *controllers.ParticipanteController
	Auditoria         *controllers.AuditoriaController
	Progress          *controllers.ProgressController
}

func RegisterRoutes(
	r chi.Router,
	jwtService services.JWTService,
	ctrls Controllers,
) {
	const apiCursosID = "/api/cursos/{id}"
	r.Get("/api/salud", ctrls.Health.Health)

	// -- comentarios (HU-24): públicos. Comentar tiene el límite estricto por IP (10/min).
	if ctrls.Comment != nil {
		r.Get("/api/comentarios", ctrls.Comment.List)
		r.With(middlewares.StrictRateLimiter()).Post("/api/comentarios", ctrls.Comment.Create)
	}

	// -- biblioteca pública (HU-25): sin login; descargar tiene el límite estricto por IP.
	if ctrls.TeacherContent != nil {
		r.Get("/api/biblioteca", ctrls.TeacherContent.ListPublic)
		r.With(middlewares.StrictRateLimiter()).Post("/api/biblioteca/{id}/descargas", ctrls.TeacherContent.PublicDownload)
	}

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
		// Cada verbo se registra suelto: r.Route haría un Mount de chi que tapa el GET del
		// mismo path registrado arriba para cualquier rol.
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
			// HU-14: en qué cursos está el material, para avisar antes de borrarlo.
			r.Get("/api/materiales/{id}/cursos", ctrls.Material.CursosQueUsan)

			// cuántos inscriptos activos tiene un curso: el aviso antes
			// de darlo de baja (HU-16; el DELETE de arriba es la baja lógica)
			r.Get("/api/cursos/{id}/inscriptos-activos", ctrls.Enrollment.CountActive)

			// registro de auditoría de Educación, solo lectura
			if ctrls.Auditoria != nil {
				r.Get("/api/auditoria", ctrls.Auditoria.List)
			}

			// moderación de comentarios (HU-24)
			if ctrls.Comment != nil {
				r.Delete("/api/comentarios/{id}", ctrls.Comment.Delete)
			}

			// moderación de contenido de docentes externos (HU-23) y
			// quién descargó de la biblioteca (HU-25)
			if ctrls.TeacherContent != nil {
				r.Get("/api/contenidos-docentes", ctrls.TeacherContent.List)
				r.Post("/api/contenidos-docentes/{id}/moderacion", ctrls.TeacherContent.Moderate)
				r.Get("/api/biblioteca/descargas", ctrls.TeacherContent.ListDownloads)
			}
		})

		// -- contenido docente (HU-23): lo sube cualquier docente,
		//    interno o externo (es lo único que ve el externo). Editar
		//    es del dueño mientras está pendiente; borrar y bajar el
		//    archivo los decide el service por rol y dueño.
		if ctrls.TeacherContent != nil {
			r.Group(func(r chi.Router) {
				r.Use(middlewares.RequireRole("role_teacher"))
				r.Post("/api/contenidos-docentes", ctrls.TeacherContent.Upload)
				r.Get("/api/contenidos-docentes/mios", ctrls.TeacherContent.ListMine)
				r.Put("/api/contenidos-docentes/{id}", ctrls.TeacherContent.Update)
			})
			r.Delete("/api/contenidos-docentes/{id}", ctrls.TeacherContent.Delete)
			r.Get("/api/contenidos-docentes/{id}/archivo", ctrls.TeacherContent.File)
		}

		// -- materiales: crear/subir/editar los puede hacer admin o docente interno.
		// Borrar es solo admin.
		r.Group(func(r chi.Router) {
			r.Use(middlewares.RequireRole("role_admin", "role_teacher"))
			r.Use(middlewares.RequireInternalTeacher)

			r.Post("/api/materiales", ctrls.Material.Create)
			r.Post("/api/materiales/upload", ctrls.Material.Upload)
			r.Put("/api/materiales/{id}", ctrls.Material.Update)
		})

		// -- inscripción, actividades, asistencia por actividad (solo lectura) y asistencia
		// general: role_admin o role_volunteer (HU-18, HU-19, HU-19bis). Rutas sueltas para
		// evitar el Mount de chi.
		r.Group(func(r chi.Router) {
			r.Use(middlewares.RequireRole("role_admin", "role_volunteer"))

			r.Post("/api/cursos/{id}/inscripciones", ctrls.Enrollment.Enroll)
			r.Get("/api/cursos/{id}/inscripciones", ctrls.Enrollment.List)
			// progreso formativo de cualquier curso (HU-20), para la
			// pantalla Seguimiento de admin y voluntario
			if ctrls.Progress != nil {
				r.Get("/api/cursos/{id}/progreso", ctrls.Progress.CourseProgress)
			}

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

			// buscador de los selectores de participantes y su historial
			// de asistencia por curso (HU-19)
			if ctrls.Participante != nil {
				r.Get("/api/participantes", ctrls.Participante.Buscar)
				r.Get("/api/participants/{id}/asistencia", ctrls.Participante.HistorialAsistencia)
			}
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
			r.Delete("/cursos/{id}", ctrls.Docente.DeleteMyCourse)
			r.Get("/cursos/{cursoId}/actividades", ctrls.Docente.ListActivities)
			r.Post("/cursos/{cursoId}/actividades", ctrls.Docente.CreateActivity)
			r.Put("/actividades/{actividadId}", ctrls.Docente.UpdateActivity)
			r.Delete("/actividades/{actividadId}", ctrls.Docente.DeleteActivity)
			r.Get("/actividades/{actividadId}/entregas", ctrls.Docente.ListSubmissions)
			r.Get("/actividades/{actividadId}/asistencia", ctrls.Docente.ListAttendance)
			r.Get("/cursos/{cursoId}/inscriptos", ctrls.Docente.ListEnrollments)
			r.Get("/cursos/{cursoId}/progreso", ctrls.Docente.GetProgress)
		})

		// -- participante: sus cursos, actividades y entregas. El participanteId sale del
		// propio token (SubmissionService.ResolveParticipantID), nunca del request.
		r.Route("/api/participante", func(r chi.Router) {
			r.Use(middlewares.RequireRole("role_participant"))

			r.Get("/cursos", ctrls.Submission.ListMyCourses)
			r.Get("/cursos/{id}", ctrls.Submission.GetMyCourse)
			r.Post("/actividades/{actividadId}/entrega", ctrls.Submission.Submit)
			r.Get("/habilidades", ctrls.Skill.MySkills)
		})
	})
}
