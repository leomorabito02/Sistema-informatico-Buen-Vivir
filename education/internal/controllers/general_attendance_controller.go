package controllers

import (
	"net/http"

	"buenvivir/education/internal/middlewares"
	"buenvivir/education/internal/models"
	"buenvivir/education/internal/services"
)

// GeneralAttendanceController expone la asistencia general del centro
// (HU-19bis/RF-64): control diario de concurrencia de todos los
// participantes, independiente de cursos/actividades. Exclusivo de
// role_admin y role_volunteer.
type GeneralAttendanceController struct {
	service services.GeneralAttendanceService
}

func NewGeneralAttendanceController(service services.GeneralAttendanceService) *GeneralAttendanceController {
	return &GeneralAttendanceController{service: service}
}

// List responde GET /api/asistencia-general?fecha=YYYY-MM-DD.
// Devuelve el listado completo de participantes del centro (resuelto
// contra auth-service) junto con su asistencia de esa fecha, si ya se
// cargó.
func (c *GeneralAttendanceController) List(w http.ResponseWriter, r *http.Request) {
	fecha := r.URL.Query().Get("fecha")
	token, _ := r.Context().Value(middlewares.TokenKey).(string)

	views, err := c.service.ListByDate(r.Context(), fecha, token)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, views)
}

// Register responde POST /api/asistencia-general?fecha=YYYY-MM-DD con
// body [{participanteId, presente, observacion}, ...]. Volver a mandar
// el mismo participanteId para la misma fecha actualiza el registro
// existente (upsert), no lo duplica.
func (c *GeneralAttendanceController) Register(w http.ResponseWriter, r *http.Request) {
	fecha := r.URL.Query().Get("fecha")
	token, _ := r.Context().Value(middlewares.TokenKey).(string)

	var inputs []models.GeneralAttendanceInput
	if err := decodeJSONBody(r, &inputs); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "body inválido: "+err.Error())
		return
	}

	if err := c.service.RegisterBatch(r.Context(), fecha, inputs, token); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"fecha": fecha, "status": "asistencia general registrada"})
}
