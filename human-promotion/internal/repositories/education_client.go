package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// SkillDirectory devuelve las habilidades que ganó un participante. Las
// habilidades son datos de Educación (HU-21 / HU-58): este servicio no
// tiene copia propia y las pide a education-service cada vez que arma el
// detalle de una ficha, respetando que cada servicio tiene su propia base.
type SkillDirectory interface {
	HabilidadesDe(ctx context.Context, participanteID string) ([]string, error)
}

type tokenContextKey struct{}

// ContextWithToken guarda el JWT del usuario que hizo el pedido, para
// reenviarlo a education-service. Así Educación aplica su propio control
// de permisos sobre la misma persona (solo admin y voluntario pueden ver
// habilidades de otro participante, igual que las fichas) y no hace falta
// un secreto compartido entre servicios. Lo llama AuthMiddleware.
func ContextWithToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, tokenContextKey{}, token)
}

func tokenFromContext(ctx context.Context) string {
	token, _ := ctx.Value(tokenContextKey{}).(string)
	return token
}

type educationClient struct {
	baseURL    string
	httpClient *http.Client
	idToken    func(ctx context.Context) (string, error)
}

// NewEducationClient devuelve el cliente de education-service, o nil si
// EDUCATION_SERVICE_URL no está configurada: en ese caso la ficha se
// muestra sin habilidades, como antes de esta funcionalidad.
func NewEducationClient(baseURL string, idTokenFn ...func(ctx context.Context) (string, error)) SkillDirectory {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return nil
	}
	var idToken func(ctx context.Context) (string, error)
	if len(idTokenFn) > 0 {
		idToken = idTokenFn[0]
	}
	// Timeout corto: las habilidades son un dato secundario de la ficha y
	// no pueden demorar la apertura del contenido confidencial.
	return &educationClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
		idToken:    idToken,
	}
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
