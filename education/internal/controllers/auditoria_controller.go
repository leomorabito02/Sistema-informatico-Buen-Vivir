package controllers

import (
	"net/http"

	"buenvivir/education/internal/middlewares"
	"buenvivir/education/internal/services"
)

// AuditoriaController expone el registro de auditoría de Educación en
// solo lectura (únicamente administración).
type AuditoriaController struct {
	service services.AuditoriaService
}

func NewAuditoriaController(service services.AuditoriaService) *AuditoriaController {
	return &AuditoriaController{service: service}
}

// List responde GET /api/auditoria?desde=&hasta=&accion=&recurso=.
func (c *AuditoriaController) List(w http.ResponseWriter, r *http.Request) {
	token, _ := r.Context().Value(middlewares.TokenKey).(string)
	q := r.URL.Query()
	out, err := c.service.List(r.Context(), services.FiltroAuditoria{
		Desde:   q.Get("desde"),
		Hasta:   q.Get("hasta"),
		Accion:  q.Get("accion"),
		Recurso: q.Get("recurso"),
	}, token)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}
