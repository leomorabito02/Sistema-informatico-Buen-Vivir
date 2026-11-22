package controllers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"buenvivir/education/internal/middlewares"
	"buenvivir/education/internal/models"
	"buenvivir/education/internal/services"
)

// DocenteController agrupa los endpoints bajo /api/docente, exclusivos
// del rol role_teacher interno (is_external=false, ver
// middlewares.RequireInternalTeacher). Compone varios services porque
// casi todas las operaciones necesitan primero verificar que el curso
// (o, en los endpoints de asistencia, la actividad a través de su
// curso) pertenece al docente autenticado antes de delegar.
type DocenteController struct {
	courseService     services.CourseService
	activityService   services.ActivityService
	enrollmentService services.EnrollmentService
	attendanceService services.AttendanceService
	progressService   services.ProgressService
	submissionService services.SubmissionService
}

func NewDocenteController(
	courseService services.CourseService,
	activityService services.ActivityService,
	enrollmentService services.EnrollmentService,
	attendanceService services.AttendanceService,
	progressService services.ProgressService,
	submissionService services.SubmissionService,
) *DocenteController {
	return &DocenteController{
		courseService:     courseService,
		activityService:   activityService,
		enrollmentService: enrollmentService,
		attendanceService: attendanceService,
		progressService:   progressService,
		submissionService: submissionService,
	}
}

// teacherID lee el "sub" del JWT, puesto en el contexto por
// AuthMiddleware.
func teacherIDFromRequest(r *http.Request) string {
	id, _ := r.Context().Value(middlewares.UserIDKey).(string)
	return id
}

// ListMyCourses responde GET /api/docente/cursos
func (c *DocenteController) ListMyCourses(w http.ResponseWriter, r *http.Request) {
	courses, err := c.courseService.ListCoursesByTeacher(r.Context(), teacherIDFromRequest(r))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, courses)
}

// ListMyDraftCourses responde GET /api/docente/cursos/borradores:
// cursos del docente que todavía no publicó un role_admin (leídos de la
// CMA). Las referencias vienen solo con su id.
func (c *DocenteController) ListMyDraftCourses(w http.ResponseWriter, r *http.Request) {
	courses, err := c.courseService.ListDraftCoursesByTeacher(r.Context(), teacherIDFromRequest(r))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, courses)
}

// CreateMyCourse responde POST /api/docente/cursos. Crea el curso en
// borrador a nombre del docente autenticado (el responsibleTeacher del
// body se ignora); queda pendiente de que un role_admin lo publique.
func (c *DocenteController) CreateMyCourse(w http.ResponseWriter, r *http.Request) {
	var input models.CourseInput
	if err := decodeJSONBody(r, &input); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "body inválido: "+err.Error())
		return
	}

	course, err := c.courseService.CreateCourseAsTeacher(r.Context(), teacherIDFromRequest(r), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, course)
}

// UpdateMyCourse responde PUT /api/docente/cursos/{id}
func (c *DocenteController) UpdateMyCourse(w http.ResponseWriter, r *http.Request) {
	var input models.CourseInput
	if err := decodeJSONBody(r, &input); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "body inválido: "+err.Error())
		return
	}

	course, err := c.courseService.UpdateCourseAsTeacher(r.Context(), chi.URLParam(r, "id"), teacherIDFromRequest(r), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, course)
}

// DeleteMyCourse responde DELETE /api/docente/cursos/{id}: el docente da
// de baja un curso suyo que sigue en borrador (pendiente de aprobación).
func (c *DocenteController) DeleteMyCourse(w http.ResponseWriter, r *http.Request) {
	if err := c.courseService.DeleteCourseAsTeacher(r.Context(), chi.URLParam(r, "id"), teacherIDFromRequest(r)); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ListActivities responde GET /api/docente/cursos/{cursoId}/actividades
func (c *DocenteController) ListActivities(w http.ResponseWriter, r *http.Request) {
	cursoID := chi.URLParam(r, "cursoId")
	if err := c.courseService.VerifyOwnership(r.Context(), cursoID, teacherIDFromRequest(r)); err != nil {
		writeError(w, err)
		return
	}

	activities, err := c.activityService.ListByCourse(r.Context(), cursoID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, activities)
}

// CreateActivity responde POST /api/docente/cursos/{cursoId}/actividades
func (c *DocenteController) CreateActivity(w http.ResponseWriter, r *http.Request) {
	cursoID := chi.URLParam(r, "cursoId")
	if err := c.courseService.VerifyOwnership(r.Context(), cursoID, teacherIDFromRequest(r)); err != nil {
		writeError(w, err)
		return
	}

	var input models.ActivityInput
	if err := decodeJSONBody(r, &input); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "body inválido: "+err.Error())
		return
	}

	activity, err := c.activityService.CreateActivity(r.Context(), cursoID, input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, activity)
}

