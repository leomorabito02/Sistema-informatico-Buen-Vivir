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

// List responde GET /api/cursos/{id}/inscripciones
func (c *EnrollmentController) List(w http.ResponseWriter, r *http.Request) {
	cursoID := chi.URLParam(r, "id")

	enrollments, err := c.enrollmentService.ListByCourse(r.Context(), cursoID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, enrollments)
}
