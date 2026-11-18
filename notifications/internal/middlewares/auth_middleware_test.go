package middlewares_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"buenvivir/notifications/internal/middlewares"
)

func TestAuthHeaders_UnauthorizedWhenMissing(t *testing.T) {
	// Arrange
	handler := middlewares.AuthHeaders(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest("GET", "/api/v1/notifications", nil)
	rec := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}

func TestAuthHeaders_SuccessWithHeaders(t *testing.T) {
	// Arrange
	var capturedUser string
	var capturedRole string

	handler := middlewares.AuthHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUser, _ = middlewares.GetUserID(r.Context())
		capturedRole = middlewares.GetUserRole(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/notifications", nil)
	req.Header.Set("X-User-ID", "usr-123")
	req.Header.Set("X-User-Role", "role_admin")
	rec := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if capturedUser != "usr-123" {
		t.Fatalf("expected user usr-123, got %s", capturedUser)
	}
	if capturedRole != "role_admin" {
		t.Fatalf("expected role role_admin, got %s", capturedRole)
	}
}

func TestGetUserID_EmptyContext(t *testing.T) {
	_, err := middlewares.GetUserID(context.Background())
	if err == nil {
		t.Fatalf("expected error on empty context, got nil")
	}
}
