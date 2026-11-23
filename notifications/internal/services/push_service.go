package services

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
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

// NewFirebasePushService initializes Firebase Admin SDK using Google Cloud ADC or inline credentials.
func NewFirebasePushService(ctx context.Context) (PushService, error) {
	projectID := strings.TrimSpace(os.Getenv("FIREBASE_PROJECT_ID"))
	if projectID == "" {
		projectID = strings.TrimSpace(os.Getenv("GCP_PROJECT_ID"))
	}

	var fbCfg *firebase.Config
	if projectID != "" {
		fbCfg = &firebase.Config{ProjectID: projectID}
	}

	opts, err := resolveFirebaseClientOptions()
	if err != nil {
		return nil, err
	}

	app, err := firebase.NewApp(ctx, fbCfg, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize firebase app: %w", err)
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize firebase messaging client: %w", err)
	}

	return &firebasePushService{client: client}, nil
}

// resolveFirebaseClientOptions builds Firebase client options prioritizing base64-encoded
// inline credentials over legacy GOOGLE_APPLICATION_CREDENTIALS and unsetting stale env paths.
func resolveFirebaseClientOptions() ([]option.ClientOption, error) {
	for _, envKey := range []string{"FIREBASE_CREDENTIALS_BASE64", "GOOGLE_APPLICATION_CREDENTIALS_BASE64"} {
		if b64 := strings.TrimSpace(os.Getenv(envKey)); b64 != "" {
			decoded, err := base64.StdEncoding.DecodeString(b64)
			if err != nil {
				return nil, fmt.Errorf("error decoding %s: %w", envKey, err)
			}
			if creds := strings.TrimSpace(string(decoded)); creds != "" {
				_ = os.Unsetenv("GOOGLE_APPLICATION_CREDENTIALS")
				return []option.ClientOption{option.WithCredentialsJSON([]byte(creds))}, nil //nolint:staticcheck // inline service account credentials
			}
		}
	}

	if credentialsJSON := strings.TrimSpace(os.Getenv("FIREBASE_CREDENTIALS_JSON")); credentialsJSON != "" {
		_ = os.Unsetenv("GOOGLE_APPLICATION_CREDENTIALS")
		return []option.ClientOption{option.WithCredentialsJSON([]byte(credentialsJSON))}, nil //nolint:staticcheck // inline service account credentials
	}

	if credentialsEnv := strings.TrimSpace(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")); credentialsEnv != "" {
		if strings.HasPrefix(credentialsEnv, "{") {
			if os.Getenv("FIREBASE_CREDENTIALS_JSON") == "" {
				_ = os.Setenv("FIREBASE_CREDENTIALS_JSON", credentialsEnv)
			}
			_ = os.Unsetenv("GOOGLE_APPLICATION_CREDENTIALS")
			return []option.ClientOption{option.WithCredentialsJSON([]byte(credentialsEnv))}, nil //nolint:staticcheck // inline service account credentials
		}
		return []option.ClientOption{option.WithCredentialsFile(credentialsEnv)}, nil //nolint:staticcheck // credential file path
	}

	return nil, nil
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
