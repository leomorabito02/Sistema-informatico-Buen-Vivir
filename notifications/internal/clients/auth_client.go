package clients

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"buenvivir/notifications/api/proto/authpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	defaultAuthTimeout    = 5 * time.Second
	internalUsersBasePath = "/api/v1/directory/internal/users/"
)

// RecipientContact represents contact details for a campaign recipient.
type RecipientContact struct {
	UserID          string
	Email           string
	FirstName       string
	LastName        string
	IsEmailVerified bool
}

// AuthClient abstracts HTTP or gRPC communication with the Auth service for device token management and recipient queries.
type AuthClient interface {
	GetUserDeviceTokens(ctx context.Context, userID string) ([]string, error)
	RevokeDeviceTokens(ctx context.Context, userID string, tokens []string) error
	GetCampaignRecipients(ctx context.Context) ([]RecipientContact, error)
	IsUserEmailVerified(ctx context.Context, userID string) (bool, error)
	GetUserContact(ctx context.Context, userID string) (*RecipientContact, error)
	GetActiveUserIDsByRoles(ctx context.Context, roleIDs []string) ([]string, error)
	GetParticipantUserID(ctx context.Context, participantID string) (string, error)
	IssueActionLink(ctx context.Context, requestID, userID, actionType string) (string, error)
	Close() error
}

type authClientOptions struct {
	idToken   func(ctx context.Context) (string, error)
	jwtSecret string
}

// AuthClientOption configures optional HTTP authentication parameters for AuthClient.
type AuthClientOption func(*authClientOptions)

// WithIDToken sets the Cloud Run ID token provider (X-Serverless-Authorization).
func WithIDToken(fn func(ctx context.Context) (string, error)) AuthClientOption {
	return func(o *authClientOptions) {
		o.idToken = fn
	}
}

// WithJWTSecret sets the shared JWT secret used to sign internal service-to-service requests.
func WithJWTSecret(secret string) AuthClientOption {
	return func(o *authClientOptions) {
		o.jwtSecret = strings.TrimSpace(secret)
	}
}

type authClient struct {
	conn   *grpc.ClientConn
	client authpb.AuthServiceClient
}

type httpAuthClient struct {
	baseURL    string
	httpClient *http.Client
	idToken    func(ctx context.Context) (string, error)
	jwtSecret  string
}

// NewAuthClient establishes an HTTP (Cloud Run) or gRPC (local Docker) client connection to the Auth service.
func NewAuthClient(target string, opts ...AuthClientOption) (AuthClient, error) {
	var o authClientOptions
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}

	target = strings.TrimRight(strings.TrimSpace(target), "/")
	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		return &httpAuthClient{
			baseURL:    target,
			httpClient: &http.Client{Timeout: defaultAuthTimeout},
			idToken:    o.idToken,
			jwtSecret:  o.jwtSecret,
		}, nil
	}

	if target == "" {
		target = "localhost:50051"
	}
	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth service: %w", err)
	}
	return &authClient{
		conn:   conn,
		client: authpb.NewAuthServiceClient(conn),
	}, nil
}

// NewAuthClientWithService creates an AuthClient using an existing AuthServiceClient (useful for testing).
func NewAuthClientWithService(client authpb.AuthServiceClient) AuthClient {
	return &authClient{client: client}
}

func (c *authClient) GetUserDeviceTokens(ctx context.Context, userID string) ([]string, error) {
	if userID == "" {
		return nil, errors.New("userID cannot be empty")
	}

	callCtx, cancel := context.WithTimeout(ctx, defaultAuthTimeout)
	defer cancel()

	resp, err := c.client.GetUserDeviceTokens(callCtx, &authpb.GetUserDeviceTokensRequest{UserId: userID})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve user device tokens: %w", err)
	}
	return resp.Tokens, nil
}

func (c *authClient) RevokeDeviceTokens(ctx context.Context, userID string, tokens []string) error {
	if userID == "" || len(tokens) == 0 {
		return nil
	}

	callCtx, cancel := context.WithTimeout(ctx, defaultAuthTimeout)
	defer cancel()

	_, err := c.client.RevokeDeviceTokens(callCtx, &authpb.RevokeDeviceTokensRequest{
		UserId: userID,
		Tokens: tokens,
	})
	if err != nil {
		return fmt.Errorf("failed to revoke device tokens: %w", err)
	}
	return nil
}

