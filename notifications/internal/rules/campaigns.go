package rules

import (
	"context"
	"fmt"

	"buenvivir/notifications/internal/events"
	"buenvivir/notifications/internal/models"
)

// ── C1: CampanaCreadaRule ────────────────────────────────────────────────

type CampanaCreadaRule struct{}

func (r *CampanaCreadaRule) EventType() string {
	return events.CampanaCreada
}

func (r *CampanaCreadaRule) Build(_ context.Context, env events.Envelope) ([]NotificationPlan, error) {
	p, err := events.DecodePayload[events.CampanaCreadaPayload](env)
	if err != nil {
		return nil, err
	}

	partRole := "role_participant"
	volRole := "role_volunteer"
	actionURL := "/campaigns/" + p.CampanaID

	return []NotificationPlan{
		{
			SourceService: "campaigns",
			EventType:     env.EventType,
			Category:      models.CategoryCampaigns,
			Priority:      models.PriorityNormal,
			AudienceType:  AudienceRole,
			TargetRole:    &partRole,
			Title:         "Nueva campaña comunitaria",
			Message:       "Hay una nueva campaña disponible en Buen Vivir.",
			ActionURL:     &actionURL,
			SendInApp:     true,
			SendEmail:     true,
			SendPush:      true,
			DedupeKey:     fmt.Sprintf("%s_participant", env.EventID),
		},
		{
			SourceService: "campaigns",
			EventType:     env.EventType,
			Category:      models.CategoryCampaigns,
			Priority:      models.PriorityNormal,
			AudienceType:  AudienceRole,
			TargetRole:    &volRole,
			Title:         "Nueva campaña comunitaria",
			Message:       "Hay una nueva campaña comunitaria disponible.",
			ActionURL:     &actionURL,
			SendInApp:     true,
			SendEmail:     true,
			SendPush:      true,
			DedupeKey:     fmt.Sprintf("%s_volunteer", env.EventID),
		},
	}, nil
}

