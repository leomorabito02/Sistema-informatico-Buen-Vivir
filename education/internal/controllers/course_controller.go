package controllers

import (
	"net/http"

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

// Get responde GET /api/cursos/{id}
func (c *CourseController) Get(w http.ResponseWriter, r *http.Request) {
	course, err := c.service.GetCourse(r.Context(), r.PathValue("id"))
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

	course, err := c.service.UpdateCourse(r.Context(), r.PathValue("id"), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, course)
}

// Delete responde DELETE /api/cursos/{id}
func (c *CourseController) Delete(w http.ResponseWriter, r *http.Request) {
	if err := c.service.DeleteCourse(r.Context(), r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Publish responde POST /api/cursos/{id}/publicar
func (c *CourseController) Publish(w http.ResponseWriter, r *http.Request) {
	if err := c.service.PublishCourse(r.Context(), r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"id": r.PathValue("id"), "status": "publicado"})
}

// Unpublish responde POST /api/cursos/{id}/despublicar
func (c *CourseController) Unpublish(w http.ResponseWriter, r *http.Request) {
	if err := c.service.UnpublishCourse(r.Context(), r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"id": r.PathValue("id"), "status": "despublicado"})
}
