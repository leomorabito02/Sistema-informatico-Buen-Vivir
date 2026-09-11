package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"buenvivir/education/internal/middlewares"
)

func TestLogging_PassesThroughAndPreservesStatus(t *testing.T) {
	handler := middlewares.Logging(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/cursos", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "ok", w.Body.String())
}

func TestLogging_DefaultsToStatus200WhenHandlerDoesNotSetIt(t *testing.T) {
	handler := middlewares.Logging(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// no llama a WriteHeader: el statusRecorder tiene que reportar 200
		_, _ = w.Write([]byte("implicit 200"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/salud", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
