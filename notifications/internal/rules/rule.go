package rules

import (
	"context"

	"buenvivir/notifications/internal/events"
)

// Rule maps an incoming domain event Envelope into one or more NotificationPlans.
type Rule interface {
	EventType() string
	Build(ctx context.Context, env events.Envelope) ([]NotificationPlan, error)
}

