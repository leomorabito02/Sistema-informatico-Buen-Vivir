package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"buenvivir/rxresume/internal/models"
	"buenvivir/rxresume/internal/repositories"
)

type memAudit struct{ filas []models.AuditEntry }

func (m *memAudit) Record(_ context.Context, e models.AuditEntry) error {
	m.filas = append(m.filas, e)
	return nil
}

func (m *memAudit) List(_ context.Context, f repositories.AuditFilter) ([]models.AuditEntry, error) {
	out := []models.AuditEntry{}
	for _, e := range m.filas {
		if f.Accion == "" || e.Accion == f.Accion {
			out = append(out, e)
		}
	}
	return out, nil
}

// D8 (22/11/2026): la acción de gestión queda registrada con quién, qué y
// sobre qué CV; sin repositorio no se registra nada ni se rompe.
func TestRegistrar_GuardaQuienQueYSobreQuien(t *testing.T) {
	audit := &memAudit{}
	c := NewResumeController(nil).WithAudit(audit)
	r := httptest.NewRequest(http.MethodGet, "/api/v1/rxresume/user?user_id=usr-part", nil)
	r.Header.Set(headerXUserID, "usr-admin")
	r.Header.Set(headerXUserRole, "role_admin")
	r.RemoteAddr = "10.0.0.5:1234"
	c.registrar(r, AccionVer, "usr-part")
	if len(audit.filas) != 1 {
		t.Fatalf("filas = %d", len(audit.filas))
	}
	f := audit.filas[0]
	if f.UsuarioID != "usr-admin" || f.Rol != "role_admin" || f.Accion != AccionVer || f.Recurso != "cv" || f.RecursoID != "usr-part" || f.IP != "10.0.0.5" {
		t.Errorf("fila inesperada: %+v", f)
	}
	NewResumeController(nil).registrar(r, AccionVer, "usr-part") // sin repo: no hace nada
}

func TestAuditoria_SoloAdministracion(t *testing.T) {
	audit := &memAudit{filas: []models.AuditEntry{{UsuarioID: "a", Accion: AccionEditar, Recurso: "cv", RecursoID: "p"}}}
	c := NewResumeController(nil).WithAudit(audit)
	casos := map[string]int{"role_admin": 200, "role_dev": 200, "role_volunteer": 403, "role_participant": 403, "": 403}
	for rol, want := range casos {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/api/v1/rxresume/auditoria", nil)
		r.Header.Set(headerXUserRole, rol)
		c.Auditoria(w, r)
		if w.Code != want {
			t.Errorf("%q: status %d, want %d", rol, w.Code, want)
		}
		if want == 200 {
			var filas []models.AuditEntry
			if err := json.Unmarshal(w.Body.Bytes(), &filas); err != nil || len(filas) != 1 {
				t.Errorf("respuesta inesperada: %s", w.Body.String())
			}
		}
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/rxresume/auditoria?desde=mal", nil)
	r.Header.Set(headerXUserRole, "role_admin")
	c.Auditoria(w, r)
	if w.Code != http.StatusBadRequest {
		t.Errorf("fecha mal formada: %d", w.Code)
	}
}
