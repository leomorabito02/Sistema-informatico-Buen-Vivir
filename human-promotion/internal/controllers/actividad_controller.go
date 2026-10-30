package controllers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"buenvivir/human-promotion/internal/services"
)

type ActividadController struct {
	service services.ActividadService
}

func NewActividadController(service services.ActividadService) *ActividadController {
	return &ActividadController{service: service}
}

// List - GET /api/v1/human-promotion/fichas/{id}/actividad (HU-33)
// Actividad del participante en el centro (inscripciones, cursos
// completados y asistencia por mes), para mezclar con los hitos.
func (c *ActividadController) List(w http.ResponseWriter, r *http.Request) {
	actividad, err := c.service.DeFicha(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, actividad)
}
