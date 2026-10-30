package controllers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"buenvivir/dashboard/internal/auth"
	"buenvivir/dashboard/internal/clients"
	"buenvivir/dashboard/internal/controllers"
	"buenvivir/dashboard/internal/models"
	"buenvivir/dashboard/internal/routes"
)

const secretoPrueba = "secreto-de-prueba-de-al-menos-32-caracteres"

func tokenDe(t *testing.T, rol string) string {
	t.Helper()
	s, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "admin-1", "role": rol, "type": "access", "exp": time.Now().Add(time.Hour).Unix(),
	}).SignedString([]byte(secretoPrueba))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// credService guarda las credenciales con que lo llamaron.
type credService struct {
	fakeService
	visto clients.Credentials
}

func (c *credService) Participantes(_ context.Context, cr clients.Credentials) ([]models.Participante, error) {
	c.visto = cr
	return []models.Participante{}, nil
}

func muxConToken(t *testing.T, svc *credService) *http.ServeMux {
	t.Helper()
	v, err := auth.NewValidator(secretoPrueba)
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, routes.Rutas{Dashboard: controllers.NewDashboardController(svc), Guard: controllers.NewGuard(v)})
	return mux
}

func TestGuardConToken(t *testing.T) {
	svc := &credService{}
	mux := muxConToken(t, svc)
	pedir := func(armar func(*http.Request)) int {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/resumen?desde=2026-01-01&hasta=2026-06-30", nil)
		armar(req)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		return rec.Code
	}

	if c := pedir(func(r *http.Request) {
		r.AddCookie(&http.Cookie{Name: "__Host-access_token", Value: tokenDe(t, "role_admin")})
	}); c != http.StatusOK {
		t.Errorf("admin con cookie: %d", c)
	}
	if svc.visto.UserID != "admin-1" || svc.visto.UserRole != "role_admin" {
		t.Errorf("las credenciales para los servicios de origen tienen que salir del token: %+v", svc.visto)
	}
	if c := pedir(func(r *http.Request) { r.Header.Set("Authorization", "Bearer "+tokenDe(t, "role_dev")) }); c != http.StatusOK {
		t.Errorf("dev con bearer: %d", c)
	}
	if c := pedir(func(r *http.Request) { r.Header.Set("Authorization", "Bearer "+tokenDe(t, "role_volunteer")) }); c != http.StatusForbidden {
		t.Errorf("voluntario: %d", c)
	}
	if c := pedir(func(*http.Request) {}); c != http.StatusUnauthorized {
		t.Errorf("sin token: %d", c)
	}
	if c := pedir(func(r *http.Request) { r.Header.Set("Authorization", "Bearer basura") }); c != http.StatusUnauthorized {
		t.Errorf("token inválido: %d", c)
	}
	// Con validador, un X-User-Role inventado no sirve: manda el token.
	if c := pedir(func(r *http.Request) {
		r.Header.Set("X-User-Role", "role_admin")
		r.Header.Set("Authorization", "Bearer "+tokenDe(t, "role_participant"))
	}); c != http.StatusForbidden {
		t.Errorf("header inventado: esperaba 403, obtuve %d", c)
	}
	if c := pedir(func(r *http.Request) { r.Header.Set("X-User-Role", "role_admin") }); c != http.StatusUnauthorized {
		t.Errorf("solo el header, sin token: esperaba 401, obtuve %d", c)
	}
}

func TestReportesSinConfigurar503(t *testing.T) {
	for _, p := range []string{"/api/v1/dashboard/reportes", "/api/v1/dashboard/reportes/abc/pdf"} {
		w := get(newMux(&fakeService{}), p, "role_admin")
		if w.Code != http.StatusServiceUnavailable || !strings.Contains(w.Body.String(), "no está configurado") {
			t.Errorf("%s: %d %s", p, w.Code, w.Body.String())
		}
	}
}
