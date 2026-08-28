package controllers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"buenvivir/education/internal/models"
	"buenvivir/education/internal/services"
)

type MaterialController struct {
	service services.MaterialService
}

func NewMaterialController(service services.MaterialService) *MaterialController {
	return &MaterialController{service: service}
}

// List responde GET /api/materiales
func (c *MaterialController) List(w http.ResponseWriter, r *http.Request) {
	materials, err := c.service.ListMaterials(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, materials)
}

// Create responde POST /api/materiales
func (c *MaterialController) Create(w http.ResponseWriter, r *http.Request) {
	var input models.MaterialInput
	if err := decodeJSONBody(r, &input); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "body inválido: "+err.Error())
		return
	}

	material, err := c.service.CreateMaterial(r.Context(), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, material)
}

// Update responde PUT /api/materiales/{id}
func (c *MaterialController) Update(w http.ResponseWriter, r *http.Request) {
	var input models.MaterialInput
	if err := decodeJSONBody(r, &input); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "body inválido: "+err.Error())
		return
	}

	material, err := c.service.UpdateMaterial(r.Context(), chi.URLParam(r, "id"), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, material)
}

// Delete responde DELETE /api/materiales/{id}
func (c *MaterialController) Delete(w http.ResponseWriter, r *http.Request) {
	if err := c.service.DeleteMaterial(r.Context(), chi.URLParam(r, "id")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
