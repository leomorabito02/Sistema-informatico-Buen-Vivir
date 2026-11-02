package services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"buenvivir/campaigns/internal/models"
)

// NotificationDispatcher defines the contract for dispatching campaign notification events.
type NotificationDispatcher interface {
	DispatchCampaignCreated(ctx context.Context, campaign models.Campaign, creadoPor string) error
}

type httpNotificationDispatcher struct {
	baseURL    string
	secret     string
	httpClient *http.Client
}

// NewHTTPNotificationDispatcher creates a NotificationDispatcher communicating over HTTP with HMAC signing.
func NewHTTPNotificationDispatcher(serviceURL, secret string) NotificationDispatcher {
	cleanURL := strings.TrimRight(strings.TrimSpace(serviceURL), "/")
	if cleanURL == "" {
		cleanURL = "http://notifications-service:8080"
	}
	return &httpNotificationDispatcher{
		baseURL: cleanURL,
		secret:  strings.TrimSpace(secret),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// notificationPayload matches the ingestion DTO of notifications-service.
type notificationPayload struct {
	SourceService string  `json:"source_service"`
	EventType     string  `json:"event_type"`
	Category      string  `json:"category"`
	Priority      string  `json:"priority"`
	SenderType    string  `json:"sender_type"`
	SenderID      string  `json:"sender_id"`
	TargetType    string  `json:"target_type"`
	Title         string  `json:"title"`
	Message       string  `json:"message"`
	ActionURL     *string `json:"action_url,omitempty"`
	Data          string  `json:"data,omitempty"`
	IsCritical    bool    `json:"is_critical"`
	SendEmail     bool    `json:"send_email"`
	SendPush      bool    `json:"send_push"`
	SendInApp     bool    `json:"send_in_app"`
}

func (d *httpNotificationDispatcher) DispatchCampaignCreated(ctx context.Context, campaign models.Campaign, creadoPor string) error {
	actionURL := "/campaigns/" + campaign.ID
	dataJSON, _ := json.Marshal(map[string]interface{}{
		"campaign_id": campaign.ID,
		"action_url":  actionURL,
		"title":       campaign.Titulo,
		"image_url":   campaign.ImagenURL,
	})

	payload := notificationPayload{
		SourceService: "campaigns",
		EventType:     "campanas.campana.creada",
		Category:      "campaigns",
		Priority:      "normal",
		SenderType:    "user",
		SenderID:      creadoPor,
		TargetType:    "broadcast",
		Title:         "Nueva Campaña: " + campaign.Titulo,
		Message:       campaign.Texto,
		ActionURL:     &actionURL,
		Data:          string(dataJSON),
		IsCritical:    false,
		SendEmail:     true,
		SendPush:      true,
		SendInApp:     true,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		slog.Warn("failed to serialize campaign notification payload", slog.String("error", err.Error()), slog.String("campaign_id", campaign.ID))
		return err
	}

	endpoint := d.baseURL + "/api/v1/notifications/send"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(bodyBytes))
	if err != nil {
		slog.Warn("failed to create notification request", slog.String("error", err.Error()), slog.String("campaign_id", campaign.ID))
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if d.secret != "" {
		mac := hmac.New(sha256.New, []byte(d.secret))
		mac.Write(bodyBytes)
		req.Header.Set("X-Internal-Signature", hex.EncodeToString(mac.Sum(nil)))
	}

	resp, err := d.httpClient.Do(req)
	if err != nil {
		slog.Warn("failed to dispatch campaign notification", slog.String("error", err.Error()), slog.String("campaign_id", campaign.ID))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err := fmt.Errorf("notifications service responded with status code: %d", resp.StatusCode)
		slog.Warn("campaign notification dispatch received non-success response", slog.Int("status", resp.StatusCode), slog.String("campaign_id", campaign.ID))
		return err
	}

	slog.Info("campaign created notification dispatched successfully", slog.String("campaign_id", campaign.ID))
	return nil
}
