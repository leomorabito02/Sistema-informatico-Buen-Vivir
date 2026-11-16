package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// GrantedSkill es una habilidad ganada por un participante, tal como la
// devuelve Educación en GET /api/habilidades/otorgadas.
type GrantedSkill struct {
	ParticipanteID string `json:"participanteId"`
	Nombre         string `json:"nombre"`
}

// SkillsClient consulta a Educación (dueña de las habilidades) para el
// cruce de ofertas compatibles (HU-44).
type SkillsClient interface {
	GrantedSkills(ctx context.Context, bearer string) ([]GrantedSkill, error)
}

type httpSkillsClient struct {
	baseURL string
	http    *http.Client
}

// NewSkillsClient arma el cliente contra EDUCATION_API_URL (la URL del
// gateway o de Educación). Devuelve nil si no hay URL: entonces el cruce
// de compatibilidad queda desactivado.
func NewSkillsClient(baseURL string) SkillsClient {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return nil
	}
	return &httpSkillsClient{baseURL: baseURL, http: &http.Client{Timeout: 5 * time.Second}}
}

// GrantedSkills reenvía el token del administrador que aprobó la oferta:
// el endpoint es solo para admin y así no hace falta otra credencial.
func (c *httpSkillsClient) GrantedSkills(ctx context.Context, bearer string) ([]GrantedSkill, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/habilidades/otorgadas", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+bearer)
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
