package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"buenvivir/education/internal/models"
)

func usageRouter(c *MaterialController) http.Handler {
	r := chi.NewRouter()
	r.Get("/api/materiales/{id}/cursos", c.CursosQueUsan)
	return r
}

// HU-14: lista publicados y borradores que tienen el material, sin repetir.
func TestCursosQueUsan(t *testing.T) {
	publicados := func(context.Context) ([]models.Course, error) {
		return []models.Course{
			{ID: "c1", Title: "Huerta", Material: []models.Material{{ID: "m1"}}},
			{ID: "c2", Title: "Cocina", Material: []models.Material{{ID: "m2"}}},
		}, nil
	}
	borradores := func(context.Context) ([]models.Course, error) {
		return []models.Course{
			{ID: "c3", Title: "Carpintería", Material: []models.Material{{ID: "m1"}}},
			{ID: "c1", Title: "Huerta", Material: []models.Material{{ID: "m1"}}},
		}, nil
	}
	c := NewMaterialController(nil).WithCourseListers(publicados, borradores)
	w := httptest.NewRecorder()
	usageRouter(c).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/materiales/m1/cursos", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("código %d", w.Code)
	}
	var out []CursoQueUsaMaterial
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	if len(out) != 2 || out[0].ID != "c1" || out[1].ID != "c3" {
		t.Fatalf("respuesta inesperada: %+v", out)
	}
}

func TestCursosQueUsan_ErrorDeContentful(t *testing.T) {
	falla := func(context.Context) ([]models.Course, error) { return nil, errors.New("contentful caído") }
	w := httptest.NewRecorder()
	usageRouter(NewMaterialController(nil).WithCourseListers(falla)).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/materiales/m1/cursos", nil))
	if w.Code < 500 {
		t.Fatalf("esperaba error 5xx, fue %d", w.Code)
	}
}
