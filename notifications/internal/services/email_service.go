package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/resend/resend-go/v4"
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
	ReplyTo    string
	SenderType EmailSenderType
	Category   string
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

func (s *resendEmailService) SendEmail(ctx context.Context, msg EmailMessage) (string, error) {
	if len(msg.To) == 0 {
		return "", errors.New("recipient list cannot be empty")
	}

	params := &resend.SendEmailRequest{
		From:    s.resolveFrom(msg),
		To:      msg.To,
		Cc:      msg.Cc,
		Subject: msg.Subject,
		Html:    msg.Html,
		ReplyTo: msg.ReplyTo,
	}

	sent, err := s.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		return "", fmt.Errorf("failed to send email via resend: %w", err)
	}

	return sent.Id, nil
}
