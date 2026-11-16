package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"buenvivir/education/api/proto/authpb"
	"buenvivir/education/internal/middlewares"
	"buenvivir/education/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const authClientTimeout = 10 * time.Second

// AuthServiceClient habla con auth-service por gRPC (Docker local) o HTTP
// (Cloud Run) para resolver datos de participantes y usuarios que viven
// en la base de auth.
type AuthServiceClient struct {
	baseURL    string
	httpClient *http.Client
	idToken    func(ctx context.Context) (string, error)
	conn       *grpc.ClientConn
	grpcClient authpb.AuthServiceClient
}

// ErrAuthServiceNotConfigured se devuelve cuando no hay target HTTP ni gRPC configurado.
var ErrAuthServiceNotConfigured = errors.New("AUTH_SERVICE_URL no está configurado")

func NewAuthServiceClient(target string, idTokenFn ...func(ctx context.Context) (string, error)) *AuthServiceClient {
	var idToken func(ctx context.Context) (string, error)
	if len(idTokenFn) > 0 {
		idToken = idTokenFn[0]
	}
	target = strings.TrimRight(strings.TrimSpace(target), "/")
	if target == "" {
		return &AuthServiceClient{
			httpClient: &http.Client{Timeout: authClientTimeout},
			idToken:    idToken,
		}
	}
	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		return &AuthServiceClient{
			baseURL:    target,
			httpClient: &http.Client{Timeout: authClientTimeout},
			idToken:    idToken,
		}
	}

	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return &AuthServiceClient{
			httpClient: &http.Client{Timeout: authClientTimeout},
			idToken:    idToken,
		}
	}
	return &AuthServiceClient{
		httpClient: &http.Client{Timeout: authClientTimeout},
		idToken:    idToken,
		conn:       conn,
		grpcClient: authpb.NewAuthServiceClient(conn),
	}
}