func (c *authClient) GetCampaignRecipients(ctx context.Context) ([]RecipientContact, error) {
	callCtx, cancel := context.WithTimeout(ctx, defaultAuthTimeout)
	defer cancel()

	resp, err := c.client.GetCampaignRecipients(callCtx, &authpb.GetCampaignRecipientsRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve campaign recipients: %w", err)
	}

	out := make([]RecipientContact, 0, len(resp.Recipients))
	for _, r := range resp.Recipients {
		out = append(out, RecipientContact{
			UserID:          r.UserId,
			Email:           r.Email,
			FirstName:       r.FirstName,
			LastName:        r.LastName,
			IsEmailVerified: r.IsEmailVerified,
		})
	}
	return out, nil
}

func (c *authClient) IsUserEmailVerified(ctx context.Context, userID string) (bool, error) {
	if strings.TrimSpace(userID) == "" {
		return false, errors.New("user_id cannot be empty")
	}
	callCtx, cancel := context.WithTimeout(ctx, defaultAuthTimeout)
	defer cancel()

	resp, err := c.client.CheckUserEmailVerified(callCtx, &authpb.CheckUserEmailVerifiedRequest{UserId: userID})
	if err != nil {
		return false, fmt.Errorf("failed to check email verification: %w", err)
	}
	return resp.IsVerified, nil
}

func (c *authClient) GetUserContact(ctx context.Context, userID string) (*RecipientContact, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, errors.New("user_id cannot be empty")
	}
	callCtx, cancel := context.WithTimeout(ctx, defaultAuthTimeout)
	defer cancel()

	resp, err := c.client.GetUserContact(callCtx, &authpb.GetUserContactRequest{UserId: userID})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve user contact: %w", err)
	}
	return &RecipientContact{
		UserID:          resp.UserId,
		Email:           resp.Email,
		FirstName:       resp.FirstName,
		LastName:        resp.LastName,
		IsEmailVerified: resp.IsEmailVerified,
	}, nil
}

func (c *authClient) GetActiveUserIDsByRoles(ctx context.Context, roleIDs []string) ([]string, error) {
	if len(roleIDs) == 0 {
		return []string{}, nil
	}
	callCtx, cancel := context.WithTimeout(ctx, defaultAuthTimeout)
	defer cancel()

	resp, err := c.client.GetActiveUserIDsByRoles(callCtx, &authpb.GetActiveUserIDsByRolesRequest{RoleIds: roleIDs})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve user IDs by roles: %w", err)
	}
	return resp.UserIds, nil
}

func (c *authClient) GetParticipantUserID(ctx context.Context, participantID string) (string, error) {
	if strings.TrimSpace(participantID) == "" {
		return "", errors.New("participant_id cannot be empty")
	}
	callCtx, cancel := context.WithTimeout(ctx, defaultAuthTimeout)
	defer cancel()

	resp, err := c.client.GetParticipant(callCtx, &authpb.GetParticipantRequest{ParticipantId: participantID})
	if err != nil {
		return "", fmt.Errorf("failed to retrieve participant: %w", err)
	}
	return resp.UserId, nil
}

func (c *authClient) IssueActionLink(ctx context.Context, requestID, userID, actionType string) (string, error) {
	callCtx, cancel := context.WithTimeout(ctx, defaultAuthTimeout)
	defer cancel()

	resp, err := c.client.IssueActionLink(callCtx, &authpb.IssueActionLinkRequest{
		RequestId: requestID,
		UserId:    userID,
		Type:      actionType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to issue action link: %w", err)
	}
	return resp.ActionUrl, nil
}

func (c *authClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (h *httpAuthClient) doRequest(ctx context.Context, method, path string, body any, out any) error {
	callCtx, cancel := context.WithTimeout(ctx, defaultAuthTimeout)
	defer cancel()

	var bodyReader *bytes.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(b)
	} else {
		bodyReader = bytes.NewReader(nil)
	}

	req, err := http.NewRequestWithContext(callCtx, method, h.baseURL+path, bodyReader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if h.idToken != nil {
		if tok, err := h.idToken(callCtx); err == nil && tok != "" {
			req.Header.Set("X-Serverless-Authorization", "Bearer "+tok)
		}
	}
	if h.jwtSecret != "" {
		if tok, err := signServiceJWT(h.jwtSecret, "notifications-service"); err == nil {
			req.Header.Set("Authorization", "Bearer "+tok)
		}
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("auth-service responded HTTP %d on %s", resp.StatusCode, path)
	}
	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return err
		}
	}
	return nil
}

func (h *httpAuthClient) GetUserDeviceTokens(ctx context.Context, userID string) ([]string, error) {
	if userID == "" {
		return nil, errors.New("userID cannot be empty")
	}
	var res authpb.GetUserDeviceTokensResponse
	if err := h.doRequest(ctx, http.MethodGet, internalUsersBasePath+url.PathEscape(userID)+"/device-tokens", nil, &res); err != nil {
		return nil, fmt.Errorf("failed to retrieve user device tokens: %w", err)
	}
	return res.Tokens, nil
}

