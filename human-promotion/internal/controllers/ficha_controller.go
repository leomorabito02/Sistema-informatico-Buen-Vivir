package controllers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"buenvivir/human-promotion/internal/middlewares"
	"buenvivir/human-promotion/internal/models"
	"buenvivir/human-promotion/internal/services"
)

// MsgFichaModificada es el mensaje de éxito que pide HU-30.
const MsgFichaModificada = "Ficha personal modificada correctamente"

type FichaController struct {
	service services.FichaService
	pdf     services.PDFService
}

func NewFichaController(service services.FichaService, pdf services.PDFService) *FichaController {
	return &FichaController{service: service, pdf: pdf}
}

// Create - POST /api/v1/human-promotion/fichas (HU-29)
func (c *FichaController) Create(w http.ResponseWriter, r *http.Request) {
	var input models.CreateFichaInput
	if err := decodeJSONBody(w, r, &input); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "cuerpo de la solicitud inválido")
		return
	}

	ficha, err := c.service.Create(r.Context(), middlewares.UserIDFromContext(r.Context()), input)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, ficha)
}

// Get - GET /api/v1/human-promotion/fichas/{id}
func (c *FichaController) Get(w http.ResponseWriter, r *http.Request) {
	ficha, err := c.service.GetDetalle(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, ficha)
}

// GetByParticipante - GET /api/v1/human-promotion/fichas?participante_id=...
// Devuelve 404 si el participante todavía no tiene ficha: el frontend
// lo usa para ofrecer el alta (HU-29).
func (c *FichaController) GetByParticipante(w http.ResponseWriter, r *http.Request) {
	participanteID := r.URL.Query().Get("participante_id")
	if participanteID == "" {
		writeErrorJSON(w, http.StatusBadRequest, "falta el parámetro participante_id")
		return
	}
	ficha, err := c.service.GetDetalleByParticipante(r.Context(), participanteID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, ficha)
}

// Update - PUT /api/v1/human-promotion/fichas/{id} (HU-30)
func (c *FichaController) Update(w http.ResponseWriter, r *http.Request) {
	var input models.UpdateFichaInput
	if err := decodeJSONBody(w, r, &input); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "cuerpo de la solicitud inválido")
		return
	}

	ficha, err := c.service.Update(r.Context(), chi.URLParam(r, "id"), input)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"message": MsgFichaModificada,
		"ficha":   ficha,
	})
}

// ExportPDF - GET /api/v1/human-promotion/fichas/{id}/pdf (HU-37)
func (c *FichaController) ExportPDF(w http.ResponseWriter, r *http.Request) {
	fichaID := chi.URLParam(r, "id")
	content, err := c.pdf.FichaPDF(r.Context(), fichaID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	// fichaID ya pasó la validación de formato (solo [A-Za-z0-9_-]).
	w.Header().Set("Content-Disposition", `attachment; filename="ficha-`+fichaID+`.pdf"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(content)
}
