package controllers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"buenvivir/education/internal/middlewares"
	"buenvivir/education/internal/services"
)

// EnrollmentController expone la inscripción de participantes a cursos
// (HU-18/RF-18) para role_admin y role_volunteer. A diferencia de
// DocenteController.ListEnrollments (bajo /api/docente, con chequeo de
// "¿es tu curso?"), estos endpoints no están restringidos a un docente
// responsable puntual: admin y voluntario pueden inscribir en
// cualquier curso publicado, según la vista de HU-06.
type EnrollmentController struct {
	enrollmentService services.EnrollmentService
}

func NewEnrollmentController(enrollmentService services.EnrollmentService) *EnrollmentController {
	return &EnrollmentController{enrollmentService: enrollmentService}
}

type enrollRequest struct {
	ParticipanteID string `json:"participanteId"`
}

// Enroll responde POST /api/cursos/{id}/inscripciones
func (c *EnrollmentController) Enroll(w http.ResponseWriter, r *http.Request) {
	cursoID := chi.URLParam(r, "id")
	token, _ := r.Context().Value(middlewares.TokenKey).(string)

	var body enrollRequest
	if err := decodeJSONBody(r, &body); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "body inválido: "+err.Error())
		return
	}

	enrollment, err := c.enrollmentService.Enroll(r.Context(), body.ParticipanteID, cursoID, token)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, enrollment)
}

// List responde GET /api/cursos/{id}/inscripciones, con nombre y
// apellido de cada participante (resueltos contra auth-service).
func (c *EnrollmentController) List(w http.ResponseWriter, r *http.Request) {
	cursoID := chi.URLParam(r, "id")
	token, _ := r.Context().Value(middlewares.TokenKey).(string)

	enrollments, err := c.enrollmentService.ListByCourseConNombres(r.Context(), cursoID, token)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, enrollments)
}

// CountActive responde GET /api/cursos/{id}/inscriptos-activos (solo
// admin): {"inscriptosActivos": n}, el aviso antes de dar de baja un
// curso (HU-16).
func (c *EnrollmentController) CountActive(w http.ResponseWriter, r *http.Request) {
	n, err := c.enrollmentService.CountActive(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"inscriptosActivos": n})
}

// ListByParticipant responde GET /api/participants/{id}/inscripciones
func (c *EnrollmentController) ListByParticipant(w http.ResponseWriter, r *http.Request) {
	participantID := chi.URLParam(r, "id")

	enrollments, err := c.enrollmentService.ListByParticipant(r.Context(), participantID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, enrollments)
}

type getBulkUserCoursesRequest struct {
	UserIDs []string `json:"user_ids"`
}

// GetBulkUserCourses responde POST /api/bulk-cursos-usuarios
func (c *EnrollmentController) GetBulkUserCourses(w http.ResponseWriter, r *http.Request) {
	var body getBulkUserCoursesRequest
	if err := decodeJSONBody(r, &body); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "body inválido: "+err.Error())
		return
	}

	summaries, err := c.enrollmentService.GetBulkUserCourses(r.Context(), body.UserIDs)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, summaries)
}
