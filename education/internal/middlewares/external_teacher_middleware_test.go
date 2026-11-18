package middlewares_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"buenvivir/education/internal/middlewares"
)

func requestWithIsExternal(isExternal bool) *http.Request {
	req := httptest.NewRequest("GET", "/api/docente/cursos", nil)
	return req.WithContext(context.WithValue(req.Context(), middlewares.IsExternalKey, isExternal))
}

func TestRequireInternalTeacher_ExternalTeacher_Forbidden(t *testing.T) {
	w := httptest.NewRecorder()
	handler := middlewares.RequireInternalTeacher(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(w, requestWithIsExternal(true))

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireInternalTeacher_InternalTeacher_Allowed(t *testing.T) {
	w := httptest.NewRecorder()
	handler := middlewares.RequireInternalTeacher(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(w, requestWithIsExternal(false))

	assert.Equal(t, http.StatusOK, w.Code)
}

// Si IsExternalKey no está en el contexto (nunca debería pasar detrás
// de AuthMiddleware, pero por las dudas), el type assertion falla a
// false y no bloquea la request.
func TestRequireInternalTeacher_MissingContext_Allowed(t *testing.T) {
	w := httptest.NewRecorder()
	handler := middlewares.RequireInternalTeacher(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(w, httptest.NewRequest("GET", "/api/docente/cursos", nil))

	assert.Equal(t, http.StatusOK, w.Code)
}