func (c *AuthServiceClient) Close() error {
	if c != nil && c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// authListResponse replica el sobre {success, data: [...], errors} con
// el que responde auth-service para los listados.
type authListResponse[T any] struct {
	Success bool `json:"success"`
	Data    []T  `json:"data"`
	Errors  any  `json:"errors"`
}

// authItemResponse es el mismo sobre, pero para un solo recurso.
type authItemResponse[T any] struct {
	Success bool `json:"success"`
	Data    T    `json:"data"`
	Errors  any  `json:"errors"`
}

// ListParticipants obtiene el listado de participantes desde auth-service.
func (c *AuthServiceClient) ListParticipants(ctx context.Context, bearerToken string) ([]models.AuthParticipant, error) {
	if c == nil || (c.baseURL == "" && c.grpcClient == nil) {
		return nil, ErrAuthServiceNotConfigured
	}
	if c.grpcClient != nil {
		callCtx, cancel := context.WithTimeout(ctx, authClientTimeout)
		defer cancel()
		resp, err := c.grpcClient.ListParticipants(callCtx, &authpb.ListParticipantsRequest{})
		if err != nil {
			return nil, fmt.Errorf("error gRPC hablando con auth-service (ListParticipants): %w", err)
		}
		out := make([]models.AuthParticipant, 0, len(resp.GetParticipants()))
		for _, p := range resp.GetParticipants() {
			out = append(out, models.AuthParticipant{
				ParticipantID:  p.GetId(),
				UserID:         p.GetUserId(),
				DNI:            p.GetDni(),
				ApprovalStatus: p.GetApprovalStatus(),
			})
		}
		return out, nil
	}

	var parsed authListResponse[models.AuthParticipant]
	if err := c.doGet(ctx, "/api/v1/participants", bearerToken, &parsed); err != nil {
		return nil, err
	}
	return parsed.Data, nil
}

// GetParticipant obtiene un participante por ID. Devuelve (nil, nil) si no existe (404 / NotFound).
func (c *AuthServiceClient) GetParticipant(ctx context.Context, participantID, bearerToken string) (*models.AuthParticipant, error) {
	if c == nil || (c.baseURL == "" && c.grpcClient == nil) {
		return nil, ErrAuthServiceNotConfigured
	}
	if c.grpcClient != nil {
		callCtx, cancel := context.WithTimeout(ctx, authClientTimeout)
		defer cancel()
		resp, err := c.grpcClient.GetParticipant(callCtx, &authpb.GetParticipantRequest{ParticipantId: participantID})
		if err != nil {
			if status.Code(err) == codes.NotFound {
				return nil, nil
			}
			return nil, fmt.Errorf("error gRPC hablando con auth-service (GetParticipant): %w", err)
		}
		return &models.AuthParticipant{
			ParticipantID:  resp.GetId(),
			UserID:         resp.GetUserId(),
			DNI:            resp.GetDni(),
			ApprovalStatus: resp.GetApprovalStatus(),
		}, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/participants/"+participantID, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+bearerToken)
	if c.idToken != nil {
		if tok, err := c.idToken(ctx); err == nil && tok != "" {
			req.Header.Set("X-Serverless-Authorization", "Bearer "+tok)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error de red hablando con auth-service: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("auth-service respondió %d al buscar el participante", resp.StatusCode)
	}

	var parsed authItemResponse[models.AuthParticipant]
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("no se pudo parsear la respuesta de auth-service: %w", err)
	}
	return &parsed.Data, nil
}

// ListUsers obtiene el listado de usuarios (nombre/apellido) desde auth-service.
func (c *AuthServiceClient) ListUsers(ctx context.Context, bearerToken string) ([]models.AuthUser, error) {
	if c == nil || (c.baseURL == "" && c.grpcClient == nil) {
		return nil, ErrAuthServiceNotConfigured
	}
	if c.grpcClient != nil {
		callCtx, cancel := context.WithTimeout(ctx, authClientTimeout)
		defer cancel()
		resp, err := c.grpcClient.ListUsers(callCtx, &authpb.ListUsersRequest{})
		if err != nil {
			return nil, fmt.Errorf("error gRPC hablando con auth-service (ListUsers): %w", err)
		}
		out := make([]models.AuthUser, 0, len(resp.GetUsers()))
		for _, u := range resp.GetUsers() {
			out = append(out, models.AuthUser{
				UserID:    u.GetUserId(),
				FirstName: u.GetFirstName(),
				LastName:  u.GetLastName(),
			})
		}
		return out, nil
	}

	var parsed authListResponse[models.AuthUser]
	if err := c.doGet(ctx, "/api/v1/users", bearerToken, &parsed); err != nil {
		return nil, err
	}
	return parsed.Data, nil
}

// GetParticipantName obtiene el nombre y apellido de un participante sin DNI ni datos de contacto.
func (c *AuthServiceClient) GetParticipantName(ctx context.Context, participantID, bearerToken string) (models.AuthUser, error) {
	if c == nil || (c.baseURL == "" && c.grpcClient == nil) {
		return models.AuthUser{}, ErrAuthServiceNotConfigured
	}
	if c.grpcClient != nil {
		callCtx, cancel := context.WithTimeout(ctx, authClientTimeout)
		defer cancel()
		resp, err := c.grpcClient.GetParticipantProfile(callCtx, &authpb.GetParticipantProfileRequest{
			Lookup: &authpb.GetParticipantProfileRequest_ParticipantId{ParticipantId: participantID},
		})
		if err != nil {
			return models.AuthUser{}, fmt.Errorf("error gRPC hablando con auth-service (GetParticipantProfile): %w", err)
		}
		return models.AuthUser{FirstName: resp.GetFirstName(), LastName: resp.GetLastName()}, nil
	}

	var parsed authItemResponse[struct {
		Nombre   string `json:"nombre"`
		Apellido string `json:"apellido"`
	}]
	if err := c.doGet(ctx, "/api/v1/directory/participants/"+url.PathEscape(participantID), bearerToken, &parsed); err != nil {
		return models.AuthUser{}, err
	}
	return models.AuthUser{FirstName: parsed.Data.Nombre, LastName: parsed.Data.Apellido}, nil
}

// GetMyParticipantID devuelve el participant_id asociado al usuario autenticado.
func (c *AuthServiceClient) GetMyParticipantID(ctx context.Context, bearerToken string) (string, error) {
	if c == nil || (c.baseURL == "" && c.grpcClient == nil) {
		return "", ErrAuthServiceNotConfigured
	}
	if c.grpcClient != nil {
		userID := extractUserID(ctx, bearerToken)
		if userID == "" {
			return "", errors.New("missing user_id in token")
		}
		callCtx, cancel := context.WithTimeout(ctx, authClientTimeout)
		defer cancel()
		resp, err := c.grpcClient.GetParticipantProfile(callCtx, &authpb.GetParticipantProfileRequest{
			Lookup: &authpb.GetParticipantProfileRequest_UserId{UserId: userID},
		})
		if err != nil {
			if status.Code(err) == codes.NotFound {
				return "", nil
			}
			return "", fmt.Errorf("error gRPC hablando con auth-service (GetParticipantProfile by user_id): %w", err)
		}
		return resp.GetParticipantId(), nil
	}

	var parsed struct {
		RoleData struct {
			ParticipantID string `json:"participant_id"`
		} `json:"role_data"`
	}
	if err := c.doGet(ctx, "/api/v1/profile/me", bearerToken, &parsed); err != nil {
		return "", err
	}
	return parsed.RoleData.ParticipantID, nil
}

func extractUserID(ctx context.Context, bearerToken string) string {
	if uid, ok := ctx.Value(middlewares.UserIDKey).(string); ok && strings.TrimSpace(uid) != "" {
		return strings.TrimSpace(uid)
	}
	var claims jwt.MapClaims
	if _, _, err := jwt.NewParser().ParseUnverified(strings.TrimSpace(bearerToken), &claims); err == nil {
		if sub, ok := claims["sub"].(string); ok && strings.TrimSpace(sub) != "" {
			return strings.TrimSpace(sub)
		}
	}
	return strings.TrimSpace(bearerToken)
}

func (c *AuthServiceClient) doGet(ctx context.Context, path, bearerToken string, out any) error {
	if c == nil || c.baseURL == "" {
		return ErrAuthServiceNotConfigured
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+bearerToken)
	if c.idToken != nil {
		if tok, err := c.idToken(ctx); err == nil && tok != "" {
			req.Header.Set("X-Serverless-Authorization", "Bearer "+tok)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error de red hablando con auth-service (%s): %w", path, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("auth-service respondió %d en %s", resp.StatusCode, path)
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("no se pudo parsear la respuesta de auth-service (%s): %w", path, err)
	}
	return nil
}
