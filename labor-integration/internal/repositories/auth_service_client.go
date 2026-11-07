package repositories

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"buenvivir/labor-integration/api/proto/authpb"
)

const defaultAuthCallTimeout = 5 * time.Second

type AuthServiceClient interface {
	GetParticipant(ctx context.Context, participantID string) (*authpb.GetParticipantResponse, error)
	GetCompany(ctx context.Context, companyID string) (*authpb.GetCompanyResponse, error)
	GetActiveUserIDsByRoles(ctx context.Context, roles []string) ([]string, error)
	Close() error
}

type authServiceOptions struct {
	idToken   func(ctx context.Context) (string, error)
	jwtSecret string
}

type AuthServiceClientOption func(*authServiceOptions)

func WithAuthIDToken(fn func(ctx context.Context) (string, error)) AuthServiceClientOption {
	return func(o *authServiceOptions) {
		o.idToken = fn
	}
}

func WithAuthJWTSecret(secret string) AuthServiceClientOption {
	return func(o *authServiceOptions) {
		o.jwtSecret = strings.TrimSpace(secret)
	}
}

type authServiceClient struct {
	conn   *grpc.ClientConn
	client authpb.AuthServiceClient
}

type httpAuthServiceClient struct {
	baseURL    string
	httpClient *http.Client
	idToken    func(ctx context.Context) (string, error)
	jwtSecret  string
}

func NewAuthServiceClient(target string, opts ...AuthServiceClientOption) (AuthServiceClient, error) {
	var o authServiceOptions
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}

	target = strings.TrimRight(strings.TrimSpace(target), "/")
	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		return &httpAuthServiceClient{
			baseURL:    target,
			httpClient: &http.Client{Timeout: defaultAuthCallTimeout},
			idToken:    o.idToken,
			jwtSecret:  o.jwtSecret,
		}, nil
	}

	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth-service: %w", err)
	}

	return &authServiceClient{
		conn:   conn,
		client: authpb.NewAuthServiceClient(conn),
	}, nil
}

func (c *authServiceClient) GetParticipant(ctx context.Context, participantID string) (*authpb.GetParticipantResponse, error) {
	req := &authpb.GetParticipantRequest{
		ParticipantId: participantID,
	}
	return c.client.GetParticipant(ctx, req)
}

func (c *authServiceClient) GetCompany(ctx context.Context, companyID string) (*authpb.GetCompanyResponse, error) {
	req := &authpb.GetCompanyRequest{
		CompanyId: companyID,
	}
	return c.client.GetCompany(ctx, req)
}

func (c *authServiceClient) GetActiveUserIDsByRoles(ctx context.Context, roles []string) ([]string, error) {
	req := &authpb.GetActiveUserIDsByRolesRequest{
		RoleIds: roles,
	}
	resp, err := c.client.GetActiveUserIDsByRoles(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.UserIds, nil
}

func (c *authServiceClient) Close() error {
	return c.conn.Close()
}

func (h *httpAuthServiceClient) doRequest(ctx context.Context, method, path string, body any, out any) (int, error) {
	callCtx, cancel := context.WithTimeout(ctx, defaultAuthCallTimeout)
	defer cancel()

	var bodyReader *bytes.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return 0, err
		}
		bodyReader = bytes.NewReader(b)
	} else {
		bodyReader = bytes.NewReader(nil)
	}

	req, err := http.NewRequestWithContext(callCtx, method, h.baseURL+path, bodyReader)
	if err != nil {
		return 0, err
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
		if tok, err := signInternalJWT(h.jwtSecret, "labor-integration-service"); err == nil {
			req.Header.Set("Authorization", "Bearer "+tok)
		}
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("error calling auth-service: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, fmt.Errorf("auth-service responded HTTP %d", resp.StatusCode)
	}
	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return resp.StatusCode, err
		}
	}
	return resp.StatusCode, nil
}

func (h *httpAuthServiceClient) GetParticipant(ctx context.Context, participantID string) (*authpb.GetParticipantResponse, error) {
	var res authpb.GetParticipantResponse
	_, err := h.doRequest(ctx, http.MethodGet, "/api/v1/directory/internal/participants/"+url.PathEscape(participantID), nil, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (h *httpAuthServiceClient) GetCompany(ctx context.Context, companyID string) (*authpb.GetCompanyResponse, error) {
	var res authpb.GetCompanyResponse
	_, err := h.doRequest(ctx, http.MethodGet, "/api/v1/directory/internal/companies/"+url.PathEscape(companyID), nil, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (h *httpAuthServiceClient) GetActiveUserIDsByRoles(ctx context.Context, roles []string) ([]string, error) {
	var res authpb.GetActiveUserIDsByRolesResponse
	_, err := h.doRequest(ctx, http.MethodPost, "/api/v1/directory/internal/users/active-by-roles", map[string][]string{"role_ids": roles}, &res)
	if err != nil {
		return nil, err
	}
	return res.UserIds, nil
}

func (h *httpAuthServiceClient) Close() error {
	return nil
}

func signInternalJWT(secret, subject string) (string, error) {
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
