package services

import (
	"context"
	"strings"
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

func TestHtmlToPlainText(t *testing.T) {
	htmlInput := `<html><head><style>body { color: red; }</style></head><body><h1>Hola Juan</h1><p>Por favor confirma tu email &amp; activa tu cuenta.<br>Haz clic en el enlace.</p><div>&copy; 2026 Buen Vivir</div></body></html>`
	plain := htmlToPlainText(htmlInput)

	if !strings.Contains(plain, "Hola Juan") {
		t.Errorf("expected plain text to contain 'Hola Juan', got: %q", plain)
	}
	if !strings.Contains(plain, "confirma tu email & activa tu cuenta.") {
		t.Errorf("expected decoded &amp;, got: %q", plain)
	}
	if !strings.Contains(plain, "© 2026 Buen Vivir") {
		t.Errorf("expected decoded &copy;, got: %q", plain)
	}
	if strings.Contains(plain, "<h1") || strings.Contains(plain, "<style") {
		t.Errorf("expected tags to be stripped, got: %q", plain)
	}
}

func TestBuildSendEmailRequest_InlineLogoAttachment(t *testing.T) {
	// Arrange
	svcRaw, err := NewResendEmailService("re_test_key", "onboarding@resend.dev")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	svc := svcRaw.(*resendEmailService)

	// Act
	withLogo := svc.buildSendEmailRequest(EmailMessage{
		To:      []string{"user@example.com"},
		Subject: "Con logo",
		Html:    `<img src="cid:buenvivir-logo" alt="Logo">`,
	})
	withoutLogo := svc.buildSendEmailRequest(EmailMessage{
		To:      []string{"user@example.com"},
		Subject: "Sin logo",
		Html:    `<p>Solo texto</p>`,
	})

	// Assert
	if len(withLogo.Attachments) != 1 {
		t.Fatalf("expected 1 inline logo attachment, got %d", len(withLogo.Attachments))
	}
	att := withLogo.Attachments[0]
	if att.ContentId != "buenvivir-logo" || att.ContentType != "image/png" || att.Filename != "logo.png" || len(att.Content) == 0 {
		t.Errorf("unexpected inline logo attachment metadata: %+v (len=%d)", att, len(att.Content))
	}
	if len(withoutLogo.Attachments) != 0 {
		t.Errorf("expected no attachments when cid is not referenced, got %d", len(withoutLogo.Attachments))
	}
}

