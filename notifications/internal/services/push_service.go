package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
)

// MulticastResult contains the summary of batch push notification delivery.
type MulticastResult struct {
	SuccessCount  int
	FailureCount  int
	InvalidTokens []string
}

// PushService defines contract for dispatching push notifications.
type PushService interface {
	SendMulticast(ctx context.Context, tokens []string, title, body string, data map[string]string) (*MulticastResult, error)
}

type fcmMessagingClient interface {
	SendEachForMulticast(ctx context.Context, message *messaging.MulticastMessage) (*messaging.BatchResponse, error)
}

type firebasePushService struct {
	client fcmMessagingClient
}

// NewFirebasePushService initializes Firebase Admin SDK using Google Cloud ADC.
func NewFirebasePushService(ctx context.Context) (PushService, error) {
	app, err := firebase.NewApp(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize firebase app: %w", err)
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize firebase messaging client: %w", err)
	}

	return &firebasePushService{client: client}, nil
}

// NewPushServiceWithClient creates a PushService with a custom or mocked messaging client.
func NewPushServiceWithClient(client fcmMessagingClient) (PushService, error) {
	if client == nil {
		return nil, errors.New("messaging client cannot be nil")
	}
	return &firebasePushService{client: client}, nil
}

// SendMulticast sends a push notification to multiple device tokens and detects invalid or unregistered tokens.
func (s *firebasePushService) SendMulticast(
	ctx context.Context,
	tokens []string,
	title, body string,
	data map[string]string,
) (*MulticastResult, error) {
	if len(tokens) == 0 {
		return &MulticastResult{}, nil
	}

	multicastMsg := &messaging.MulticastMessage{
		Tokens: tokens,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
		Webpush: &messaging.WebpushConfig{
			Notification: &messaging.WebpushNotification{
				Title: title,
				Body:  body,
				Icon:  "/favicon.ico",
			},
		},
	}

	batchResp, err := s.client.SendEachForMulticast(ctx, multicastMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to dispatch multicast push notification: %w", err)
	}

	invalidTokens := make([]string, 0)
	for idx, resp := range batchResp.Responses {
		if resp.Success {
			continue
		}
		if resp.Error != nil && isUnregisteredTokenError(resp.Error) {
			if idx < len(tokens) {
				invalidTokens = append(invalidTokens, tokens[idx])
			}
		}
	}

	return &MulticastResult{
		SuccessCount:  batchResp.SuccessCount,
		FailureCount:  batchResp.FailureCount,
		InvalidTokens: invalidTokens,
	}, nil
}

func isUnregisteredTokenError(err error) bool {
	if messaging.IsUnregistered(err) {
		return true
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "registration-token-not-registered") ||
		strings.Contains(errStr, "unregistered") ||
		strings.Contains(errStr, "invalid registration token")
}
