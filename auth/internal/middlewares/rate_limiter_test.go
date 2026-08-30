package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"auth/internal/middlewares"
	"github.com/stretchr/testify/assert"
)

func TestRateLimiter(t *testing.T) {
	middleware := middlewares.RateLimiter()

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Enviar algunas solicitudes, no debería bloquear de inmediato (100 por min)
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "127.0.0.1:1234"
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}
}

func TestStrictRateLimiter(t *testing.T) {
	middleware := middlewares.StrictRateLimiter()

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Enviar 11 solicitudes (el límite es 10)
	var lastCode int
	for i := 0; i < 11; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.168.0.1:1234"
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)
		lastCode = w.Code
	}

	assert.Equal(t, http.StatusTooManyRequests, lastCode)
}
