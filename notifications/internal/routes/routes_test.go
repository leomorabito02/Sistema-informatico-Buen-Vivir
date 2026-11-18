package routes_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"buenvivir/notifications/internal/controllers"
	"buenvivir/notifications/internal/repositories"
	"buenvivir/notifications/internal/routes"
	"buenvivir/notifications/internal/services"
)

func TestRegisterRoutes(t *testing.T) {
	repo := repositories.NewMemoryNotificationRepository()
	svc := services.NewNotificationService(repo)
	ctrl := controllers.NewNotificationController(svc)

	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, ctrl)

	t.Run("HealthCheckRouteExists", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK for /health, got %d", rec.Code)
		}
	})

	t.Run("LegacySendRouteDoesNotExist", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/notifications/send", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code == http.StatusOK || rec.Code == http.StatusCreated || rec.Code == http.StatusUnauthorized {
			t.Fatalf("expected legacy /send endpoint to be gone, got %d", rec.Code)
		}
	})
}
