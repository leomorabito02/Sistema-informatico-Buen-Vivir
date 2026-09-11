package controllers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"buenvivir/education/internal/config"
	"buenvivir/education/internal/controllers"
)

func TestHealthController_Health_Success(t *testing.T) {
	cfg := &config.Config{
		AppEnv:                config.AppEnvTesting,
		ContentfulEnvironment: config.ContentfulEnvTesting,
	}
	ctrl := controllers.NewHealthController(cfg)

	req := httptest.NewRequest("GET", "/api/salud", nil)
	w := httptest.NewRecorder()
	ctrl.Health(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.Equal(t, "UP", body["status"])
	assert.Equal(t, "education-service", body["service"])
	assert.Equal(t, "testing", body["appEnv"])
	assert.Equal(t, "testing", body["contentfulEnvironment"])
}
