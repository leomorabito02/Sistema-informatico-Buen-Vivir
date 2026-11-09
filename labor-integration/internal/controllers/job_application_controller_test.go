package controllers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"buenvivir/labor-integration/internal/models"
	"buenvivir/labor-integration/internal/repositories"
	"buenvivir/labor-integration/internal/services"
)

// appsRepo es un JobApplicationRepository mínimo: GetByID devuelve err.
type appsRepo struct{ err error }

func (f *appsRepo) Create(context.Context, *models.JobApplication) error { return nil }
func (f *appsRepo) ListByOffer(context.Context, string) ([]*models.JobApplication, error) {
	return nil, nil
}
func (f *appsRepo) HasApplicationsForOffer(context.Context, string) (bool, error) { return false, nil }
func (f *appsRepo) HasApplied(context.Context, string, []string) (bool, error)    { return false, nil }
func (f *appsRepo) ListByParticipant(context.Context, []string) ([]*models.JobApplication, error) {
	return nil, nil
}
func (f *appsRepo) UpdateStatus(context.Context, string, string) error { return nil }
func (f *appsRepo) GetByID(context.Context, string) (*models.JobApplication, error) {
	return nil, f.err
}
func (f *appsRepo) Delete(context.Context, string) error { return nil }

func withID(r *http.Request, id string) *http.Request {
	rc := chi.NewRouteContext()
	rc.URLParams.Add("id", id)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rc))
}

func newAppController(err error) *JobApplicationController {
	return NewJobApplicationController(services.NewJobApplicationService(&appsRepo{err: err}, nil, nil))
}

// ZAP llamaba a /applications/id (postulación inexistente) y recibía 500 con
// el texto del error. Tiene que ser 404.
func TestJobApplicationController_PostulacionInexistenteDa404(t *testing.T) {
	c := newAppController(repositories.ErrApplicationNotFound)

	req := withID(httptest.NewRequest(http.MethodDelete, "/applications/id", nil), "id")
	req.Header.Set(headerUserRole, "role_admin")
	w := httptest.NewRecorder()
	c.Delete(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("Delete: código %d, quiero 404", w.Code)
	}

	req = withID(httptest.NewRequest(http.MethodPut, "/applications/id/status", strings.NewReader(`{"status":"aceptado"}`)), "id")
	req.Header.Set(headerUserID, "empresa-1")
	w = httptest.NewRecorder()
	c.UpdateStatus(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("UpdateStatus: código %d, quiero 404", w.Code)
	}
}

// Un error interno responde 500 sin exponer el detalle.
func TestJobApplicationController_ErrorInternoNoExponeDetalle(t *testing.T) {
	const secreto = "rpc error: code = Unavailable desc = firestore.googleapis.com"
	c := newAppController(errors.New(secreto))

	req := withID(httptest.NewRequest(http.MethodDelete, "/applications/a-1", nil), "a-1")
	req.Header.Set(headerUserRole, "role_admin")
	w := httptest.NewRecorder()
	c.Delete(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("código %d, quiero 500", w.Code)
	}
	if strings.Contains(w.Body.String(), "firestore") {
		t.Errorf("la respuesta expone el error interno: %q", w.Body.String())
	}
}
