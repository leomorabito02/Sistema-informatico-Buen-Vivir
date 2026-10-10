package repositories

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"cloud.google.com/go/auth"
	"cloud.google.com/go/auth/credentials"

	"buenvivir/education/internal/models"
)

const vertexScope = "https://www.googleapis.com/auth/cloud-platform"

// GeminiClient pide sugerencias de habilidades a Gemini a través de
// Vertex AI (proveedor principal, HU-58).
//
// Se habla con la API REST de Vertex directamente, sin el SDK de
// Gemini, para no sumar una dependencia nueva: alcanza con un POST y la
// autenticación de Google que el servicio ya trae (la misma cuenta de
// servicio que usa Firestore). Por eso Gemini no necesita API key.
type GeminiClient struct {
	httpClient *http.Client
	endpoint   string
	// token devuelve un token de acceso de Google. Es una función para
	// poder reemplazarla en los tests y no llamar nunca a Google de verdad.
	token func(ctx context.Context) (string, error)
}

// NewGeminiClient arma el cliente de Vertex AI. credentialsJSON es la
// service account inline (ver GoogleCredentialsJSON); si es nil se usan
// las credenciales por defecto del entorno (en Cloud Run, la cuenta de
// servicio del propio servicio).
func NewGeminiClient(projectID, location, model string, credentialsJSON []byte) (*GeminiClient, error) {
	opts := &credentials.DetectOptions{Scopes: []string{vertexScope}}
	var (
		creds *auth.Credentials
		err   error
	)
	if len(credentialsJSON) > 0 {
		creds, err = credentials.NewCredentialsFromJSON(credentials.ServiceAccount, credentialsJSON, opts)
	} else {
		creds, err = credentials.DetectDefault(opts)
	}
	if err != nil {
		return nil, fmt.Errorf("no se pudieron cargar las credenciales de Google para Vertex AI: %w", err)
	}

	return &GeminiClient{
		httpClient: &http.Client{},
		endpoint:   vertexEndpoint(projectID, location, model),
		token: func(ctx context.Context) (string, error) {
			t, err := creds.Token(ctx)
			if err != nil {
				return "", err
			}
			return t.Value, nil
		},
	}, nil
}

// vertexEndpoint arma la URL de generateContent. La ubicación "global"
// usa el host sin región.
func vertexEndpoint(projectID, location, model string) string {
	host := location + "-aiplatform.googleapis.com"
	if location == "global" {
		host = "aiplatform.googleapis.com"
	}
	return fmt.Sprintf("https://%s/v1/projects/%s/locations/%s/publishers/google/models/%s:generateContent", host, projectID, location, model)
}

// Nombre identifica al proveedor en la respuesta y en los logs.
func (g *GeminiClient) Nombre() string { return "gemini" }

// Sugerir manda el pedido y devuelve las habilidades propuestas.
func (g *GeminiClient) Sugerir(ctx context.Context, p models.SkillPrompt) ([]models.SkillSuggestion, error) {
	token, err := g.token(ctx)
	if err != nil {
		return nil, fmt.Errorf("gemini: no se pudo obtener el token de Google: %w", err)
	}

	body := map[string]any{
		"systemInstruction": map[string]any{
			"parts": []map[string]string{{"text": skillsSystemInstruction(p.MaxSugerencias)}},
		},
		"contents": []map[string]any{{
			"role":  "user",
			"parts": []map[string]string{{"text": skillsUserPrompt(p)}},
		}},
		// Structured Output: responseMimeType + responseSchema obligan a
		// Gemini a devolver JSON con exactamente la forma del esquema.
		// Temperatura baja: queremos respuestas estables, no creativas.
		// thinkingBudget 0 apaga el "razonamiento" previo de Gemini 2.5
		// Flash: para nombrar habilidades no hace falta, y esos tokens se
		// cobran como salida y hacen más lenta la respuesta.
		"generationConfig": map[string]any{
			"responseMimeType": "application/json",
			"responseSchema":   skillsResponseSchema(false),
			"temperature":      0.2,
			"maxOutputTokens":  1024,
			"thinkingConfig":   map[string]any{"thinkingBudget": 0},
		},
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gemini: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("gemini: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, &AIHTTPError{Provider: "gemini", StatusCode: resp.StatusCode}
	}

	var out struct {
		Candidates []struct {
			FinishReason string `json:"finishReason"`
			Content      struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrAIInvalidResponse, err)
	}
	if len(out.Candidates) == 0 || len(out.Candidates[0].Content.Parts) == 0 {
		// Pasa, por ejemplo, si Gemini bloqueó la respuesta por sus
		// filtros de seguridad: no hay texto que leer.
		return nil, fmt.Errorf("%w: gemini no devolvió contenido", ErrAIInvalidResponse)
	}
	var text bytes.Buffer
	for _, part := range out.Candidates[0].Content.Parts {
		text.WriteString(part.Text)
	}
	return parseSkillsJSON(text.String())
}
