package controllers

import (
	"net/http"

	"buenvivir/human-promotion/internal/services"
)

// AuditoriaController expone el log de auditoría en solo lectura (HU-35:
// "consultable solo por administradores").
type AuditoriaController struct {
	service services.AuditoriaService
}

func NewAuditoriaController(service services.AuditoriaService) *AuditoriaController {
	return &AuditoriaController{service: service}
}

// List - GET /api/v1/human-promotion/auditoria?desde=&hasta=&accion=&recurso=
func (c *AuditoriaController) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	out, err := c.service.List(r.Context(), services.FiltroAuditoria{
		Desde:   q.Get("desde"),
		Hasta:   q.Get("hasta"),
		Accion:  q.Get("accion"),
		Recurso: q.Get("recurso"),
	})
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}
