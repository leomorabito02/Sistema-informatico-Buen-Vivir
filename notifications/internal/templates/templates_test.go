package templates

import (
	"strings"
	"testing"
)

func TestNewTemplateEngine(t *testing.T) {
	// Act
	engine, err := NewTemplateEngine()

	// Assert
	if err != nil {
		t.Fatalf("expected template engine to initialize, got error: %v", err)
	}
	if engine == nil {
		t.Fatal("expected non-nil template engine")
	}
}

func TestRenderEmail_AllTemplates(t *testing.T) {
	// Arrange
	engine, err := NewTemplateEngine()
	if err != nil {
		t.Fatalf("init error: %v", err)
	}

	tests := []struct {
		name       string
		templateID EmailTemplateID
		data       any
		contains   string
	}{
		{
			name:       "Email Verification",
			templateID: TemplateEmailVerification,
			data: EmailVerificationData{
				RecipientName: "Maria Garcia",
				ActionURL:     "https://buenvivir.org/verify?token=abc",
				ExpiryMinutes: 60,
				Year:          2026,
			},
			contains: "https://buenvivir.org/verify?token=abc",
		},
		{
			name:       "Account Approved",
			templateID: TemplateAccountApproved,
			data: AccountApprovedData{
				RecipientName: "Carlos Lopez",
				ActionURL:     "https://buenvivir.org/login",
				Year:          2026,
			},
			contains: "¡Tu cuenta ha sido aprobada!",
		},
		{
			name:       "Two Factor Code",
			templateID: TemplateTwoFactorCode,
			data: TwoFactorCodeData{
				RecipientName: "Ana Silva",
				Code:          "482910",
				ExpiryMinutes: 10,
				Year:          2026,
			},
			contains: "482910",
		},
		{
			name:       "Password Recovery",
			templateID: TemplatePasswordRecovery,
			data: PasswordRecoveryData{
				RecipientName: "Juan Perez",
				ActionURL:     "https://buenvivir.org/reset-password?token=xyz",
				ExpiryMinutes: 30,
				Year:          2026,
			},
			contains: "Solicitud de restablecimiento de contraseña",
		},
		{
			name:       "Email Change",
			templateID: TemplateEmailChange,
			data: EmailChangeData{
				RecipientName: "Sofia Cuozzo",
				NewEmail:      "sofia.new@example.com",
				ActionURL:     "https://buenvivir.org/confirm-email?token=123",
				Year:          2026,
			},
			contains: "sofia.new@example.com",
		},
		{
			name:       "General Alert",
			templateID: TemplateGeneralAlert,
			data: GeneralAlertData{
				RecipientName: "Admin",
				Title:         "Mantenimiento Programado",
				Message:       "El sistema se suspenderá por 15 minutos.",
				Details:       "Servidores en mantenimiento preventivo.",
				ActionURL:     "https://buenvivir.org/status",
				Category:      "system",
				IsCritical:    true,
				Year:          2026,
			},
			contains: "Mantenimiento Programado",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			html, err := engine.RenderEmail(tc.templateID, tc.data)

			// Assert
			if err != nil {
				t.Fatalf("unexpected error rendering %s: %v", tc.name, err)
			}
			if !strings.Contains(html, tc.contains) {
				t.Errorf("expected rendered html to contain %q, but got:\n%s", tc.contains, html)
			}
			if !strings.Contains(html, "Centro de Ecología Integral Buen Vivir") {
				t.Errorf("expected rendered html to contain institutional name 'Centro de Ecología Integral Buen Vivir'")
			}
			if !strings.Contains(html, "/assets/logo.png") {
				t.Errorf("expected rendered html to contain logo path '/assets/logo.png'")
			}
			if !strings.Contains(html, "ceibuenvivir7@gmail.com") {
				t.Errorf("expected rendered html to contain contact email 'ceibuenvivir7@gmail.com'")
			}
			if strings.Contains(strings.ToLower(html), "asociación civil") {
				t.Errorf("expected rendered html not to contain 'Asociación Civil'")
			}
		})
	}
}

func TestRenderEmail_NotFound(t *testing.T) {
	// Arrange
	engine, err := NewTemplateEngine()
	if err != nil {
		t.Fatalf("init error: %v", err)
	}

	// Act
	_, err = engine.RenderEmail("non_existent_template", nil)

	// Assert
	if err == nil {
		t.Fatalf("expected error for non-existent template, got nil")
	}
}
