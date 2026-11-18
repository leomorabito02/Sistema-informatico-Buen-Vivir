package controllers

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"

	"buenvivir/education/internal/models"
)

// CourseLister devuelve cursos (publicados o en borrador) para buscar en
// cuáles está un material.
type CourseLister func(ctx context.Context) ([]models.Course, error)

// CursoQueUsaMaterial es una fila de GET /api/materiales/{id}/cursos.
type CursoQueUsaMaterial struct {
	ID     string `json:"id"`
	Titulo string `json:"titulo"`
}

// WithCourseListers activa el aviso de HU-14: antes de borrar un material
// se muestra en qué cursos está.
func (c *MaterialController) WithCourseListers(listers ...CourseLister) *MaterialController {
	c.courseListers = listers
	return c
}

// CursosQueUsan responde GET /api/materiales/{id}/cursos (solo admin):
// los cursos que tienen ese material adjunto, para avisar antes de
// borrarlo (HU-14). Si se borra igual, el material desaparece de esos
// cursos.
func (c *MaterialController) CursosQueUsan(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	out := []CursoQueUsaMaterial{}
	seen := map[string]bool{}
	for _, list := range c.courseListers {
		courses, err := list(r.Context())
		if err != nil {
			writeError(w, err)
			return
		}
		for _, course := range courses {
			if seen[course.ID] || !courseUsesMaterial(course, id) {
				continue
			}
			seen[course.ID] = true
			out = append(out, CursoQueUsaMaterial{ID: course.ID, Titulo: course.Title})
		}
	}
	writeJSON(w, http.StatusOK, out)
}

func courseUsesMaterial(course models.Course, materialID string) bool {
	for _, m := range course.Material {
		if m.ID == materialID {
			return true
		}
	}
	return false
}
