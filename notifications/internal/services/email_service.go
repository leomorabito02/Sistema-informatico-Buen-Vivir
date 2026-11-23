package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/mail"
	"regexp"
	"strings"

	"buenvivir/notifications/internal/templates"
	"github.com/resend/resend-go/v4"
)

var (
	tagRegex          = regexp.MustCompile(`<[^>]*>`)
	headRegex         = regexp.MustCompile(`(?s)<head.*?</head>`)
	styleRegex        = regexp.MustCompile(`(?s)<style.*?</style>`)
	scriptRegex       = regexp.MustCompile(`(?s)<script.*?</script>`)
	multiSpaceRegex   = regexp.MustCompile(`[ \t]+`)
	multiNewlineRegex = regexp.MustCompile(`\n{3,}`)
)

type EmailSenderType string

const (
	EmailSenderNotification EmailSenderType = "notification"
	EmailSenderAlert        EmailSenderType = "alert"
	defaultFromEmail                        = "onboarding@resend.dev"
)

type EmailMessage struct {
	To         []string
	Cc         []string
	Subject    string
	Html       string
	Text       string
	ReplyTo    string
	SenderType EmailSenderType
	Category   string
	Headers    map[string]string
}

type EmailService interface {
	SendEmail(ctx context.Context, msg EmailMessage) (string, error)
}

type resendEmailService struct {
	client            *resend.Client
	fromNotifications string
	fromAlerts        string
}

func NewResendEmailService(apiKey string, fromEmail string) (EmailService, error) {
	return NewDualResendEmailService(apiKey, fromEmail, fromEmail)
}

func NewDualResendEmailService(apiKey string, fromNotifications string, fromAlerts string) (EmailService, error) {
	if apiKey == "" {
		return nil, errors.New("RESEND_API_KEY is not configured")
	}

	normNotif := normalizeFromEmail(fromNotifications)
	normAlerts := normalizeFromEmail(fromAlerts)
	if normAlerts == "" || normAlerts == defaultFromEmail {
		normAlerts = normNotif
	}
	if (normNotif == "" || normNotif == defaultFromEmail) && normAlerts != "" && normAlerts != defaultFromEmail {
		normNotif = normAlerts
	}

	slog.Info("resend email service configured",
		slog.String("from_notifications", normNotif),
		slog.String("from_alerts", normAlerts),
	)

	return &resendEmailService{
		client:            resend.NewClient(apiKey),
		fromNotifications: normNotif,
		fromAlerts:        normAlerts,
	}, nil
}

// normalizeFromEmail coerces env-provided sender strings into a Resend-valid
// `email@example.com` or `Name <email@example.com>` address.
func normalizeFromEmail(from string) string {
	trimmed := strings.TrimSpace(strings.Trim(strings.TrimSpace(from), `"'`))
	if trimmed == "" {
		return defaultFromEmail
	}

	if !strings.Contains(trimmed, "@") {
		cleaned := strings.TrimSpace(strings.Trim(trimmed, "<>\"'"))
		if cleaned != "" && strings.Contains(cleaned, ".") && !strings.ContainsAny(cleaned, " \t\r\n") {
			return fmt.Sprintf("Buen Vivir <notificaciones@%s>", cleaned)
		}
		return defaultFromEmail
	}

	if addr, err := mail.ParseAddress(trimmed); err == nil {
		return formatResendAddress(addr.Name, addr.Address)
	}

	// Recover common secret formatting mistakes such as missing angle brackets
	// ("Buen Vivir user@domain.com") or an unclosed "<".
	fields := strings.Fields(trimmed)
	for i, f := range fields {
		if !strings.Contains(f, "@") {
			continue
		}
		candidate := strings.Trim(f, "<>\"',;")
		if addr, err := mail.ParseAddress(candidate); err == nil {
			name := strings.Join(fields[:i], " ")
			return formatResendAddress(name, addr.Address)
		}
	}

	slog.Warn("invalid RESEND_FROM_EMAIL format, falling back to default sender", slog.String("fallback", defaultFromEmail))
	return defaultFromEmail
}

