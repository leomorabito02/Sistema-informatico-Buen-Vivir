package middlewares_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"buenvivir/education/internal/middlewares"
)

func TestRecoverer_ConvertsPanicInto500JSON(t *testing.T) {
	handler := middlewares.Recoverer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/cursos", nil)
	w := httptest.NewRecorder()

	assert.NotPanics(t, func() { handler.ServeHTTP(w, req) })

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var body map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "error interno del servidor", body["error"])
}

func TestRecoverer_PassesThroughWhenNoPanic(t *testing.T) {
	handler := middlewares.Recoverer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("sano"))
	}))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/salud", nil))

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "sano", w.Body.String())
}
