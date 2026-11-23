package templates

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"os"
	"strings"
	"time"
)

//go:embed email/*.html
var emailTemplatesFS embed.FS

type EmailTemplateID string

const (
	TemplateEmailVerification EmailTemplateID = "email_verification"
	TemplateAccountApproved   EmailTemplateID = "account_approved"
	TemplateTwoFactorCode     EmailTemplateID = "two_factor_code"
	TemplatePasswordRecovery  EmailTemplateID = "password_recovery"
	TemplateEmailChange       EmailTemplateID = "email_change"
	TemplateGeneralAlert      EmailTemplateID = "general_alert"
)

type EmailVerificationData struct {
	RecipientName string
	ActionURL     string
	ExpiryMinutes int
	Year          int
}

type AccountApprovedData struct {
	RecipientName string
	ActionURL     string
	Year          int
}

type TwoFactorCodeData struct {
	RecipientName string
	Code          string
	ExpiryMinutes int
	Year          int
}

type PasswordRecoveryData struct {
	RecipientName string
	ActionURL     string
	ExpiryMinutes int
	Year          int
}

type EmailChangeData struct {
	RecipientName string
	NewEmail      string
	ActionURL     string
	Year          int
}

type GeneralAlertData struct {
	RecipientName string
	Title         string
	Message       string
	Details       string
	ActionURL     string
	ActionText    string
	Category      string
	IsCritical    bool
	Year          int
}

type TemplateEngine interface {
	RenderEmail(templateID EmailTemplateID, data any) (string, error)
}

type templateEngine struct {
	tmpl *template.Template
}

func NewTemplateEngine() (TemplateEngine, error) {
	parsed, err := template.New("").Funcs(funcionesMarca).ParseFS(emailTemplatesFS, "email/*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse embedded email templates: %w", err)
	}
	return &templateEngine{tmpl: parsed}, nil
}

func (e *templateEngine) RenderEmail(templateID EmailTemplateID, data any) (string, error) {
	filename := string(templateID) + ".html"
	t := e.tmpl.Lookup(filename)
	if t == nil {
		return "", fmt.Errorf("email template %q not found", templateID)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to render template %q: %w", templateID, err)
	}

	return buf.String(), nil
}

func CurrentYear() int {
	return time.Now().Year()
}

// LogoPorDefecto: logo del Centro en el front de producción (D7,
// 22/11/2026). En testing sale de FRONTEND_URL (testing.ceibuenvivir.pages.dev).
const LogoPorDefecto = "https://ceibuenvivir.pages.dev/assets/logo.png"

// LogoURL devuelve la URL pública del logo para los correos: FRONTEND_URL
// + /assets/logo.png, o el de producción si FRONTEND_URL no está o es local
// (un cliente de correo no puede leer localhost).
func LogoURL() string {
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("FRONTEND_URL")), "/")
	if base == "" || strings.Contains(base, "localhost") || strings.Contains(base, "127.0.0.1") {
		return LogoPorDefecto
	}
	return base + "/assets/logo.png"
}

// encabezadoMarca son los datos del parcial bv_encabezado.
type encabezadoMarca struct {
	Color     string
	Subtitulo string
}

var funcionesMarca = template.FuncMap{
	"logoURL": LogoURL,
	"marca": func(color, subtitulo string) encabezadoMarca {
		return encabezadoMarca{Color: color, Subtitulo: subtitulo}
	},
}
