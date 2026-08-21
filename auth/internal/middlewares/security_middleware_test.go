package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"auth/internal/middlewares"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestSecurityHeaders(t *testing.T) {
	r := chi.NewRouter()
	middlewares.SetupMiddlewares(r)

	r.Get("/", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Contains(t, w.Header().Get("Strict-Transport-Security"), "max-age=31536000")
	assert.Contains(t, w.Header().Get("Content-Security-Policy"), "default-src 'self'")
}

func TestSecurityHeaders_CORS(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://example.com, http://example.org ")
	
	r := chi.NewRouter()
	middlewares.SetupMiddlewares(r)

	r.Get("/", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("OPTIONS", "/", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, "http://example.com", w.Header().Get("Access-Control-Allow-Origin"))
}
