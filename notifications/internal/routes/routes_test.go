package routes_test

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"buenvivir/notifications/internal/controllers"
	"buenvivir/notifications/internal/repositories"
	"buenvivir/notifications/internal/routes"
	"buenvivir/notifications/internal/services"
)

func TestRegisterRoutes_InternalHMACProtection(t *testing.T) {
	hmacSecret := "super-secure-internal-secret"
	repo := repositories.NewMemoryNotificationRepository()
	svc := services.NewNotificationService(repo)
	ctrl := controllers.NewNotificationController(svc)

	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, ctrl, hmacSecret)

	payload := []byte(`{
		"source_service": "campaigns",
		"event_type": "campanas.campana.creada",
		"category": "campaigns",
		"priority": "normal",
		"target_type": "broadcast",
		"recipient_user_ids": ["usr-1"],
		"title": "Nueva Campaña",
		"message": "Detalle de campaña",
		"send_in_app": true
	}`)

	t.Run("MissingHMACSignature_Returns401", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/notifications/send", bytes.NewBuffer(payload))
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 Unauthorized for missing signature, got %d", rec.Code)
		}
	})

	t.Run("InvalidHMACSignature_Returns401", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/notifications/send", bytes.NewBuffer(payload))
		req.Header.Set("X-Internal-Signature", "deadbeef1234")
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 Unauthorized for invalid signature, got %d", rec.Code)
		}
	})

	t.Run("ValidHMACSignature_Returns201", func(t *testing.T) {
		mac := hmac.New(sha256.New, []byte(hmacSecret))
		mac.Write(payload)
		sig := hex.EncodeToString(mac.Sum(nil))

		req := httptest.NewRequest("POST", "/api/v1/notifications/send", bytes.NewBuffer(payload))
		req.Header.Set("X-Internal-Signature", sig)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201 Created for valid signature, got %d (%s)", rec.Code, rec.Body.String())
		}
	})
}
