package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"buenvivir/education/internal/middlewares"
)

func TestChain_AppliesMiddlewaresInOrder(t *testing.T) {
	var order []string

	mw := func(name string) middlewares.Middleware {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name)
				next.ServeHTTP(w, r)
			})
		}
	}

	final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
		w.WriteHeader(http.StatusOK)
	})

	h := middlewares.Chain(final, mw("primero"), mw("segundo"))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, []string{"primero", "segundo", "handler"}, order)
}

func TestChain_NoMiddlewaresReturnsHandlerUntouched(t *testing.T) {
	final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})

	w := httptest.NewRecorder()
	middlewares.Chain(final).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, http.StatusTeapot, w.Code)
}

func TestAuthHeaders_PopulatesContextFromGatewayHeaders(t *testing.T) {
	var gotUser, gotRole any

	handler := middlewares.AuthHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUser = r.Context().Value(middlewares.UserIDKey)
		gotRole = r.Context().Value(middlewares.RoleIDKey)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/cursos", nil)
	req.Header.Set("X-User-ID", "user-42")
	req.Header.Set("X-User-Role", "role_admin")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "user-42", gotUser)
	assert.Equal(t, "role_admin", gotRole)
}

func TestAuthHeaders_NoHeadersLeavesContextEmpty(t *testing.T) {
	var gotUser, gotRole any

	handler := middlewares.AuthHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUser = r.Context().Value(middlewares.UserIDKey)
		gotRole = r.Context().Value(middlewares.RoleIDKey)
		w.WriteHeader(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/cursos", nil))

	assert.Nil(t, gotUser)
	assert.Nil(t, gotRole)
}
