package controllers

import (
	"net/http"

	"buenvivir/human-promotion/internal/repositories"
)

type IndicadoresController struct {
	repo repositories.IndicadoresRepository
}

func NewIndicadoresController(repo repositories.IndicadoresRepository) *IndicadoresController {
	return &IndicadoresController{repo: repo}
}

// Get - GET /api/v1/human-promotion/indicadores (solo administración).
// Lo consume el Dashboard para "Fichas activas" y "Participantes con
// seguimiento". Solo ids, estados y fechas (ver models.Indicadores).
func (c *IndicadoresController) Get(w http.ResponseWriter, r *http.Request) {
	ind, err := c.repo.Indicadores(r.Context())
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, ind)
}
