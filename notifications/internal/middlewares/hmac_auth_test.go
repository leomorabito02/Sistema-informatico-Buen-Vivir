package middlewares_test

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"buenvivir/notifications/internal/middlewares"
)

func TestRequireInternalHMAC(t *testing.T) {
	secret := "test-secret-key-12345"
	handler := middlewares.RequireInternalHMAC(secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true}`))
	}))

	payload := []byte(`{"event":"test"}`)

	t.Run("MissingSignature", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/notifications/send", bytes.NewBuffer(payload))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("InvalidSignatureHex", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/notifications/send", bytes.NewBuffer(payload))
		req.Header.Set(middlewares.HeaderInternalSignature, "invalid-hex")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("WrongSignature", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/notifications/send", bytes.NewBuffer(payload))
		req.Header.Set(middlewares.HeaderInternalSignature, hex.EncodeToString([]byte("wrongsignaturewrongsignature1234")))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("ValidSignature", func(t *testing.T) {
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(payload)
		sig := hex.EncodeToString(mac.Sum(nil))

		req := httptest.NewRequest("POST", "/api/v1/notifications/send", bytes.NewBuffer(payload))
		req.Header.Set(middlewares.HeaderInternalSignature, sig)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("EmptySecret_BypassesValidation", func(t *testing.T) {
		bypassHandler := middlewares.RequireInternalHMAC("")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		req := httptest.NewRequest("POST", "/api/v1/notifications/send", bytes.NewBuffer(payload))
		rec := httptest.NewRecorder()

		bypassHandler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 when secret is empty, got %d", rec.Code)
		}
	})
}
