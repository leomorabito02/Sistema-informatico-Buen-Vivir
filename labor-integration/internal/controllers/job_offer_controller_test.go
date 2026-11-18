package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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
