package routes_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"auth/internal/controllers"
	"auth/internal/routes"
)

func TestRegisterRoutes(t *testing.T) {
	mux := http.NewServeMux()
	
	// Create a dummy controller with nil services since we won't invoke the handlers
	authCtrl := controllers.NewAuthController(nil)

	routes.RegisterRoutes(mux, authCtrl)

	// In go 1.22+, ServeMux doesn't export the registered routes easily, 
	// but we can test that the handlers exist by creating a request for them.
	// We won't actually ServeHTTP because it would panic inside the handler with nil services,
	// but we can check if Handler() returns a valid handler instead of the default 404.
	
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	handler, pattern := mux.Handler(req)
	assert.NotNil(t, handler)
	assert.Equal(t, "/health", pattern)

	req2, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	handler2, pattern2 := mux.Handler(req2)
	assert.NotNil(t, handler2)
	assert.Equal(t, "/api/v1/auth/login", pattern2)
}
