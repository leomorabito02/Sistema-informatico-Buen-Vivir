package controllers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"buenvivir/education/internal/services"
)

// ProgressController expone el progreso formativo de un curso (HU-20)
// para role_admin y role_volunteer, sin el chequeo de "¿es tu curso?" de
// DocenteController.GetProgress: según HU-06 admin y voluntario operan
// sobre cualquier curso. Agregado en la auditoría 2026-11-11: la pantalla
// "Seguimiento" estaba en el menú de admin y voluntario pero solo podía
// usar /api/docente/* (403 para el voluntario, lista vacía para el admin).
type ProgressController struct {
	courseService   services.CourseService
	progressService services.ProgressService
}

func NewProgressController(courseService services.CourseService, progressService services.ProgressService) *ProgressController {
	return &ProgressController{courseService: courseService, progressService: progressService}
}

// CourseProgress responde GET /api/cursos/{id}/progreso: por cada
// inscripto, actividades del curso, actividades cumplidas y porcentaje.
func (c *ProgressController) CourseProgress(w http.ResponseWriter, r *http.Request) {
	cursoID := chi.URLParam(r, "id")
	if err := c.courseService.VerifyExists(r.Context(), cursoID); err != nil {
		writeError(w, err)
		return
	}
	progress, err := c.progressService.CalculateCourseProgress(r.Context(), cursoID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, progress)
}
