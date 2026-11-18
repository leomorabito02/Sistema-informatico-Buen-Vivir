package config

import "testing"

func TestLoadEventsConfig(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("PUBSUB_PROJECT_ID", "ceibuenvivir")
	t.Setenv("PUBSUB_TOPIC_PROMOCION_HUMANA", "")
	t.Setenv("PUBSUB_PUSH_AUDIENCE", "https://hp.run.app")
	t.Setenv("PUBSUB_PUSH_SERVICE_ACCOUNTS", " a@x , ,b@x")
	t.Setenv("PUBSUB_PUSH_AUTH", "")

	e := LoadConfig().Events
	if !e.Production || !e.PublishEnabled() || !e.ConsumeEnabled() {
		t.Errorf("config: %+v", e)
	}
	if e.TopicPromocion != "promocion-humana-eventos" || len(e.PushServiceAccounts) != 2 || e.PushServiceAccounts[1] != "b@x" {
		t.Errorf("config: %+v", e)
	}

	t.Setenv("PUBSUB_PROJECT_ID", "")
	t.Setenv("PUBSUB_PUSH_AUDIENCE", "")
	t.Setenv("PUBSUB_TOPIC_PROMOCION_HUMANA", "otro")
	e = LoadConfig().Events
	if e.PublishEnabled() || e.ConsumeEnabled() || e.TopicPromocion != "otro" {
		t.Errorf("sin Pub/Sub: %+v", e)
	}
	e.PushAuth = "disabled"
	if !e.ConsumeEnabled() {
		t.Error("local con el emulador")
	}
}
