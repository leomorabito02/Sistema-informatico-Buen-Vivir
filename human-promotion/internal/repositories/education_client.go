package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"buenvivir/human-promotion/api/proto/educationpb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const educationCallTimeout = 5 * time.Second

// SkillDirectory devuelve las habilidades que ganó un participante, pidiéndolas a
// education-service (este servicio no guarda copia).
type SkillDirectory interface {
	HabilidadesDe(ctx context.Context, participanteID string) ([]string, error)
}

type tokenContextKey struct{}

// ContextWithToken guarda el JWT del usuario que hizo el pedido, para
// reenviarlo a education-service.
func ContextWithToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, tokenContextKey{}, token)
}

func tokenFromContext(ctx context.Context) string {
	token, _ := ctx.Value(tokenContextKey{}).(string)
	return token
}

type grpcEducationClient struct {
	conn   *grpc.ClientConn
	client educationpb.EducationServiceClient
}

type educationClient struct {
	baseURL    string
	httpClient *http.Client
	idToken    func(ctx context.Context) (string, error)
}

// NewEducationClient devuelve el cliente gRPC (Docker local) o HTTP (Cloud Run)
// de education-service, o nil si target está vacío.
func NewEducationClient(target string, idTokenFn ...func(ctx context.Context) (string, error)) SkillDirectory {
	target = strings.TrimRight(strings.TrimSpace(target), "/")
	if target == "" {
		return nil
	}
	var idToken func(ctx context.Context) (string, error)
	if len(idTokenFn) > 0 {
		idToken = idTokenFn[0]
	}
	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		return &educationClient{
			baseURL:    target,
			httpClient: &http.Client{Timeout: educationCallTimeout},
			idToken:    idToken,
		}
	}

	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil
	}
	return &grpcEducationClient{
		conn:   conn,
		client: educationpb.NewEducationServiceClient(conn),
	}
}

func (c *grpcEducationClient) HabilidadesDe(ctx context.Context, participanteID string) ([]string, error) {
	callCtx, cancel := context.WithTimeout(ctx, educationCallTimeout)
	defer cancel()

	resp, err := c.client.GetParticipantSkills(callCtx, &educationpb.GetParticipantSkillsRequest{
		ParticipanteId: participanteID,
	})
	if err != nil {
		return nil, fmt.Errorf("education-service (gRPC) no respondió: %w", err)
	}
	nombres := make([]string, 0, len(resp.GetNombres()))
	for _, n := range resp.GetNombres() {
		if n != "" {
			nombres = append(nombres, n)
		}
	}
	return nombres, nil
}

func (c *educationClient) HabilidadesDe(ctx context.Context, participanteID string) ([]string, error) {
	endpoint := fmt.Sprintf("%s/api/participants/%s/habilidades", c.baseURL, url.PathEscape(participanteID))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	if c.idToken != nil {
		if tok, err := c.idToken(ctx); err == nil && tok != "" {
			req.Header.Set("X-Serverless-Authorization", "Bearer "+tok)
		}
	}
	if token := tokenFromContext(ctx); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("education-service no respondió: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("education-service respondió HTTP %d", resp.StatusCode)
	}

	var habilidades []struct {
		Nombre string `json:"nombre"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&habilidades); err != nil {
		return nil, fmt.Errorf("respuesta inválida de education-service: %w", err)
	}
	nombres := make([]string, 0, len(habilidades))
	for _, h := range habilidades {
		if h.Nombre != "" {
			nombres = append(nombres, h.Nombre)
		}
	}
	return nombres, nil
}