func formatResendAddress(name, address string) string {
	cleanName := strings.TrimSpace(strings.NewReplacer(`"`, "", `'`, "", "<", "", ">", "", "\r", "", "\n", "").Replace(name))
	if cleanName == "" {
		return address
	}
	return fmt.Sprintf("%s <%s>", cleanName, address)
}

func (s *resendEmailService) resolveFrom(msg EmailMessage) string {
	if msg.SenderType == EmailSenderAlert || strings.EqualFold(msg.Category, "security") {
		if s.fromAlerts != "" {
			return s.fromAlerts
		}
	}
	if s.fromNotifications != "" {
		return s.fromNotifications
	}
	return defaultFromEmail
}

func htmlToPlainText(htmlStr string) string {
	if htmlStr == "" {
		return ""
	}
	s := headRegex.ReplaceAllString(htmlStr, "")
	s = styleRegex.ReplaceAllString(s, "")
	s = scriptRegex.ReplaceAllString(s, "")

	s = strings.ReplaceAll(s, "<br>", "\n")
	s = strings.ReplaceAll(s, "<br/>", "\n")
	s = strings.ReplaceAll(s, "<br />", "\n")
	s = strings.ReplaceAll(s, "</p>", "\n\n")
	s = strings.ReplaceAll(s, "</div>", "\n")
	s = strings.ReplaceAll(s, "</h1>", "\n\n")
	s = strings.ReplaceAll(s, "</h2>", "\n\n")
	s = strings.ReplaceAll(s, "</h3>", "\n\n")
	s = strings.ReplaceAll(s, "</li>", "\n")

	s = tagRegex.ReplaceAllString(s, "")

	s = strings.ReplaceAll(s, "&nbsp;", " ")
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&quot;", "\"")
	s = strings.ReplaceAll(s, "&#39;", "'")
	s = strings.ReplaceAll(s, "&copy;", "©")

	lines := strings.Split(s, "\n")
	var cleanedLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(multiSpaceRegex.ReplaceAllString(line, " "))
		cleanedLines = append(cleanedLines, trimmed)
	}
	result := strings.Join(cleanedLines, "\n")
	result = multiNewlineRegex.ReplaceAllString(result, "\n\n")
	return strings.TrimSpace(result)
}

func (s *resendEmailService) buildSendEmailRequest(msg EmailMessage) *resend.SendEmailRequest {
	textBody := strings.TrimSpace(msg.Text)
	if textBody == "" && msg.Html != "" {
		textBody = htmlToPlainText(msg.Html)
	}

	headers := map[string]string{
		"Auto-Submitted":           "auto-generated",
		"X-Auto-Response-Suppress": "All",
	}
	for k, v := range msg.Headers {
		if strings.TrimSpace(k) != "" && strings.TrimSpace(v) != "" {
			headers[k] = v
		}
	}

	replyTo := strings.TrimSpace(msg.ReplyTo)
	if replyTo == "" {
		replyTo = s.fromNotifications
	}

	params := &resend.SendEmailRequest{
		From:    s.resolveFrom(msg),
		To:      msg.To,
		Cc:      msg.Cc,
		Subject: msg.Subject,
		Html:    msg.Html,
		Text:    textBody,
		ReplyTo: replyTo,
		Headers: headers,
	}

	if msg.Category != "" {
		params.Tags = []resend.Tag{{Name: "category", Value: msg.Category}}
	}

	if strings.Contains(msg.Html, templates.LogoCID) {
		if logoBytes := templates.EmbeddedLogoPNG(); len(logoBytes) > 0 {
			params.Attachments = append(params.Attachments, &resend.Attachment{
				Content:     logoBytes,
				Filename:    "logo.png",
				ContentType: "image/png",
				ContentId:   templates.LogoContentID,
			})
		}
	}

	return params
}

func (s *resendEmailService) SendEmail(ctx context.Context, msg EmailMessage) (string, error) {
	if len(msg.To) == 0 {
		return "", errors.New("recipient list cannot be empty")
	}

	params := s.buildSendEmailRequest(msg)

	sent, err := s.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		return "", fmt.Errorf("failed to send email via resend: %w", err)
	}

	return sent.Id, nil
}
