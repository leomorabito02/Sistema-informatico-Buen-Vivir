package services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	authSenderID         = "auth-service"
	notificationsSendURL = "/api/v1/notifications/send"
)

type NotificationSyncService interface {
	InitUserPreferencesAsync(userID string)
	SendVerificationEmail(ctx context.Context, userID, email, name, actionURL string) error
	SendAccountApprovedEmail(ctx context.Context, userID, email, name, actionURL string) error
	SendPasswordRecoveryEmail(ctx context.Context, userID, email, name, actionURL string) error
}

type notificationSyncService struct {
	client *http.Client
}

func NewNotificationSyncService() NotificationSyncService {
	return &notificationSyncService{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (s *notificationSyncService) InitUserPreferencesAsync(userID string) {
	if strings.TrimSpace(userID) == "" {
		return
	}

	rawURL := strings.TrimSpace(os.Getenv("NOTIFICATIONS_SERVICE_URL"))
	if rawURL == "" {
		return
	}

	baseURL := strings.TrimRight(rawURL, "/")
	initURL := baseURL + "/api/v1/notifications/preferences/init"

	go s.performInit(initURL, userID)
}

func (s *notificationSyncService) performInit(targetURL, userID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	payload := map[string]string{
		"user_id": userID,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		slog.Warn("notification_sync: failed to serialize payload", "error", err, "user_id", userID)
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewBuffer(body))
	if err != nil {
		slog.Warn("notification_sync: failed to construct request", "error", err, "user_id", userID)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", userID)

	resp, err := s.client.Do(req)
	if err != nil {
		slog.Warn("notification_sync: request failed", "error", err, "user_id", userID)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		slog.Warn("notification_sync: unexpected status response", "status", resp.StatusCode, "user_id", userID)
		return
	}

	slog.Info("notification_sync: default preferences successfully initiated", "user_id", userID)
}

func (s *notificationSyncService) SendVerificationEmail(ctx context.Context, userID, email, name, actionURL string) error {
	rawURL := strings.TrimSpace(os.Getenv("NOTIFICATIONS_SERVICE_URL"))
	if rawURL == "" {
		slog.Warn("notification_sync: NOTIFICATIONS_SERVICE_URL not configured, skipping verification email")
		return nil
	}

	dataJSON, _ := json.Marshal(map[string]interface{}{
		"template_id":    "email_verification",
		"recipient_name": name,
		"email":          email,
		"action_url":     actionURL,
		"expiry_minutes": 1440,
	})

	payload := map[string]interface{}{
		"source_service":     "auth",
		"event_type":         "auth.email_verification",
		"category":           "security",
		"priority":           "high",
		"sender_type":        "system",
		"sender_id":          authSenderID,
		"recipient_user_ids": []string{userID},
		"title":              "Verifica tu correo electrónico - Buen Vivir",
		"message":            "Por favor verifica tu correo electrónico para activar tu cuenta.",
		"action_url":         actionURL,
		"data":               string(dataJSON),
		"send_email":         true,
		"is_critical":        true,
	}

	return s.postNotification(ctx, strings.TrimRight(rawURL, "/")+notificationsSendURL, payload)
}

func (s *notificationSyncService) SendAccountApprovedEmail(ctx context.Context, userID, email, name, actionURL string) error {
	rawURL := strings.TrimSpace(os.Getenv("NOTIFICATIONS_SERVICE_URL"))
	if rawURL == "" {
		slog.Warn("notification_sync: NOTIFICATIONS_SERVICE_URL not configured, skipping approval email")
		return nil
	}

	dataJSON, _ := json.Marshal(map[string]interface{}{
		"template_id":    "account_approved",
		"recipient_name": name,
		"email":          email,
		"action_url":     actionURL,
	})

	payload := map[string]interface{}{
		"source_service":     "auth",
		"event_type":         "auth.account_approved",
		"category":           "system",
		"priority":           "high",
		"sender_type":        "system",
		"sender_id":          authSenderID,
		"recipient_user_ids": []string{userID},
		"title":              "¡Tu cuenta ha sido aprobada! - Buen Vivir",
		"message":            "Tu cuenta ha sido aprobada exitosamente.",
		"action_url":         actionURL,
		"data":               string(dataJSON),
		"send_email":         true,
	}

	return s.postNotification(ctx, strings.TrimRight(rawURL, "/")+notificationsSendURL, payload)
}

func (s *notificationSyncService) SendPasswordRecoveryEmail(ctx context.Context, userID, email, name, actionURL string) error {
	rawURL := strings.TrimSpace(os.Getenv("NOTIFICATIONS_SERVICE_URL"))
	if rawURL == "" {
		slog.Warn("notification_sync: NOTIFICATIONS_SERVICE_URL not configured, skipping password recovery email")
		return nil
	}

	dataJSON, _ := json.Marshal(map[string]interface{}{
		"template_id":    "password_recovery",
		"recipient_name": name,
		"email":          email,
		"action_url":     actionURL,
		"expiry_minutes": 30,
	})

	payload := map[string]interface{}{
		"source_service":     "auth",
		"event_type":         "auth.password_recovery",
		"category":           "security",
		"priority":           "high",
		"sender_type":        "system",
		"sender_id":          authSenderID,
		"recipient_user_ids": []string{userID},
		"title":              "Recuperación de contraseña - Buen Vivir",
		"message":            "Recibimos una solicitud para restablecer tu contraseña.",
		"action_url":         actionURL,
		"data":               string(dataJSON),
		"send_email":         true,
		"is_critical":        true,
	}

	return s.postNotification(ctx, strings.TrimRight(rawURL, "/")+notificationsSendURL, payload)
}

func (s *notificationSyncService) postNotification(ctx context.Context, targetURL string, payload map[string]interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	secret := strings.TrimSpace(os.Getenv("INTERNAL_HMAC_SECRET"))
	if secret != "" {
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(body)
		req.Header.Set("X-Internal-Signature", hex.EncodeToString(mac.Sum(nil)))
	}

	resp, err := s.client.Do(req)
	if err != nil {
		slog.Warn("notification_sync: request to notifications service failed", "error", err, "url", targetURL)
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		slog.Warn("notification_sync: notifications service returned non-2xx status", "status", resp.StatusCode)
		return nil
	}

	return nil
}
