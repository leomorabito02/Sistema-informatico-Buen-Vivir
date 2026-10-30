package services

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

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

	return &resendEmailService{
		client:            resend.NewClient(apiKey),
		fromNotifications: normNotif,
		fromAlerts:        normAlerts,
	}, nil
}

func normalizeFromEmail(from string) string {
	trimmed := strings.TrimSpace(from)
	if trimmed == "" {
		return defaultFromEmail
	}
	cleaned := strings.Trim(trimmed, "<>")
	if !strings.Contains(cleaned, "@") {
		return fmt.Sprintf("Buen Vivir <notificaciones@%s>", cleaned)
	}
	return trimmed
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

func (s *resendEmailService) SendEmail(ctx context.Context, msg EmailMessage) (string, error) {
	if len(msg.To) == 0 {
		return "", errors.New("recipient list cannot be empty")
	}

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

	sent, err := s.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		return "", fmt.Errorf("failed to send email via resend: %w", err)
	}

	return sent.Id, nil
}
