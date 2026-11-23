package controllers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"buenvivir/education/internal/services"
)

// AttendanceController expone la asistencia por actividad (HU-19) a admin y
// voluntario. Es de solo lectura: la asistencia la generan las entregas.
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
