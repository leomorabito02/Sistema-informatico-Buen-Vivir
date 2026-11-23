package repositories

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// maxIDsPorPedido: tope de ids por pedido a /directory/users/summaries.
const maxIDsPorPedido = 50

// GetUserNames devuelve user_id -> nombre completo (POST /api/v1/directory/users/summaries).
// Usa un token interno de servicio porque el participante no puede consultar el
// directorio. Solo pide nombres.
func (c *AuthServiceClient) GetUserNames(ctx context.Context, userIDs []string) (map[string]string, error) {
	if c == nil || c.baseURL == "" {
		return nil, ErrAuthServiceNotConfigured
	}
	out := make(map[string]string, len(userIDs))
	for inicio := 0; inicio < len(userIDs); inicio += maxIDsPorPedido {
		fin := inicio + maxIDsPorPedido
		if fin > len(userIDs) {
			fin = len(userIDs)
		}
		if err := c.pedirNombres(ctx, userIDs[inicio:fin], out); err != nil {
			return out, err
		}
	}
	return out, nil
}

func (c *AuthServiceClient) pedirNombres(ctx context.Context, ids []string, out map[string]string) error {
	token, err := tokenInternoServicio()
	if err != nil {
		return err
	}
	body, _ := json.Marshal(map[string][]string{"user_ids": ids})
	const path = "/api/v1/directory/users/summaries"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
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
	var res struct {
		Data []struct {
			UserID   string `json:"user_id"`
			FullName string `json:"full_name"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return fmt.Errorf("no se pudo parsear la respuesta de auth-service (%s): %w", path, err)
	}
	for _, u := range res.Data {
		if n := strings.TrimSpace(u.FullName); n != "" {
			out[u.UserID] = n
		}
	}
	return nil
}

// tokenInternoServicio firma un JWT de un minuto con el JWT_SECRET común,
// rol role_dev, como hace labor-integration para sus consultas internas.
func tokenInternoServicio() (string, error) {
	secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if secret == "" {
		return "", fmt.Errorf("JWT_SECRET no configurado")
	}
	ahora := time.Now()
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  "education-service",
		"role": "role_dev",
		"type": "access",
		"iat":  ahora.Unix(),
		"exp":  ahora.Add(time.Minute).Unix(),
	})
	return t.SignedString([]byte(secret))
}
