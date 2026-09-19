package controllers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"buenvivir/education/internal/middlewares"
	"buenvivir/education/internal/models"
	"buenvivir/education/internal/services"
)

// AttendanceController expone la asistencia por actividad (HU-19/RF-19)
// a role_admin y role_volunteer, sin el chequeo de ownership que sí
// aplica DocenteController bajo /api/docente. Ahí tiene sentido
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

// Register responde POST /api/actividades/{actividadId}/asistencia
func (c *AttendanceController) Register(w http.ResponseWriter, r *http.Request) {
	actividadID := chi.URLParam(r, "actividadId")
	activity, err := c.activityService.GetActivity(r.Context(), actividadID)
	if err != nil {
		writeError(w, err)
		return
	}

	var inputs []models.AttendanceInput
	if err := decodeJSONBody(r, &inputs); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "body inválido: "+err.Error())
		return
	}

	token, _ := r.Context().Value(middlewares.TokenKey).(string)
	if err := c.attendanceService.RegisterBatch(r.Context(), actividadID, activity.CursoID, inputs, token); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"actividadId": actividadID, "status": "asistencia registrada"})
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
