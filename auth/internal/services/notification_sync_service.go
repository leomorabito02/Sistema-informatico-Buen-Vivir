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

	"auth/api/proto/notificationpb"

	"github.com/golang-jwt/jwt/v5"

	"google.golang.org/api/idtoken"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	target := strings.TrimSpace(os.Getenv("NOTIFICATIONS_GRPC_TARGET"))
	if target == "" {
		target = strings.TrimSpace(os.Getenv("NOTIFICATIONS_SERVICE_URL"))
	}
	if target == "" {
		return
	}

	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		baseURL := strings.TrimRight(target, "/")
		go s.performInitHTTP(baseURL, userID)
		return
	}

	go s.performInitGRPC(target, userID)
}

func (s *notificationSyncService) performInitGRPC(target, userID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Warn("notification_sync: failed to connect gRPC client", "error", err, "target", target, "user_id", userID)
		return
	}
	defer func() { _ = conn.Close() }()

	client := notificationpb.NewNotificationIngestionServiceClient(conn)
	if _, err := client.InitDefaultPreferences(ctx, &notificationpb.InitDefaultPreferencesRequest{UserId: userID}); err != nil {
		slog.Warn("notification_sync: gRPC InitDefaultPreferences failed", "error", err, "user_id", userID)
		return
	}

	slog.Info("notification_sync: default preferences successfully initiated via gRPC", "user_id", userID)
}

func (s *notificationSyncService) performInitHTTP(baseURL, userID string) {
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

	targetURL := baseURL + "/api/v1/notifications/preferences/init"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewBuffer(body))
	if err != nil {
		slog.Warn("notification_sync: failed to construct request", "error", err, "user_id", userID)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", userID)
	// Con JWT_SECRET, Notificaciones ya no confía en X-User-ID: se
	// identifica como servicio interno (role_dev) y el user_id va en el
	// body (auditoría 17/11/2026).
	if tok, err := serviceToken(os.Getenv("JWT_SECRET")); err == nil && tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	if strings.EqualFold(strings.TrimSpace(os.Getenv("UPSTREAM_ID_TOKEN")), "true") {
		if ts, err := idtoken.NewTokenSource(ctx, baseURL); err == nil {
			if tok, err := ts.Token(); err == nil && tok.AccessToken != "" {
				req.Header.Set("X-Serverless-Authorization", "Bearer "+tok.AccessToken)
			}
		}
	}

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

// serviceToken firma un access token corto (60 s) que identifica a auth
// como servicio interno ante Notificaciones. Sin secreto no firma nada.
func serviceToken(secret string) (string, error) {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return "", nil
	}
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":  "auth-service",
		"role": "role_dev",
		"type": "access",
		"iat":  now.Unix(),
		"exp":  now.Add(60 * time.Second).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}
