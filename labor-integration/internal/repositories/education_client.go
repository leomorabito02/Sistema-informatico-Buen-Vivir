package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"buenvivir/labor-integration/api/proto/educationpb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const defaultEducationCallTimeout = 5 * time.Second

// GrantedSkill es una habilidad ganada por un participante, tal como la
// devuelve Educación en gRPC (GetGrantedSkills) o GET /api/habilidades/otorgadas.
type GrantedSkill struct {
	ParticipanteID string `json:"participanteId"`
	Nombre         string `json:"nombre"`
}

// SkillsClient consulta a Educación (dueña de las habilidades) para el
// cruce de ofertas compatibles (HU-44).
type SkillsClient interface {
	GrantedSkills(ctx context.Context, bearer string) ([]GrantedSkill, error)
}

type grpcSkillsClient struct {
	conn   *grpc.ClientConn
	client educationpb.EducationServiceClient
}

type httpSkillsClient struct {
	baseURL string
	http    *http.Client
	idToken func(ctx context.Context) (string, error)
}

// NewSkillsClient arma el cliente gRPC (Docker local) o HTTP (Cloud Run)
// contra Educación. Devuelve nil si target está vacío.
func NewSkillsClient(target string, idTokenFn ...func(ctx context.Context) (string, error)) SkillsClient {
	target = strings.TrimRight(strings.TrimSpace(target), "/")
	if target == "" {
		return nil
	}
	var idToken func(ctx context.Context) (string, error)
	if len(idTokenFn) > 0 {
		idToken = idTokenFn[0]
	}
	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		return &httpSkillsClient{
			baseURL: target,
			http:    &http.Client{Timeout: defaultEducationCallTimeout},
			idToken: idToken,
		}
	}

	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil
	}
	return &grpcSkillsClient{
		conn:   conn,
		client: educationpb.NewEducationServiceClient(conn),
	}
}

func (c *grpcSkillsClient) GrantedSkills(ctx context.Context, _ string) ([]GrantedSkill, error) {
	callCtx, cancel := context.WithTimeout(ctx, defaultEducationCallTimeout)
	defer cancel()

	resp, err := c.client.GetGrantedSkills(callCtx, &educationpb.GetGrantedSkillsRequest{})
	if err != nil {
		return nil, fmt.Errorf("educación (gRPC) falló: %w", err)
	}
	out := make([]GrantedSkill, 0, len(resp.GetSkills()))
	for _, s := range resp.GetSkills() {
		out = append(out, GrantedSkill{
			ParticipanteID: s.GetParticipanteId(),
			Nombre:         s.GetNombre(),
		})
	}
	return out, nil
}

// GrantedSkills reenvía el token del administrador que aprobó la oferta y,
// en Cloud Run, adjunta X-Serverless-Authorization con el ID Token de GCP.
func (c *httpSkillsClient) GrantedSkills(ctx context.Context, bearer string) ([]GrantedSkill, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/habilidades/otorgadas", nil)
	if err != nil {
		return nil, err
	}
	if c.idToken != nil {
		if tok, err := c.idToken(ctx); err == nil && tok != "" {
			req.Header.Set("X-Serverless-Authorization", "Bearer "+tok)
		}
	}
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("educación respondió %d", resp.StatusCode)
	}
	var out []GrantedSkill
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}
