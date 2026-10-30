package controllers

import (
	"net/http"

	"buenvivir/human-promotion/internal/middlewares"
	"buenvivir/human-promotion/internal/services"
)

type ParticipanteController struct {
	service services.ParticipanteService
}

func NewParticipanteController(service services.ParticipanteService) *ParticipanteController {
	return &ParticipanteController{service: service}
}

// Buscar - GET /api/v1/human-promotion/participantes?q=... (HU-36)
// Una lista vacía significa "sin resultados"; el mensaje lo muestra el frontend.
func (c *ParticipanteController) Buscar(w http.ResponseWriter, r *http.Request) {
	resultados, err := c.service.Buscar(r.Context(), r.URL.Query().Get("q"))
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, resultados)
}

// MiLineaDeTiempo - GET /api/v1/human-promotion/mi-linea-de-tiempo (HU-34)
// El participante se identifica solo por su token: no recibe ids de
// ficha ni de participante, así no puede pedir la línea de otra persona.
// Responde {"hitos": [...], "actividad": [...]}.
func (c *ParticipanteController) MiLineaDeTiempo(w http.ResponseWriter, r *http.Request) {
	linea, err := c.service.LineaDeTiempoPropia(r.Context(), middlewares.UserIDFromContext(r.Context()))
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, linea)
}
