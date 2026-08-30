package middlewares_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"auth/internal/middlewares"
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
