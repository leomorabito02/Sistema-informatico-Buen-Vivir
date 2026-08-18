package controllers

import (
	"net/http"

	"buenvivir/education/internal/models"
	"buenvivir/education/internal/services"
)

type TrainingAreaController struct {
	service services.TrainingAreaService
}

func NewTrainingAreaController(service services.TrainingAreaService) *TrainingAreaController {
	return &TrainingAreaController{service: service}
}

// List responde GET /api/areas-formativas
func (c *TrainingAreaController) List(w http.ResponseWriter, r *http.Request) {
	areas, err := c.service.ListTrainingAreas(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, areas)
}

// Create responde POST /api/areas-formativas
func (c *TrainingAreaController) Create(w http.ResponseWriter, r *http.Request) {
	var input models.TrainingAreaInput
	if err := decodeJSONBody(r, &input); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "body inválido: "+err.Error())
		return
	}

	area, err := c.service.CreateTrainingArea(r.Context(), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, area)
}

// Update responde PUT /api/areas-formativas/{id}
func (c *TrainingAreaController) Update(w http.ResponseWriter, r *http.Request) {
	var input models.TrainingAreaInput
	if err := decodeJSONBody(r, &input); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "body inválido: "+err.Error())
		return
	}

	area, err := c.service.UpdateTrainingArea(r.Context(), r.PathValue("id"), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, area)
}

// Delete responde DELETE /api/areas-formativas/{id}
func (c *TrainingAreaController) Delete(w http.ResponseWriter, r *http.Request) {
	if err := c.service.DeleteTrainingArea(r.Context(), r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
