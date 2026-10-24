package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoadEventsConfig(t *testing.T) {
	t.Setenv("PUBSUB_PROJECT_ID", " ceibuenvivir ")
	t.Setenv("PUBSUB_TOPIC_EDUCACION", "")
	t.Setenv("PUBSUB_PUSH_SERVICE_ACCOUNTS", "a@x.iam.gserviceaccount.com, ,b@x.iam.gserviceaccount.com")
	t.Setenv("PUBSUB_PUSH_AUDIENCE", "https://education.run.app")
	t.Setenv("CONTENT_CACHE_TTL", "90s")

	e := loadEventsConfig()
	assert.Equal(t, "ceibuenvivir", e.ProjectID)
	assert.Equal(t, "educacion-eventos", e.TopicEducacion)
	assert.Equal(t, "contenido-eventos", e.TopicContenido)
	assert.Equal(t, []string{"a@x.iam.gserviceaccount.com", "b@x.iam.gserviceaccount.com"}, e.PushServiceAccounts)
	assert.Equal(t, 90*time.Second, e.ContentCacheTTL)
	assert.True(t, e.PublishEnabled())
	assert.True(t, e.ConsumeEnabled())
}

func TestLoadEventsConfig_PorDefecto(t *testing.T) {
	t.Setenv("PUBSUB_PROJECT_ID", "")
	t.Setenv("PUBSUB_PUSH_AUDIENCE", "")
	t.Setenv("PUBSUB_PUSH_SERVICE_ACCOUNTS", "")
	t.Setenv("PUBSUB_PUSH_AUTH", "")
	t.Setenv("CONTENT_CACHE_TTL", "no-es-duracion")

	e := loadEventsConfig()
	assert.False(t, e.PublishEnabled())
	assert.False(t, e.ConsumeEnabled())
	assert.Equal(t, 5*time.Minute, e.ContentCacheTTL)

	e.PushAuth = "disabled"
	assert.True(t, e.ConsumeEnabled(), "local con el emulador")
}

func TestConsumeEnabled_AudienceWithoutServiceAccounts(t *testing.T) {
	e := EventsConfig{
		PushAuth:            "",
		PushAudience:        "https://education.run.app",
		PushServiceAccounts: []string{},
	}
	assert.False(t, e.ConsumeEnabled(), "audience without service accounts must not enable consume")
}

func TestLoadEventsConfig_ContentCacheTTL_ZeroBoundary(t *testing.T) {
	t.Setenv("CONTENT_CACHE_TTL", "0s")
	e := loadEventsConfig()
	assert.Equal(t, 0*time.Second, e.ContentCacheTTL)
}

