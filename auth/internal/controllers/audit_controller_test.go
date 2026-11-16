package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"auth/internal/models"
	"auth/internal/repositories"
)

type fakeAuditRepo struct{ last repositories.AuditFilter }

func (f *fakeAuditRepo) Create(context.Context, *models.AuditLog) error { return nil }
func (f *fakeAuditRepo) List(_ context.Context, fl repositories.AuditFilter) ([]models.AuditLog, error) {
	f.last = fl
	return []models.AuditLog{{ID: 1, UsuarioID: "usr-1", Accion: "baja", Recurso: "perfil_participante", RecursoID: "part-1", Resultado: "ok", Fecha: time.Now()}}, nil
}
func (f *fakeAuditRepo) UserNames(context.Context, []string) (map[string]string, error) {
	return map[string]string{"usr-1": "Ana Paz"}, nil
}

func TestAuditController_List(t *testing.T) {
	repo := &fakeAuditRepo{}
	rec := httptest.NewRecorder()
	NewAuditController(repo).List(rec, httptest.NewRequest(http.MethodGet, "/api/v1/auth/auditoria?desde=2026-11-01&hasta=2026-11-15&accion=baja", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	var out []RegistroAuditoria
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0].UsuarioNombre != "Ana Paz" || out[0].Accion != "baja" {
		t.Errorf("respuesta inesperada: %+v", out)
	}
	if repo.last.Accion != "baja" || repo.last.Hasta.Sub(repo.last.Desde) != 15*24*time.Hour {
		t.Errorf("filtro inesperado: %+v", repo.last)
	}
}

func TestAuditController_List_FechaInvalida(t *testing.T) {
	rec := httptest.NewRecorder()
	NewAuditController(&fakeAuditRepo{}).List(rec, httptest.NewRequest(http.MethodGet, "/x?desde=15/11/2026", nil))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("esperaba 400, fue %d", rec.Code)
	}
}
