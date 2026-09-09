package middlewares_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"buenvivir/education/internal/middlewares"
)

func requestWithRole(role string) *http.Request {
	req := httptest.NewRequest("GET", "/api/cursos", nil)
	if role == "" {
		return req
	}
	return req.WithContext(context.WithValue(req.Context(), middlewares.RoleIDKey, role))
}

func TestRequireRole_MissingRoleInContext(t *testing.T) {
	middleware := middlewares.RequireRole("role_admin")
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, requestWithRole(""))

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireRole_InsufficientRole(t *testing.T) {
	middleware := middlewares.RequireRole("role_admin")
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, requestWithRole("role_participant"))

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireRole_AllowedRole(t *testing.T) {
	middleware := middlewares.RequireRole("role_teacher")
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, requestWithRole("role_teacher"))

	assert.Equal(t, http.StatusOK, w.Code)
}

// role_admin y role_dev tienen bypass global, igual que en auth-service.
func TestRequireRole_AdminBypass(t *testing.T) {
	middleware := middlewares.RequireRole("role_teacher")
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, requestWithRole("role_admin"))

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireRole_DevBypass(t *testing.T) {
	middleware := middlewares.RequireRole("role_teacher")
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, requestWithRole("role_dev"))

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireRole_NoRolesConfigured_AllowsAny(t *testing.T) {
	middleware := middlewares.RequireRole()
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, requestWithRole(""))

	assert.Equal(t, http.StatusOK, w.Code)
}
