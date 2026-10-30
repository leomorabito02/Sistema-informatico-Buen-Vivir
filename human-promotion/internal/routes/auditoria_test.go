package routes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"buenvivir/human-promotion/internal/controllers"
	"buenvivir/human-promotion/internal/models"
	"buenvivir/human-promotion/internal/services"
)

type auditoriaFalsa struct{}

func (auditoriaFalsa) List(_ context.Context, f services.FiltroAuditoria) ([]models.RegistroAuditoria, error) {
	if f.Desde == "mal" {
		return nil, services.NewValidationError("Fecha desde inválida: se espera AAAA-MM-DD")
	}
	return []models.RegistroAuditoria{{ID: "1", UsuarioID: "u-1", Accion: "actualizar", Recurso: f.Recurso}}, nil
}

func TestAuditoria_SoloAdministracion(t *testing.T) {
	t.Setenv("JWT_SECRET", secretoPrueba)
	r := chi.NewRouter()
	RegisterRoutes(r, services.NewJWTService(), Controllers{Auditoria: controllers.NewAuditoriaController(auditoriaFalsa{})})

	for rol, esperado := range map[string]int{
		"role_admin":       http.StatusOK,
		"role_dev":         http.StatusOK,
		"role_volunteer":   http.StatusForbidden,
		"role_participant": http.StatusForbidden,
		"role_teacher":     http.StatusForbidden,
	} {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/human-promotion/auditoria?recurso=ficha", nil)
		req.Header.Set("Authorization", "Bearer "+token(t, rol))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != esperado {
			t.Errorf("%s: esperaba %d, obtuve %d", rol, esperado, rec.Code)
		}
		if esperado == http.StatusOK && !strings.Contains(rec.Body.String(), `"recurso":"ficha"`) {
			t.Errorf("%s: cuerpo inesperado %s", rol, rec.Body.String())
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/human-promotion/auditoria?desde=mal", nil)
	req.Header.Set("Authorization", "Bearer "+token(t, "role_admin"))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("fecha inválida: esperaba 400, obtuve %d", rec.Code)
	}

	// el log es de solo lectura: no hay otros métodos
	for _, m := range []string{http.MethodPost, http.MethodPut, http.MethodDelete} {
		req := httptest.NewRequest(m, "/api/v1/human-promotion/auditoria", nil)
		req.Header.Set("Authorization", "Bearer "+token(t, "role_admin"))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusMethodNotAllowed && rec.Code != http.StatusNotFound {
			t.Errorf("%s: esperaba 405/404, obtuve %d", m, rec.Code)
		}
	}
}
