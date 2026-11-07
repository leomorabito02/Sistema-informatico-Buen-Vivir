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
	assert.Equal(t, "SAMEORIGIN", w.Header().Get("X-Frame-Options"))
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

func TestOptionsInterceptor(t *testing.T) {
	calledNext := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calledNext = true
		w.WriteHeader(http.StatusOK)
	})

	interceptor := middlewares.OptionsInterceptor(next)

	// OPTIONS request with origin
	req := httptest.NewRequest("OPTIONS", "/api/test", nil)
	req.Header.Set("Origin", "http://localhost:4200")
	w := httptest.NewRecorder()
	interceptor.ServeHTTP(w, req)
	assert.False(t, calledNext)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "http://localhost:4200", w.Header().Get("Access-Control-Allow-Origin"))

	// Non-OPTIONS request passes to next
	calledNext = false
	reqGet := httptest.NewRequest("GET", "/api/test", nil)
	wGet := httptest.NewRecorder()
	interceptor.ServeHTTP(wGet, reqGet)
	assert.True(t, calledNext)
}

func TestCorsMiddleware_Origins(t *testing.T) {
	r := chi.NewRouter()
	middlewares.SetupMiddlewares(r)
	r.Get("/data", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Allowed origin via default rules (.ceibuenvivir.pages.dev)
	req := httptest.NewRequest("GET", "/data", nil)
	req.Header.Set("Origin", "https://preview.ceibuenvivir.pages.dev")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, "https://preview.ceibuenvivir.pages.dev", w.Header().Get("Access-Control-Allow-Origin"))

	// Disallowed origin
	reqDis := httptest.NewRequest("GET", "/data", nil)
	reqDis.Header.Set("Origin", "http://evil.com")
	wDis := httptest.NewRecorder()
	r.ServeHTTP(wDis, reqDis)
	assert.Empty(t, wDis.Header().Get("Access-Control-Allow-Origin"))
}
