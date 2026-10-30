package controllers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"buenvivir/education/internal/middlewares"
	"buenvivir/education/internal/models"
	"buenvivir/education/internal/services"
)

// ActivityController expone el CRUD de actividades de un curso y el
// listado de entregas de cada actividad a role_admin y role_volunteer,
// sin el chequeo de ownership que sí aplica DocenteController bajo
// /api/docente — admin y voluntario operan sobre cualquier curso del
// centro. Comparte los mismos services que DocenteController: una
// actividad creada acá es indistinguible de una creada por el docente
// responsable.
type ActivityController struct {
	courseService     services.CourseService
	activityService   services.ActivityService
	submissionService services.SubmissionService
}

func NewActivityController(courseService services.CourseService, activityService services.ActivityService, submissionService services.SubmissionService) *ActivityController {
	return &ActivityController{courseService: courseService, activityService: activityService, submissionService: submissionService}
}

// List responde GET /api/cursos/{id}/actividades
func (c *ActivityController) List(w http.ResponseWriter, r *http.Request) {
	activities, err := c.activityService.ListByCourse(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, activities)
}

// Create responde POST /api/cursos/{id}/actividades. Valida que el
// curso exista antes de crear la actividad, para no dejar actividades
// huérfanas en Firestore apuntando a un cursoId inexistente. Acepta
// también borradores (el admin puede armar las actividades antes de
// publicar el curso).
func (c *ActivityController) Create(w http.ResponseWriter, r *http.Request) {
	cursoID := chi.URLParam(r, "id")
	if err := c.courseService.VerifyExists(r.Context(), cursoID); err != nil {
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

// Update responde PUT /api/actividades/{actividadId}
func (c *ActivityController) Update(w http.ResponseWriter, r *http.Request) {
	var input models.ActivityInput
	if err := decodeJSONBody(r, &input); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "body inválido: "+err.Error())
		return
	}

	activity, err := c.activityService.UpdateActivity(r.Context(), chi.URLParam(r, "actividadId"), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, activity)
}

// Delete responde DELETE /api/actividades/{actividadId}. Borra también
// las entregas y asistencias de la actividad.
func (c *ActivityController) Delete(w http.ResponseWriter, r *http.Request) {
	if err := c.activityService.DeleteActivity(r.Context(), chi.URLParam(r, "actividadId")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ListSubmissions responde GET /api/actividades/{actividadId}/entregas
func (c *ActivityController) ListSubmissions(w http.ResponseWriter, r *http.Request) {
	activity, err := c.activityService.GetActivity(r.Context(), chi.URLParam(r, "actividadId"))
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
