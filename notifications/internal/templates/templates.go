package templates

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
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
	parsed, err := template.ParseFS(emailTemplatesFS, "email/*.html")
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
