package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"

	"buenvivir/education/internal/middlewares"
)

func newRouterWithSetup() *chi.Mux {
	r := chi.NewRouter()
	middlewares.SetupMiddlewares(r)
	r.Get("/api/salud", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"UP"}`))
	})
	return r
}

func TestSetupMiddlewares_AddsOWASPSecurityHeaders(t *testing.T) {
	r := newRouterWithSetup()

	req := httptest.NewRequest(http.MethodGet, "/api/salud", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	h := w.Header()
	assert.Equal(t, "nosniff", h.Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", h.Get("X-Frame-Options"))
	assert.Contains(t, h.Get("Strict-Transport-Security"), "max-age=31536000")
	assert.Contains(t, h.Get("Content-Security-Policy"), "default-src 'self'")
	assert.Equal(t, "strict-origin-when-cross-origin", h.Get("Referrer-Policy"))
	assert.Contains(t, h.Get("Permissions-Policy"), "geolocation=()")
}

func TestSetupMiddlewares_CORSPreflightReflectsAllowedOrigin(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://ceibuenvivir.org, https://testing.ceibuenvivir.pages.dev")
	r := newRouterWithSetup()

	req := httptest.NewRequest(http.MethodOptions, "/api/salud", nil)
	req.Header.Set("Origin", "https://ceibuenvivir.org")
	req.Header.Set("Access-Control-Request-Method", http.MethodGet)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, "https://ceibuenvivir.org", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
}

func TestSetupMiddlewares_CORSRejectsUnlistedOrigin(t *testing.T) {
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://ceibuenvivir.org")
	r := newRouterWithSetup()

	req := httptest.NewRequest(http.MethodGet, "/api/salud", nil)
	req.Header.Set("Origin", "https://atacante.example")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, http.StatusOK, w.Code)
}
