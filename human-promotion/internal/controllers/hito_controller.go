package controllers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"buenvivir/human-promotion/internal/middlewares"
	"buenvivir/human-promotion/internal/models"
	"buenvivir/human-promotion/internal/services"
)

type HitoController struct {
	service services.HitoService
}

func NewHitoController(service services.HitoService) *HitoController {
	return &HitoController{service: service}
}

// Create - POST /api/v1/human-promotion/fichas/{id}/hitos (HU-32)
func (c *HitoController) Create(w http.ResponseWriter, r *http.Request) {
	var input models.HitoInput
	if err := decodeJSONBody(w, r, &input); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "cuerpo de la solicitud inválido")
		return
	}

	hito, err := c.service.Create(r.Context(), middlewares.UserIDFromContext(r.Context()), chi.URLParam(r, "id"), input)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, hito)
}

// List - GET /api/v1/human-promotion/fichas/{id}/hitos (HU-33)
// Línea de tiempo profesional: todos los hitos, en orden cronológico.
func (c *HitoController) List(w http.ResponseWriter, r *http.Request) {
	hitos, err := c.service.ListDetalle(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, hitos)
}

// Update - PUT /api/v1/human-promotion/fichas/{id}/hitos/{hitoId} (HU-32)
func (c *HitoController) Update(w http.ResponseWriter, r *http.Request) {
	var input models.HitoInput
	if err := decodeJSONBody(w, r, &input); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "cuerpo de la solicitud inválido")
		return
	}

	hito, err := c.service.Update(r.Context(), chi.URLParam(r, "id"), chi.URLParam(r, "hitoId"), input)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, hito)
}
