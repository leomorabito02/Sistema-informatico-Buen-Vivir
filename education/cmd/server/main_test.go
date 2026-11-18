package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"buenvivir/education/internal/config"
	"buenvivir/education/internal/events/producers"
	"buenvivir/education/internal/outbox"
)

// TestSkillProviders_EquivalenceClasses tests AI provider initialization based on config.
func TestSkillProviders_EquivalenceClasses(t *testing.T) {
	t.Run("Neither provider enabled returns empty slice", func(t *testing.T) {
		// Arrange
		cfg := &config.Config{
			SkillsAI: config.SkillsAIConfig{},
		}

		// Act
		providers := skillProviders(cfg, nil)

		// Assert
		assert.Empty(t, providers)
	})

	t.Run("OpenAI enabled returns OpenAI provider", func(t *testing.T) {
		// Arrange
		cfg := &config.Config{
			SkillsAI: config.SkillsAIConfig{
				OpenAIAPIKey: "sk-test-key",
				OpenAIModel:  "gpt-4o-mini",
			},
		}

		// Act
		providers := skillProviders(cfg, nil)

		// Assert
		assert.Len(t, providers, 1)
		assert.Equal(t, "openai", providers[0].Nombre())
	})

	t.Run("Gemini enabled but invalid credentials logs error and continues", func(t *testing.T) {
		// Arrange
		cfg := &config.Config{
			SkillsAI: config.SkillsAIConfig{
				VertexProjectID: "proj",
				VertexLocation:  "us-central1",
				GeminiModel:     "gemini-1.5",
				OpenAIAPIKey:    "sk-test",
				OpenAIModel:     "gpt-4o",
			},
		}

		// Act
		providers := skillProviders(cfg, []byte("invalid-json"))

		// Assert
		assert.Len(t, providers, 1)
		assert.Equal(t, "openai", providers[0].Nombre())
	})
}

// TestStartRelay_DisabledReturnsNil tests safe no-op when PubSub publishing is disabled.
func TestStartRelay_DisabledReturnsNil(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		Events: config.EventsConfig{
			ProjectID: "",
		},
	}

	// Act
	relay := startRelay(context.Background(), cfg, nil, nil)

	// Assert
	assert.Nil(t, relay)
}

// TestAsyncEndpoints_DecisionTable tests async endpoint configuration matrix.
func TestAsyncEndpoints_DecisionTable(t *testing.T) {
	t.Run("Disabled returns empty", func(t *testing.T) {
		// Arrange
		cfg := &config.Config{
			Events: config.EventsConfig{
				ProjectID: "",
			},
		}

		// Act
		endpoints := asyncEndpoints(cfg, nil, nil, nil, nil, nil, nil)

		// Assert
		assert.Nil(t, endpoints.Usuarios)
		assert.Nil(t, endpoints.Contenido)
		assert.Nil(t, endpoints.ContentfulWebhook)
	})

	t.Run("Consumers enabled with push auth disabled configures endpoints", func(t *testing.T) {
		// Arrange
		cfg := &config.Config{
			AppEnv: config.AppEnvTesting,
			Events: config.EventsConfig{
				ProjectID:               "proj",
				ContentfulWebhookSecret: "secret",
				PushAuth:                "disabled",
			},
		}
		writer := outbox.NewWriter(nil, "producer-test", nil)
		producer := producers.NewEducacion(writer, "t1", "t2")
		relay := &outbox.Relay{}

		// Act
		endpoints := asyncEndpoints(cfg, nil, nil, nil, nil, producer, relay)

		// Assert
		assert.NotNil(t, endpoints.ContentfulWebhook)
		assert.NotNil(t, endpoints.Usuarios)
		assert.NotNil(t, endpoints.Contenido)
		assert.NotNil(t, endpoints.OutboxFlush)
	})

	t.Run("Webhook configured but producer disabled skips webhook", func(t *testing.T) {
		// Arrange
		cfg := &config.Config{
			Events: config.EventsConfig{
				ProjectID:               "",
				ContentfulWebhookSecret: "secret",
			},
		}
		producer := producers.NewEducacion(nil, "", "")

		// Act
		endpoints := asyncEndpoints(cfg, nil, nil, nil, nil, producer, nil)

		// Assert
		assert.Nil(t, endpoints.ContentfulWebhook)
		assert.Nil(t, endpoints.Usuarios)
		assert.Nil(t, endpoints.Contenido)
		assert.Nil(t, endpoints.OutboxFlush)
	})

	t.Run("Consumers disabled but webhook enabled configures only webhook", func(t *testing.T) {
		// Arrange
		cfg := &config.Config{
			Events: config.EventsConfig{
				ProjectID:               "",
				ContentfulWebhookSecret: "secret",
			},
		}
		writer := outbox.NewWriter(nil, "producer-test", nil)
		producer := producers.NewEducacion(writer, "t1", "t2")

		// Act
		endpoints := asyncEndpoints(cfg, nil, nil, nil, nil, producer, nil)

		// Assert
		assert.NotNil(t, endpoints.ContentfulWebhook)
		assert.Nil(t, endpoints.Usuarios)
		assert.Nil(t, endpoints.Contenido)
		assert.Nil(t, endpoints.OutboxFlush)
	})

	t.Run("Consumers enabled but nil relay leaves OutboxFlush nil", func(t *testing.T) {
		// Arrange
		cfg := &config.Config{
			AppEnv: config.AppEnvTesting,
			Events: config.EventsConfig{
				ProjectID: "proj",
				PushAuth:  "disabled",
			},
		}

		// Act
		endpoints := asyncEndpoints(cfg, nil, nil, nil, nil, nil, nil)

		// Assert
		assert.NotNil(t, endpoints.Usuarios)
		assert.NotNil(t, endpoints.Contenido)
		assert.Nil(t, endpoints.OutboxFlush)
	})
}
