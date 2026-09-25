package controllers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"buenvivir/education/internal/models"
	"buenvivir/education/internal/services"
)

type CourseController struct {
	service services.CourseService
}

func NewCourseController(service services.CourseService) *CourseController {
	return &CourseController{service: service}
}

// List responde GET /api/cursos
func (c *CourseController) List(w http.ResponseWriter, r *http.Request) {
	courses, err := c.service.ListCourses(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, courses)
}

// ListDrafts responde GET /api/cursos/borradores (solo role_admin):
// todos los cursos en borrador, para que administración no pierda de
// vista un curso creado y todavía no publicado.
func (c *CourseController) ListDrafts(w http.ResponseWriter, r *http.Request) {
	courses, err := c.service.ListAllDraftCourses(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, courses)
}

// ListByTeacherAdmin responde GET /api/docentes/{id}/cursos (solo admin)
func (c *CourseController) ListByTeacherAdmin(w http.ResponseWriter, r *http.Request) {
	teacherID := chi.URLParam(r, "id")
	courses, err := c.service.ListCoursesByTeacher(r.Context(), teacherID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, courses)
}

// Get responde GET /api/cursos/{id}
func (c *CourseController) Get(w http.ResponseWriter, r *http.Request) {
	course, err := c.service.GetCourse(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, course)
}

// Create responde POST /api/cursos
func (c *CourseController) Create(w http.ResponseWriter, r *http.Request) {
	var input models.CourseInput
	if err := decodeJSONBody(r, &input); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "body inválido: "+err.Error())
		return
	}

	course, err := c.service.CreateCourse(r.Context(), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, course)
}

// Update responde PUT /api/cursos/{id}
func (c *CourseController) Update(w http.ResponseWriter, r *http.Request) {
	var input models.CourseInput
	if err := decodeJSONBody(r, &input); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "body inválido: "+err.Error())
		return
	}

	course, err := c.service.UpdateCourse(r.Context(), chi.URLParam(r, "id"), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, course)
}

// Delete responde DELETE /api/cursos/{id}
func (c *CourseController) Delete(w http.ResponseWriter, r *http.Request) {
	if err := c.service.DeleteCourse(r.Context(), chi.URLParam(r, "id")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Publish responde POST /api/cursos/{id}/publicar
func (c *CourseController) Publish(w http.ResponseWriter, r *http.Request) {
	if err := c.service.PublishCourse(r.Context(), chi.URLParam(r, "id")); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"id": chi.URLParam(r, "id"), "status": "publicado"})
}

// Unpublish responde POST /api/cursos/{id}/despublicar
func (c *CourseController) Unpublish(w http.ResponseWriter, r *http.Request) {
	if err := c.service.UnpublishCourse(r.Context(), chi.URLParam(r, "id")); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"id": chi.URLParam(r, "id"), "status": "despublicado"})
}