// verifyActivityOwnership resuelve el curso de una actividad y valida
// que pertenezca al docente autenticado. Los endpoints de actividad,
// entregas y asistencia se direccionan por actividadId, no por cursoId, así que hace falta
// este paso extra antes de poder reusar CourseService.VerifyOwnership.
// Devuelve la actividad ya resuelta (además del error) para que el
// caller no tenga que volver a leerla de Firestore solo para sacarle
// el CursoID.
func (c *DocenteController) verifyActivityOwnership(r *http.Request, actividadID string) (*models.Activity, error) {
	activity, err := c.activityService.GetActivity(r.Context(), actividadID)
	if err != nil {
		return nil, err
	}
	if err := c.courseService.VerifyOwnership(r.Context(), activity.CursoID, teacherIDFromRequest(r)); err != nil {
		return nil, err
	}
	return activity, nil
}

// UpdateActivity responde PUT /api/docente/actividades/{actividadId}
func (c *DocenteController) UpdateActivity(w http.ResponseWriter, r *http.Request) {
	actividadID := chi.URLParam(r, "actividadId")
	if _, err := c.verifyActivityOwnership(r, actividadID); err != nil {
		writeError(w, err)
		return
	}

	var input models.ActivityInput
	if err := decodeJSONBody(r, &input); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "body inválido: "+err.Error())
		return
	}

	activity, err := c.activityService.UpdateActivity(r.Context(), actividadID, input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, activity)
}

// DeleteActivity responde DELETE /api/docente/actividades/{actividadId}
func (c *DocenteController) DeleteActivity(w http.ResponseWriter, r *http.Request) {
	actividadID := chi.URLParam(r, "actividadId")
	if _, err := c.verifyActivityOwnership(r, actividadID); err != nil {
		writeError(w, err)
		return
	}

	if err := c.activityService.DeleteActivity(r.Context(), actividadID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ListSubmissions responde GET /api/docente/actividades/{actividadId}/entregas
func (c *DocenteController) ListSubmissions(w http.ResponseWriter, r *http.Request) {
	activity, err := c.verifyActivityOwnership(r, chi.URLParam(r, "actividadId"))
	if err != nil {
		writeError(w, err)
		return
	}

	token, _ := r.Context().Value(middlewares.TokenKey).(string)
	rows, err := c.submissionService.ListByActivity(r.Context(), activity, token)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rows)
}

// ListAttendance responde GET /api/docente/actividades/{actividadId}/asistencia
func (c *DocenteController) ListAttendance(w http.ResponseWriter, r *http.Request) {
	actividadID := chi.URLParam(r, "actividadId")
	if _, err := c.verifyActivityOwnership(r, actividadID); err != nil {
		writeError(w, err)
		return
	}

	attendance, err := c.attendanceService.ListByActivity(r.Context(), actividadID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, attendance)
}

// ListEnrollments responde GET /api/docente/cursos/{cursoId}/inscriptos
func (c *DocenteController) ListEnrollments(w http.ResponseWriter, r *http.Request) {
	cursoID := chi.URLParam(r, "cursoId")
	if err := c.courseService.VerifyOwnership(r.Context(), cursoID, teacherIDFromRequest(r)); err != nil {
		writeError(w, err)
		return
	}

	// Con nombre y apellido: el docente veía solo el id del participante.
	token, _ := r.Context().Value(middlewares.TokenKey).(string)
	enrollments, err := c.enrollmentService.ListByCourseConNombres(r.Context(), cursoID, token)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, enrollments)
}

// GetProgress responde GET /api/docente/cursos/{cursoId}/progreso
func (c *DocenteController) GetProgress(w http.ResponseWriter, r *http.Request) {
	cursoID := chi.URLParam(r, "cursoId")
	if err := c.courseService.VerifyOwnership(r.Context(), cursoID, teacherIDFromRequest(r)); err != nil {
		writeError(w, err)
		return
	}

	progress, err := c.progressService.CalculateCourseProgress(r.Context(), cursoID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, progress)
}
