package repositories

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"buenvivir/education/internal/models"
)

const openAIDefaultBaseURL = "https://api.openai.com/v1"

// OpenAIClient pide sugerencias de habilidades a OpenAI. Es el
// proveedor de RESPALDO (HU-21 / HU-58): solo se usa si Gemini falla.
//
// Igual que con Gemini, se usa la API REST directamente (sin SDK) para
// no sumar dependencias. La API key llega por variable de entorno desde
// Secret Manager y nunca se escribe en logs ni en respuestas.
type OpenAIClient struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	model      string
}

func NewOpenAIClient(apiKey, model string) *OpenAIClient {
	return &OpenAIClient{httpClient: &http.Client{}, baseURL: openAIDefaultBaseURL, apiKey: apiKey, model: model}
}

// Nombre identifica al proveedor en la respuesta y en los logs.
func (o *OpenAIClient) Nombre() string { return "openai" }

// Sugerir manda el pedido y devuelve las habilidades propuestas.
func (o *OpenAIClient) Sugerir(ctx context.Context, p models.SkillPrompt) ([]models.SkillSuggestion, error) {
	body := map[string]any{
		"model": o.model,
		"messages": []map[string]string{
			{"role": "system", "content": skillsSystemInstruction(p.MaxSugerencias)},
			{"role": "user", "content": skillsUserPrompt(p)},
		},
		// Structured Output de OpenAI: con "strict" el modelo solo puede
		// devolver JSON que cumpla el esquema.
		"response_format": map[string]any{
			"type": "json_schema",
			"json_schema": map[string]any{
				"name":   "habilidades_sugeridas",
				"strict": true,
				"schema": skillsResponseSchema(true),
			},
		},
		"temperature": 0.2,
		"max_tokens":  1024,
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.baseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+o.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openai: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("openai: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, &AIHTTPError{Provider: "openai", StatusCode: resp.StatusCode}
	}

	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
				// Refusal viene cargado si el modelo se negó a responder.
				Refusal string `json:"refusal"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrAIInvalidResponse, err)
	}
	if len(out.Choices) == 0 || out.Choices[0].Message.Content == "" {
		return nil, fmt.Errorf("%w: openai no devolvió contenido", ErrAIInvalidResponse)
	}
	return parseSkillsJSON(out.Choices[0].Message.Content)
}
