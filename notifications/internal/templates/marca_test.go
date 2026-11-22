package templates

import (
	"strings"
	"testing"
)

// D7 (22/11/2026): todas las plantillas comparten encabezado con logo y pie
// institucional, sin etiquetas que Outlook ignora, y conservan su contenido.
func TestPlantillas_IdentidadDelCentro(t *testing.T) {
	t.Setenv("FRONTEND_URL", "https://testing.ceibuenvivir.pages.dev")
	eng, err := NewTemplateEngine()
	if err != nil {
		t.Fatal(err)
	}
	casos := []struct {
		id       EmailTemplateID
		data     any
		contiene []string
	}{
		{TemplateEmailVerification, EmailVerificationData{RecipientName: "Paula", ActionURL: "https://x/verify-email?token=abc", Year: 2026}, []string{"https://x/verify-email?token=abc", "Paula"}},
		{TemplateAccountApproved, AccountApprovedData{RecipientName: "Paula", ActionURL: "https://x/login", Year: 2026}, []string{"https://x/login"}},
		{TemplateTwoFactorCode, TwoFactorCodeData{Code: "123456", ExpiryMinutes: 10, Year: 2026}, []string{"123456", "Seguridad y Control de Acceso"}},
		{TemplatePasswordRecovery, PasswordRecoveryData{ActionURL: "https://x/change-password?token=t", Year: 2026}, []string{"https://x/change-password?token=t", "Seguridad y Recuperación"}},
		{TemplateEmailChange, EmailChangeData{NewEmail: "nuevo@example.com", ActionURL: "https://x/c", Year: 2026}, []string{"nuevo@example.com", "Seguridad de la Cuenta"}},
		{TemplateGeneralAlert, GeneralAlertData{Title: "Aviso", Message: "Mensaje", IsCritical: true, Year: 2026}, []string{"Aviso", "Mensaje", "#b91c1c"}},
		{TemplateGeneralAlert, GeneralAlertData{Title: "Aviso", Message: "Mensaje", Year: 2026}, []string{"Centro de Alertas y Notificaciones", "#485839"}},
	}
	for _, c := range casos {
		html, err := eng.RenderEmail(c.id, c.data)
		if err != nil {
			t.Fatalf("%s: %v", c.id, err)
		}
		esperado := append([]string{
			"https://testing.ceibuenvivir.pages.dev/assets/logo.png",
			"Centro de Ecología Integral",
			"&copy; 2026 Centro de Ecología Integral Buen Vivir",
			"ceibuenvivir7@gmail.com",
		}, c.contiene...)
		for _, e := range esperado {
			if !strings.Contains(html, e) {
				t.Errorf("%s: falta %q", c.id, e)
			}
		}
		for _, prohibido := range []string{"<header", "<main", "<footer", "Sistema Informático Integral"} {
			if strings.Contains(html, prohibido) {
				t.Errorf("%s: no debería contener %q", c.id, prohibido)
			}
		}
	}
}

func TestLogoURL(t *testing.T) {
	casos := map[string]string{
		"":                      LogoPorDefecto,
		"http://localhost:4200": LogoPorDefecto,
		"https://testing.ceibuenvivir.pages.dev/": "https://testing.ceibuenvivir.pages.dev/assets/logo.png",
	}
	for env, want := range casos {
		t.Setenv("FRONTEND_URL", env)
		if got := LogoURL(); got != want {
			t.Errorf("FRONTEND_URL=%q: %q, want %q", env, got, want)
		}
	}
}
