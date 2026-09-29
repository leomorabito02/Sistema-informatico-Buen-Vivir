package controllers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"buenvivir/education/internal/services"
)

// AttendanceController expone la consulta de asistencia por actividad
// (HU-19/RF-19) a role_admin y role_volunteer. Es de solo lectura: la
// asistencia la generan las entregas de los participantes. No tiene el
// chequeo de ownership que sí aplica DocenteController bajo
// /api/docente. Ahí tiene sentido
// verificar "¿es tu curso?" porque un docente solo opera sobre los
// cursos de los que es responsable; acá no aplica: admin y voluntario
// tienen visibilidad sobre cualquier curso del centro (ver RBAC de
// HU-06). Comparte el mismo ActivityService/AttendanceService que
// DocenteController — no hay lógica de negocio duplicada, solo un
// controller sin el gate de ownership.
type AttendanceController struct {
	activityService   services.ActivityService
	attendanceService services.AttendanceService
}

func NewAttendanceController(activityService services.ActivityService, attendanceService services.AttendanceService) *AttendanceController {
	return &AttendanceController{activityService: activityService, attendanceService: attendanceService}
}

// List responde GET /api/actividades/{actividadId}/asistencia
func (c *AttendanceController) List(w http.ResponseWriter, r *http.Request) {
	actividadID := chi.URLParam(r, "actividadId")
	if _, err := c.activityService.GetActivity(r.Context(), actividadID); err != nil {
		writeError(w, err)
		return
	}

	attendance, err := c.attendanceService.ListByActivity(r.Context(), actividadID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, attendance)
}
