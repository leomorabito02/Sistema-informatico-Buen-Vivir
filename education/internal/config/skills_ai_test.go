package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoadSkillsAIConfig_Defaults(t *testing.T) {
	for _, k := range []string{"VERTEX_PROJECT_ID", "VERTEX_LOCATION", "GEMINI_MODEL", "OPENAI_API_KEY", "OPENAI_MODEL", "SKILLS_AI_TIMEOUT_SECONDS", "SKILLS_AI_MAX_RETRIES"} {
		t.Setenv(k, "")
	}
	c := loadSkillsAIConfig()
	assert.False(t, c.GeminiEnabled())
	assert.False(t, c.OpenAIEnabled())
	assert.Equal(t, "us-central1", c.VertexLocation)
	assert.Equal(t, "gemini-2.5-flash", c.GeminiModel)
	assert.Equal(t, 20*time.Second, c.Timeout)
	assert.Equal(t, 1, c.MaxRetries)
}

func TestLoadSkillsAIConfig_FromEnvWithLimits(t *testing.T) {
	t.Setenv("VERTEX_PROJECT_ID", "ceibuenvivir")
	t.Setenv("OPENAI_API_KEY", "sk-test")
	t.Setenv("GEMINI_MODEL", "otro-modelo")
	t.Setenv("SKILLS_AI_TIMEOUT_SECONDS", "600")
	t.Setenv("SKILLS_AI_MAX_RETRIES", "9")
	c := loadSkillsAIConfig()
	assert.True(t, c.GeminiEnabled())
	assert.True(t, c.OpenAIEnabled())
	assert.Equal(t, "otro-modelo", c.GeminiModel)
	assert.Equal(t, 60*time.Second, c.Timeout, "el timeout tiene un tope")
	assert.Equal(t, 3, c.MaxRetries, "los reintentos tienen un tope")
}
