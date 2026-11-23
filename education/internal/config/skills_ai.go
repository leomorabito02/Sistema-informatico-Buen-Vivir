package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Valores por defecto de la sugerencia de habilidades con IA. Se pueden
// cambiar por variable de entorno sin tocar código (ver docs/ia-habilidades.md).
const (
	defaultVertexLocation   = "us-central1"
	defaultGeminiModel      = "gemini-2.5-flash"
	defaultOpenAIModel      = "gpt-4o-mini"
	defaultSkillsAITimeout  = 20 * time.Second
	defaultSkillsAIRetries  = 1
	maxSkillsAIRetries      = 3
	maxSkillsAITimeoutLimit = 60 * time.Second
)

// SkillsAIConfig agrupa la configuración de los proveedores de IA: Gemini (Vertex AI,
// sin API key) y OpenAI (respaldo, con API key desde Secret Manager).
type SkillsAIConfig struct {
	VertexProjectID string
	VertexLocation  string
	GeminiModel     string

	OpenAIAPIKey string
	OpenAIModel  string

	// Timeout es el tiempo máximo de CADA intento contra un proveedor.
	Timeout time.Duration
	// MaxRetries es cuántas veces se reintenta un proveedor ante una
	// falla pasajera (timeout, saturación, error 5xx) antes de pasar al
	// siguiente. Acotado a 3 para que el administrador no espere minutos.
	MaxRetries int
}

// GeminiEnabled indica si Gemini está configurado.
func (c SkillsAIConfig) GeminiEnabled() bool { return c.VertexProjectID != "" }

// OpenAIEnabled indica si OpenAI está configurado.
func (c SkillsAIConfig) OpenAIEnabled() bool { return c.OpenAIAPIKey != "" }

func loadSkillsAIConfig() SkillsAIConfig {
	c := SkillsAIConfig{
		VertexProjectID: strings.TrimSpace(os.Getenv("VERTEX_PROJECT_ID")),
		VertexLocation:  envOr("VERTEX_LOCATION", defaultVertexLocation),
		GeminiModel:     envOr("GEMINI_MODEL", defaultGeminiModel),
		OpenAIAPIKey:    strings.TrimSpace(os.Getenv("OPENAI_API_KEY")),
		OpenAIModel:     envOr("OPENAI_MODEL", defaultOpenAIModel),
		Timeout:         defaultSkillsAITimeout,
		MaxRetries:      defaultSkillsAIRetries,
	}
	if secs, err := strconv.Atoi(strings.TrimSpace(os.Getenv("SKILLS_AI_TIMEOUT_SECONDS"))); err == nil && secs > 0 {
		c.Timeout = min(time.Duration(secs)*time.Second, maxSkillsAITimeoutLimit)
	}
	if n, err := strconv.Atoi(strings.TrimSpace(os.Getenv("SKILLS_AI_MAX_RETRIES"))); err == nil && n >= 0 {
		c.MaxRetries = min(n, maxSkillsAIRetries)
	}
	return c
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
