package middlewares

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"auth/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
)

type fakeAuditWriter struct{ entries []models.AuditLog }

func (f *fakeAuditWriter) Create(_ context.Context, e *models.AuditLog) error {
	f.entries = append(f.entries, *e)
	return nil
}

type fakeJWT struct{}

func (fakeJWT) ValidateToken(tok string) (jwt.MapClaims, error) {
	switch tok {
	case "admin":
		return jwt.MapClaims{"sub": "usr-admin", "role": "role_admin"}, nil
	case "part":
		return jwt.MapClaims{"sub": "usr-part", "role": "role_participant"}, nil
	}
	return nil, errors.New("invalid")
}

func auditRouter(w *fakeAuditWriter) http.Handler {
	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(AuditTrail(w))
		r.Use(AuthMiddleware(fakeJWT{}))
		r.With(RequireRole("role_volunteer")).Put("/api/v1/participants/{id}", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
		r.Post("/api/v1/users/", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"success":true,"data":{"user_id":"usr-nuevo"}}`))
		})
		r.Patch("/api/v1/users/{id}/approval", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
		r.Delete("/api/v1/participants/{id}", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusBadRequest) })
		r.Get("/api/v1/auth/me", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	})
	return r
}

func doReq(h http.Handler, method, path, token, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// HU-08 y HU-10: una edición queda registrada con quién, qué recurso y
// los NOMBRES de los campos (nunca los valores ni la contraseña).
func TestAuditTrail_EdicionRegistraCamposSinValores(t *testing.T) {
	w := &fakeAuditWriter{}
	rec := doReq(auditRouter(w), http.MethodPut, "/api/v1/participants/part-9", "admin", `{"dni":"30111222","address":"Calle 1","password":"x"}`)
	if rec.Code != http.StatusOK || len(w.entries) != 1 {
		t.Fatalf("status %d, entries %d", rec.Code, len(w.entries))
	}
	e := w.entries[0]
	if e.UsuarioID != "usr-admin" || e.Accion != AuditActualizar || e.Recurso != "perfil_participante" || e.RecursoID != "part-9" || e.Resultado != "ok" {
		t.Errorf("registro inesperado: %+v", e)
	}
	if e.Campos != "address,dni" || strings.Contains(e.Campos, "30111222") {
		t.Errorf("campos inesperados: %q", e.Campos)
	}
}

// HU-07: el alta toma el id de la respuesta.
func TestAuditTrail_AltaTomaIDDeLaRespuesta(t *testing.T) {
	w := &fakeAuditWriter{}
	doReq(auditRouter(w), http.MethodPost, "/api/v1/users/", "admin", `{"email":"a@gmail.com"}`)
	if len(w.entries) != 1 || w.entries[0].Accion != AuditCrear || w.entries[0].RecursoID != "usr-nuevo" || w.entries[0].Recurso != "usuario" {
		t.Fatalf("registro inesperado: %+v", w.entries)
	}
}

// HU-12: aprobar y rechazar se distinguen.
func TestAuditTrail_AprobarYRechazar(t *testing.T) {
	w := &fakeAuditWriter{}
	h := auditRouter(w)
	doReq(h, http.MethodPatch, "/api/v1/users/usr-1/approval", "admin", `{"approval_status":"approved"}`)
	doReq(h, http.MethodPatch, "/api/v1/users/usr-2/approval", "admin", `{"approval_status":"rejected"}`)
	if len(w.entries) != 2 || w.entries[0].Accion != AuditAprobar || w.entries[1].Accion != AuditRechazar {
		t.Fatalf("registros inesperados: %+v", w.entries)
	}
}

// HU-06: un acceso denegado por rol queda registrado; HU-03: un token
// inválido también, pero un pedido sin sesión no.
func TestAuditTrail_Denegaciones(t *testing.T) {
	w := &fakeAuditWriter{}
	h := auditRouter(w)
	if rec := doReq(h, http.MethodPut, "/api/v1/participants/part-9", "part", `{}`); rec.Code != http.StatusForbidden {
		t.Fatalf("esperaba 403, fue %d", rec.Code)
	}
	doReq(h, http.MethodGet, "/api/v1/auth/me", "vencido", "")
	doReq(h, http.MethodGet, "/api/v1/auth/me", "", "")
	if len(w.entries) != 2 {
		t.Fatalf("esperaba 2 registros, hubo %d: %+v", len(w.entries), w.entries)
	}
	if w.entries[0].Accion != AuditAccesoDenegado || w.entries[0].UsuarioID != "usr-part" || w.entries[0].Resultado != "denegado" {
		t.Errorf("403 mal registrado: %+v", w.entries[0])
	}
	if w.entries[1].Accion != AuditTokenRechazado {
		t.Errorf("401 mal registrado: %+v", w.entries[1])
	}
}

// Lecturas y errores de validación no se registran.
func TestAuditTrail_IgnoraLecturasYErrores(t *testing.T) {
	w := &fakeAuditWriter{}
	h := auditRouter(w)
	doReq(h, http.MethodGet, "/api/v1/auth/me", "admin", "")
	doReq(h, http.MethodDelete, "/api/v1/participants/part-1", "admin", "")
	if len(w.entries) != 0 {
		t.Fatalf("no debería registrar: %+v", w.entries)
	}
}
