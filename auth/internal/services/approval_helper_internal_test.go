package services

import (
	"context"
	"testing"

	"auth/internal/middlewares"
)

func TestResolveApprover(t *testing.T) {
	logged := context.WithValue(context.Background(), middlewares.UserIDKey, "usr-admin")

	if got := resolveApprover(logged, "usr-explicito"); got == nil || *got != "usr-explicito" {
		t.Errorf("approved_by explícito: got %v", got)
	}
	if got := resolveApprover(logged, ""); got == nil || *got != "usr-admin" {
		t.Errorf("sin approved_by usa el usuario logueado: got %v", got)
	}
	if got := resolveApprover(context.Background(), ""); got != nil {
		t.Errorf("sin approved_by ni sesión: got %v, want nil", *got)
	}
}
