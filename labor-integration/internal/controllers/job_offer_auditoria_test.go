package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"buenvivir/labor-integration/internal/repositories"
)

type fakeAudit struct {
	entries []repositories.AuditEntry
	filtro  repositories.AuditFilter
}

func (f *fakeAudit) Record(_ context.Context, e repositories.AuditEntry) error {
	f.entries = append(f.entries, e)
	return nil
}

func (f *fakeAudit) List(_ context.Context, fl repositories.AuditFilter) ([]repositories.AuditEntry, error) {
	f.filtro = fl
	return f.entries, nil
}

// HU-42: la auditoría de ofertas es solo para administración.
func TestJobOfferController_Auditoria_SoloAdmin(t *testing.T) {
	audit := &fakeAudit{entries: []repositories.AuditEntry{{ID: "a1", UsuarioID: "usr-admin", Accion: "baja", Recurso: "oferta", RecursoID: "of-1", Fecha: time.Now()}}}
	c := NewJobOfferController(nil).WithAudit(audit)

	for _, role := range []string{"role_volunteer", "role_company", ""} {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/labor-integration/auditoria", nil)
		req.Header.Set(headerUserRole, role)
		w := httptest.NewRecorder()
		c.Auditoria(w, req)
		if w.Code != http.StatusForbidden {
			t.Errorf("rol %q: código %d, quiero 403", role, w.Code)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/labor-integration/auditoria?desde=2026-11-01&hasta=2026-11-15&accion=baja", nil)
	req.Header.Set(headerUserRole, "role_admin")
	w := httptest.NewRecorder()
	c.Auditoria(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("código %d: %s", w.Code, w.Body.String())
	}
	var out []repositories.AuditEntry
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil || len(out) != 1 || out[0].Accion != "baja" {
		t.Fatalf("respuesta inesperada: %s", w.Body.String())
	}
	if audit.filtro.Accion != "baja" || audit.filtro.Hasta.Sub(audit.filtro.Desde) != 15*24*time.Hour {
		t.Errorf("filtro inesperado: %+v", audit.filtro)
	}
}

// record guarda quién, qué y desde qué IP, sin el contenido de la oferta.
func TestJobOfferController_Record(t *testing.T) {
	audit := &fakeAudit{}
	c := NewJobOfferController(nil).WithAudit(audit)
	req := httptest.NewRequest(http.MethodPut, "/x", nil)
	req.RemoteAddr = "10.0.0.7:5555"
	req.Header.Set(headerUserID, "usr-admin")
	req.Header.Set(headerUserRole, "role_admin")
	c.record(req, "aprobar", "of-1")
	if len(audit.entries) != 1 {
		t.Fatal("no registró")
	}
	e := audit.entries[0]
	if e.UsuarioID != "usr-admin" || e.Accion != "aprobar" || e.RecursoID != "of-1" || e.IP != "10.0.0.7" || e.Recurso != "oferta" {
		t.Errorf("registro inesperado: %+v", e)
	}
}
