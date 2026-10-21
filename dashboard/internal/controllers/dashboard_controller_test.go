package controllers_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"buenvivir/dashboard/internal/clients"
	"buenvivir/dashboard/internal/controllers"
	"buenvivir/dashboard/internal/models"
	"buenvivir/dashboard/internal/routes"
)

type fakeService struct {
	err  error
	cred clients.Credentials
}

func (f *fakeService) Participantes(_ context.Context, c clients.Credentials) ([]models.Participante, error) {
	f.cred = c
	return []models.Participante{{ID: "u1", Estado: "activo"}}, f.err
}

func (f *fakeService) Cursos(context.Context, clients.Credentials) ([]models.Curso, error) {
	return []models.Curso{{ID: "c1"}}, f.err
}

func (f *fakeService) Habilidades(context.Context, clients.Credentials) ([]models.ParticipanteHabilidad, error) {
	return []models.ParticipanteHabilidad{{ID: "h1"}}, f.err
}

func (f *fakeService) Postulaciones(context.Context, clients.Credentials) ([]models.Postulacion, error) {
	return []models.Postulacion{{ID: "p1"}}, f.err
}

func (f *fakeService) Areas(context.Context, clients.Credentials) ([]models.Area, error) {
	return []models.Area{{ID: "a1"}}, f.err
}

func (f *fakeService) Promocion(context.Context, clients.Credentials) (models.Promocion, error) {
	return models.Promocion{Fichas: []models.FichaPromocion{{ID: "f1"}}}, f.err
}

func newMux(svc *fakeService) *http.ServeMux {
	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, routes.Rutas{Dashboard: controllers.NewDashboardController(svc)})
	return mux
}

func get(mux http.Handler, path, role string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if role != "" {
		req.Header.Set("X-User-Role", role)
		req.Header.Set("X-User-ID", "usr-1")
		req.Header.Set("Cookie", "access_token=jwt")
	}
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	return w
}

func TestRutas_AdminVeCadaFuente(t *testing.T) {
	svc := &fakeService{}
	mux := newMux(svc)
	for path, id := range map[string]string{
		"/api/v1/dashboard/participantes": `"id":"u1"`,
		"/api/v1/dashboard/cursos":        `"id":"c1"`,
		"/api/v1/dashboard/habilidades":   `"id":"h1"`,
		"/api/v1/dashboard/postulaciones": `"id":"p1"`,
		"/api/v1/dashboard/areas":         `"id":"a1"`,
	} {
		for _, role := range []string{"role_admin", "role_dev"} {
			w := get(mux, path, role)
			if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), id) {
				t.Errorf("%s (%s): %d %s", path, role, w.Code, w.Body.String())
			}
		}
	}
	if svc.cred.Cookie != "access_token=jwt" || svc.cred.UserID != "usr-1" {
		t.Fatalf("pasa las credenciales del pedido: %+v", svc.cred)
	}
}

func TestRutas_SoloAdmin(t *testing.T) {
	mux := newMux(&fakeService{})
	if w := get(mux, "/api/v1/dashboard/cursos", ""); w.Code != http.StatusUnauthorized {
		t.Fatalf("sin sesión: %d", w.Code)
	}
	for _, role := range []string{"role_volunteer", "role_participant", "role_teacher", "role_company"} {
		if w := get(mux, "/api/v1/dashboard/cursos", role); w.Code != http.StatusForbidden {
			t.Fatalf("%s: %d", role, w.Code)
		}
	}
}

func TestRutas_ErroresDeOrigen(t *testing.T) {
	svc := &fakeService{err: &clients.UpstreamError{Service: "educación", Status: http.StatusForbidden}}
	if w := get(newMux(svc), "/api/v1/dashboard/cursos", "role_admin"); w.Code != http.StatusForbidden {
		t.Fatalf("un 403 del origen se respeta: %d", w.Code)
	}
	svc.err = &clients.UpstreamError{Service: "educación", Status: http.StatusInternalServerError}
	w := get(newMux(svc), "/api/v1/dashboard/cursos", "role_admin")
	if w.Code != http.StatusBadGateway || !strings.Contains(w.Body.String(), "no respondió") {
		t.Fatalf("otro error del origen es 502: %d %s", w.Code, w.Body.String())
	}
	svc.err = errors.New("cualquier otra cosa")
	if w := get(newMux(svc), "/api/v1/dashboard/areas", "role_admin"); w.Code != http.StatusBadGateway {
		t.Fatalf("error genérico: %d", w.Code)
	}
}

func TestHealth_Y_Metodos(t *testing.T) {
	mux := newMux(&fakeService{})
	if w := get(mux, "/health", ""); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "UP") {
		t.Fatalf("health: %d", w.Code)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/dashboard/cursos", nil)
	req.Header.Set("X-User-Role", "role_admin")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("solo lectura: %d", w.Code)
	}
}

func TestRegisterRoutes_EventosSoloSiHayHandler(t *testing.T) {
	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, routes.Rutas{Dashboard: controllers.NewDashboardController(&fakeService{}), Eventos: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})})
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/eventos", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("POST /eventos: %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	newMux(&fakeService{}).ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/eventos", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("sin configuración no se expone: %d", rec.Code)
	}
}