func (h *httpAuthClient) RevokeDeviceTokens(ctx context.Context, userID string, tokens []string) error {
	if userID == "" || len(tokens) == 0 {
		return nil
	}
	path := internalUsersBasePath + url.PathEscape(userID) + "/device-tokens/revoke"
	if err := h.doRequest(ctx, http.MethodPost, path, map[string][]string{"tokens": tokens}, nil); err != nil {
		return fmt.Errorf("failed to revoke device tokens: %w", err)
	}
	return nil
}

func (h *httpAuthClient) GetCampaignRecipients(ctx context.Context) ([]RecipientContact, error) {
	var res authpb.GetCampaignRecipientsResponse
	if err := h.doRequest(ctx, http.MethodGet, "/api/v1/directory/internal/campaigns/recipients", nil, &res); err != nil {
		return nil, fmt.Errorf("failed to retrieve campaign recipients: %w", err)
	}
	out := make([]RecipientContact, 0, len(res.Recipients))
	for _, r := range res.Recipients {
		out = append(out, RecipientContact{
			UserID:          r.UserId,
			Email:           r.Email,
			FirstName:       r.FirstName,
			LastName:        r.LastName,
			IsEmailVerified: r.IsEmailVerified,
		})
	}
	return out, nil
}

func (h *httpAuthClient) IsUserEmailVerified(ctx context.Context, userID string) (bool, error) {
	if strings.TrimSpace(userID) == "" {
		return false, errors.New("user_id cannot be empty")
	}
	var res authpb.CheckUserEmailVerifiedResponse
	if err := h.doRequest(ctx, http.MethodGet, internalUsersBasePath+url.PathEscape(userID)+"/email-verified", nil, &res); err != nil {
		return false, fmt.Errorf("failed to check email verification: %w", err)
	}
	return res.IsVerified, nil
}

func (h *httpAuthClient) GetUserContact(ctx context.Context, userID string) (*RecipientContact, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, errors.New("user_id cannot be empty")
	}
	var res authpb.GetUserContactResponse
	if err := h.doRequest(ctx, http.MethodGet, internalUsersBasePath+url.PathEscape(userID)+"/contact", nil, &res); err != nil {
		return nil, fmt.Errorf("failed to retrieve user contact: %w", err)
	}
	return &RecipientContact{
		UserID:          res.UserId,
		Email:           res.Email,
		FirstName:       res.FirstName,
		LastName:        res.LastName,
		IsEmailVerified: res.IsEmailVerified,
	}, nil
}

func (h *httpAuthClient) GetActiveUserIDsByRoles(ctx context.Context, roleIDs []string) ([]string, error) {
	if len(roleIDs) == 0 {
		return []string{}, nil
	}
	var res authpb.GetActiveUserIDsByRolesResponse
	if err := h.doRequest(ctx, http.MethodPost, "/api/v1/directory/internal/users/active-by-roles", map[string][]string{"role_ids": roleIDs}, &res); err != nil {
		return nil, fmt.Errorf("failed to retrieve user IDs by roles: %w", err)
	}
	return res.UserIds, nil
}

func (h *httpAuthClient) GetParticipantUserID(ctx context.Context, participantID string) (string, error) {
	if strings.TrimSpace(participantID) == "" {
		return "", errors.New("participant_id cannot be empty")
	}
	var res authpb.GetParticipantResponse
	if err := h.doRequest(ctx, http.MethodGet, "/api/v1/directory/internal/participants/"+url.PathEscape(participantID), nil, &res); err != nil {
		return "", fmt.Errorf("failed to retrieve participant: %w", err)
	}
	return res.UserId, nil
}

func (h *httpAuthClient) IssueActionLink(ctx context.Context, requestID, userID, actionType string) (string, error) {
	var res authpb.IssueActionLinkResponse
	body := map[string]string{
		"request_id": requestID,
		"user_id":    userID,
		"type":       actionType,
	}
	if err := h.doRequest(ctx, http.MethodPost, "/api/v1/directory/internal/action-links", body, &res); err != nil {
		return "", fmt.Errorf("failed to issue action link: %w", err)
	}
	return res.ActionUrl, nil
}

func (h *httpAuthClient) Close() error {
	return nil
}

func signServiceJWT(secret, subject string) (string, error) {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	now := time.Now().Unix()
	payloadBytes, err := json.Marshal(map[string]any{
		"sub":  subject,
		"role": "role_dev",
		"type": "access",
		"iat":  now,
		"exp":  now + 60,
	})
	if err != nil {
		return "", err
	}
	payload := base64.RawURLEncoding.EncodeToString(payloadBytes)
	signingInput := header + "." + payload
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(signingInput))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return signingInput + "." + sig, nil
}
