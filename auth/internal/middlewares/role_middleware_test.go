package middlewares_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"auth/internal/middlewares"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestRequireRole_NoRolesSpecified(t *testing.T) {
	middleware := middlewares.RequireRole()

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireRole_MissingRoleInContext(t *testing.T) {
	middleware := middlewares.RequireRole("role_company")

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireRole_ForbiddenRole(t *testing.T) {
	middleware := middlewares.RequireRole("role_company")

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), middlewares.RoleIDKey, "role_participant")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req.WithContext(ctx))

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireRole_AllowedRole(t *testing.T) {
	middleware := middlewares.RequireRole("role_company")

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), middlewares.RoleIDKey, "role_company")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req.WithContext(ctx))

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireRole_AdminBypass(t *testing.T) {
	middleware := middlewares.RequireRole("role_company")

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), middlewares.RoleIDKey, "role_admin")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req.WithContext(ctx))

	assert.Equal(t, http.StatusOK, w.Code)
}

func reqConUsuario(userID, role, urlUserID string) *http.Request {
	req := httptest.NewRequest("GET", "/participants/by-user/"+urlUserID, nil)
	ctx := context.WithValue(req.Context(), middlewares.UserIDKey, userID)
	ctx = context.WithValue(ctx, middlewares.RoleIDKey, role)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("userId", urlUserID)
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	return req.WithContext(ctx)
}

func TestRequireRoleOrSelf(t *testing.T) {
	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	h := middlewares.RequireRoleOrSelf("userId", "role_volunteer")(ok)

	casos := []struct {
		nombre        string
		user, role, u string
		want          int
	}{
		{"participante pide su propio perfil", "usr-1", "role_participant", "usr-1", http.StatusOK},
		{"participante pide el perfil de otro", "usr-1", "role_participant", "usr-2", http.StatusForbidden},
		{"voluntario pide cualquiera", "usr-9", "role_volunteer", "usr-2", http.StatusOK},
		{"admin por bypass", "usr-8", "role_admin", "usr-2", http.StatusOK},
		{"sin usuario en el contexto", "", "role_participant", "", http.StatusForbidden},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			w := httptest.NewRecorder()
			h.ServeHTTP(w, reqConUsuario(c.user, c.role, c.u))
			assert.Equal(t, c.want, w.Code)
		})
	}
}
