package services

import (
	"context"
	"testing"
)

func TestNewResendEmailService_EmptyAPIKey(t *testing.T) {
	// Arrange
	apiKey := ""
	fromEmail := "onboarding@resend.dev"

	// Act
	svc, err := NewResendEmailService(apiKey, fromEmail)

	// Assert
	if err == nil {
		t.Fatalf("expected error when apiKey is empty, got nil")
	}
	if svc != nil {
		t.Fatalf("expected nil service when apiKey is empty, got %v", svc)
	}
}

func TestResendEmailService_SendEmail_EmptyRecipients(t *testing.T) {
	// Arrange
	svc, err := NewResendEmailService("re_test_dummy_key", "onboarding@resend.dev")
	if err != nil {
		t.Fatalf("unexpected error creating service: %v", err)
	}
	msg := EmailMessage{
		To:      []string{},
		Subject: "Test",
		Html:    "<p>Hello</p>",
	}

	// Act
	id, err := svc.SendEmail(context.Background(), msg)

	// Assert
	if err == nil {
		t.Fatalf("expected error when recipients list is empty, got nil")
	}
	if id != "" {
		t.Fatalf("expected empty id on error, got %s", id)
	}
}

func TestNormalizeFromEmail(t *testing.T) {
	// Arrange & Act & Assert
	if got := normalizeFromEmail(""); got != "onboarding@resend.dev" {
		t.Errorf("expected onboarding@resend.dev, got %s", got)
	}
	if got := normalizeFromEmail("<notificaciones.ceibuenvivir.org>"); got != "Buen Vivir <notificaciones@notificaciones.ceibuenvivir.org>" {
		t.Errorf("expected Buen Vivir <notificaciones@notificaciones.ceibuenvivir.org>, got %s", got)
	}
	if got := normalizeFromEmail("alerts@example.com"); got != "alerts@example.com" {
		t.Errorf("expected alerts@example.com, got %s", got)
	}
}

func TestResolveFromEmail_DualRouting(t *testing.T) {
	// Arrange
	svcRaw, err := NewDualResendEmailService(
		"re_test_key",
		"Buen Vivir <notificaciones@ceibuenvivir.org>",
		"Buen Vivir Seguridad <seguridad@ceibuenvivir.org>",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	svc, ok := svcRaw.(*resendEmailService)
	if !ok {
		t.Fatalf("expected *resendEmailService")
	}

	// Act & Assert 1: Standard notification
	msgNotif := EmailMessage{
		SenderType: EmailSenderNotification,
		Category:   "system",
	}
	if from := svc.resolveFrom(msgNotif); from != "Buen Vivir <notificaciones@ceibuenvivir.org>" {
		t.Errorf("expected notifications sender, got %s", from)
	}

	// Act & Assert 2: Security alert by SenderType
	msgAlert := EmailMessage{
		SenderType: EmailSenderAlert,
		Category:   "system",
	}
	if from := svc.resolveFrom(msgAlert); from != "Buen Vivir Seguridad <seguridad@ceibuenvivir.org>" {
		t.Errorf("expected alert sender, got %s", from)
	}

	// Act & Assert 3: Security alert by Category
	msgSecCat := EmailMessage{
		SenderType: EmailSenderNotification,
		Category:   "security",
	}
	if from := svc.resolveFrom(msgSecCat); from != "Buen Vivir Seguridad <seguridad@ceibuenvivir.org>" {
		t.Errorf("expected alert sender for security category, got %s", from)
	}
}

