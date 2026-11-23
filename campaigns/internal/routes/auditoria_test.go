package routes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/go-chi/chi/v5"

	"buenvivir/campaigns/internal/controllers"
	"buenvivir/campaigns/internal/models"
	"buenvivir/campaigns/internal/repositories"
	"buenvivir/campaigns/internal/services"
)

// memAudit guarda la auditoría en memoria.
type memAudit struct {
	mu    sync.Mutex
	filas []repositories.AuditEntry
}

func (m *memAudit) Record(_ context.Context, e repositories.AuditEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.filas = append(m.filas, e)
	return nil
}

func (m *memAudit) List(_ context.Context, f repositories.AuditFilter) ([]repositories.AuditEntry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := []repositories.AuditEntry{}
	for _, e := range m.filas {
		if f.Accion == "" || e.Accion == f.Accion {
			out = append(out, e)
		}
	}
	return out, nil
}

func servidorConAuditoria(t *testing.T) (*httptest.Server, *memAudit) {
	t.Helper()
	audit := &memAudit{}
	r := chi.NewRouter()
	svc := services.NewCampaignService(&memRepo{datos: map[string]models.Campaign{}})
	RegisterRoutes(r, services.NewJWTServiceWithSecret(secreto), controllers.NewCampaignController(svc).WithAudit(audit))
	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)
	return srv, audit
}

// Crear, editar, publicar y dar de baja quedan registrados con quién, qué y sobre qué campaña.
func TestAuditoria_RegistraLasEscrituras(t *testing.T) {
	srv, audit := servidorConAuditoria(t)
	vol := token(t, "role_volunteer", "voluntario-1")
	var creada struct {
		Campana models.Campaign `json:"campana"`
	}
	cuerpo := cuerpoValido
	if st := pedir(t, srv, http.MethodPost, "/api/v1/campaigns", vol, cuerpo, &creada); st != http.StatusCreated {
		t.Fatalf("crear: %d", st)
	}
	id := creada.Campana.ID
	pedir(t, srv, http.MethodPut, "/api/v1/campaigns/"+id, vol, cuerpo, nil)
	pedir(t, srv, http.MethodPost, "/api/v1/campaigns/"+id+"/publicar", vol, "", nil)
	pedir(t, srv, http.MethodDelete, "/api/v1/campaigns/"+id, vol, "", nil)

	want := []string{controllers.AccionCrear, controllers.AccionEditar, controllers.AccionPublicar, controllers.AccionBaja}
	if len(audit.filas) != len(want) {
		t.Fatalf("filas de auditoría = %d, want %d: %+v", len(audit.filas), len(want), audit.filas)
	}
	for i, f := range audit.filas {
		if f.Accion != want[i] || f.UsuarioID != "voluntario-1" || f.Rol != "role_volunteer" || f.RecursoID != id || f.Recurso != "campana" {
			t.Errorf("fila %d inesperada: %+v", i, f)
		}
	}
}

func TestAuditoria_SoloAdministracionLaLee(t *testing.T) {
	srv, _ := servidorConAuditoria(t)
	casos := map[string]int{
		"role_admin":       http.StatusOK,
		"role_dev":         http.StatusOK,
		"role_volunteer":   http.StatusForbidden,
		"role_participant": http.StatusForbidden,
	}
	for rol, want := range casos {
		var out any
		if want == http.StatusOK {
			out = &[]repositories.AuditEntry{}
		}
		if st := pedir(t, srv, http.MethodGet, "/api/v1/campaigns/auditoria", token(t, rol, "u"), "", out); st != want {
			t.Errorf("%s: status %d, want %d", rol, st, want)
		}
	}
	if st := pedir(t, srv, http.MethodGet, "/api/v1/campaigns/auditoria", "", "", nil); st != http.StatusUnauthorized {
		t.Errorf("sin token: status %d, want 401", st)
	}
	if st := pedir(t, srv, http.MethodGet, "/api/v1/campaigns/auditoria?desde=22-11-2026", token(t, "role_admin", "u"), "", nil); st != http.StatusBadRequest {
		t.Errorf("fecha mal formada: status %d, want 400", st)
	}
}
