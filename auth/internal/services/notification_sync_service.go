package services

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"
)

type NotificationSyncService interface {
	InitUserPreferencesAsync(userID string)
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
