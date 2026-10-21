package routes

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"

	"buenvivir/human-promotion/internal/controllers"
	"buenvivir/human-promotion/internal/models"
	"buenvivir/human-promotion/internal/services"
)

const secretoPrueba = "secreto-de-prueba-de-al-menos-32-caracteres"

type indicadoresFalsos struct{ err error }

func (f indicadoresFalsos) Indicadores(context.Context) (models.Indicadores, error) {
	return models.Indicadores{
		Fichas: []models.IndicadorFicha{{ID: "f1", ParticipanteID: "p1", Estado: "activa", CreadaEn: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)}},
		Hitos:  []models.IndicadorHito{{ParticipanteID: "p1", Fecha: "2026-09-10"}},
	}, f.err
}

func token(t *testing.T, rol string) string {
	t.Helper()
	s, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "u-1", "role": rol, "type": "access", "exp": time.Now().Add(time.Hour).Unix(),
	}).SignedString([]byte(secretoPrueba))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func servidor(t *testing.T, repo indicadoresFalsos) *chi.Mux {
	t.Helper()
	t.Setenv("JWT_SECRET", secretoPrueba)
	r := chi.NewRouter()
	RegisterRoutes(r, services.NewJWTService(), Controllers{Indicadores: controllers.NewIndicadoresController(repo)})
	return r
}

func TestIndicadores_SoloAdministracion(t *testing.T) {
	r := servidor(t, indicadoresFalsos{})
	for rol, esperado := range map[string]int{
		"role_admin":       http.StatusOK,
		"role_dev":         http.StatusOK,
		"role_volunteer":   http.StatusForbidden,
		"role_participant": http.StatusForbidden,
		"role_teacher":     http.StatusForbidden,
	} {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/human-promotion/indicadores", nil)
		req.Header.Set("Authorization", "Bearer "+token(t, rol))
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != esperado {
			t.Errorf("%s: esperaba %d, obtuve %d", rol, esperado, rec.Code)
		}
		if esperado == http.StatusOK && !strings.Contains(rec.Body.String(), `"participante_id":"p1"`) {
			t.Errorf("%s: cuerpo inesperado %s", rol, rec.Body.String())
		}
	}

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/human-promotion/indicadores", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("sin token: esperaba 401, obtuve %d", rec.Code)
	}
}

func TestIndicadores_ErrorInterno(t *testing.T) {
	r := servidor(t, indicadoresFalsos{err: errors.New("firestore caído")})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/human-promotion/indicadores", nil)
	req.Header.Set("Authorization", "Bearer "+token(t, "role_admin"))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError || strings.Contains(rec.Body.String(), "firestore") {
		t.Fatalf("esperaba 500 sin detalles internos, obtuve %d %s", rec.Code, rec.Body.String())
	}
}
