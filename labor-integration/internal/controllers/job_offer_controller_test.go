package controllers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"buenvivir/labor-integration/internal/models"
	"buenvivir/labor-integration/internal/services"
)

type stubOfferRepo struct {
	offers map[string]*models.JobOffer
}

func (s *stubOfferRepo) Create(context.Context, *models.JobOffer) error { return nil }
func (s *stubOfferRepo) List(context.Context, string) ([]*models.JobOffer, error) {
	return nil, nil
}
func (s *stubOfferRepo) Update(context.Context, string, *models.JobOffer) error { return nil }
func (s *stubOfferRepo) Delete(context.Context, string) error                   { return nil }
func (s *stubOfferRepo) UpdateStatus(context.Context, string, string) error     { return nil }
func (s *stubOfferRepo) UpdateStatusAndApprover(context.Context, string, string, string) error {
	return nil
}
func (s *stubOfferRepo) Close(_ context.Context, id, closedBy string, closedAt time.Time) error {
	if o, ok := s.offers[id]; ok {
		o.Status = "Cerrada"
		o.ClosedBy = closedBy
		o.ClosedAt = &closedAt
	}
	return nil
}
func (s *stubOfferRepo) GetByID(_ context.Context, id string) (*models.JobOffer, error) {
	return s.offers[id], nil
}

// Aprobar y rechazar ofertas es solo del administrador (HU-41). Los
// rechazos se cortan antes de llegar al servicio.
func TestJobOfferController_AprobarYRechazarSoloAdmin(t *testing.T) {
	c := NewJobOfferController(nil)
	for _, role := range []string{"role_volunteer", "role_company", "role_participant", ""} {
		for nombre, h := range map[string]http.HandlerFunc{"approve": c.Approve, "reject": c.Reject} {
			req := httptest.NewRequest(http.MethodPut, "/offers/of-1/"+nombre, nil)
			req.Header.Set(headerUserRole, role)
			w := httptest.NewRecorder()
			h(w, req)
			if w.Code != http.StatusForbidden {
				t.Errorf("%s con rol %q: código %d, quiero 403", nombre, role, w.Code)
			}
		}
	}
}

func TestJobOfferController_List_BusquedaLarga(t *testing.T) {
	c := NewJobOfferController(nil)
	req := httptest.NewRequest(http.MethodGet, "/offers?q="+strings.Repeat("a", 101), nil)
	w := httptest.NewRecorder()
	c.List(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("código %d, quiero 400", w.Code)
	}
}

func TestJobOfferController_Close_AuditoriaYPermisos(t *testing.T) {
	// Arrange
	repo := &stubOfferRepo{offers: map[string]*models.JobOffer{
		"of-1": {ID: "of-1", CompanyID: "usr-empresa", Title: "Panadero", Status: "Activa"},
	}}
	audit := &fakeAudit{}
	svc := services.NewJobOfferService(repo, nil, nil)
	c := NewJobOfferController(svc).WithAudit(audit)

	withOfferID := func(r *http.Request, id string) *http.Request {
		rc := chi.NewRouteContext()
		rc.URLParams.Add("id", id)
		return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rc))
	}

	// Act 1: unauthorized participant
	reqForbidden := withOfferID(httptest.NewRequest(http.MethodPut, "/offers/of-1/close", nil), "of-1")
	reqForbidden.Header.Set(headerUserRole, "role_participant")
	reqForbidden.Header.Set(headerUserID, "usr-p")
	wForbidden := httptest.NewRecorder()
	c.Close(wForbidden, reqForbidden)

	// Assert 1
	if wForbidden.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for participant, got %d", wForbidden.Code)
	}
	if len(audit.entries) != 0 {
		t.Fatalf("expected no audit entry on forbidden close")
	}

	// Act 2: owning company closes active offer
	reqOK := withOfferID(httptest.NewRequest(http.MethodPut, "/offers/of-1/close", nil), "of-1")
	reqOK.Header.Set(headerUserRole, "role_company")
	reqOK.Header.Set(headerUserID, "usr-empresa")
	wOK := httptest.NewRecorder()
	c.Close(wOK, reqOK)

	// Assert 2
	if wOK.Code != http.StatusOK {
		t.Fatalf("expected 200 on owner close, got %d: %s", wOK.Code, wOK.Body.String())
	}
	if repo.offers["of-1"].Status != "Cerrada" {
		t.Errorf("expected offer status Cerrada, got %q", repo.offers["of-1"].Status)
	}
	if len(audit.entries) != 1 || audit.entries[0].Accion != "cerrar" || audit.entries[0].RecursoID != "of-1" {
		t.Errorf("unexpected audit entries: %+v", audit.entries)
	}
}
