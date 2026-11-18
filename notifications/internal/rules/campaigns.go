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

	if p.CampanaID == "" {
		return nil, nil
	}

	adminRole := "role_admin"
	partRole := "role_participant"
	volRole := "role_volunteer"
	staffActionURL := "/dashboard/campanas"
	publicActionURL := "/campanas-publico"

	return []NotificationPlan{
		{
			SourceService: "campaigns",
			EventType:     env.EventType,
			Category:      models.CategoryCampaigns,
			Priority:      models.PriorityNormal,
			AudienceType:  AudienceRole,
			TargetRole:    &adminRole,
			Title:         "Nueva campaña comunitaria",
			Message:       "Se ha registrado una nueva campaña comunitaria.",
			ActionURL:     &staffActionURL,
			SendInApp:     true,
			SendEmail:     true,
			SendPush:      true,
			DedupeKey:     fmt.Sprintf("%s_admin", env.EventID),
		},
		{
			SourceService: "campaigns",
			EventType:     env.EventType,
			Category:      models.CategoryCampaigns,
			Priority:      models.PriorityNormal,
			AudienceType:  AudienceRole,
			TargetRole:    &partRole,
			Title:         "Nueva campaña comunitaria",
			Message:       "Hay una nueva campaña disponible en Buen Vivir.",
			ActionURL:     &publicActionURL,
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
			ActionURL:     &staffActionURL,
			SendInApp:     true,
			SendEmail:     true,
			SendPush:      true,
			DedupeKey:     fmt.Sprintf("%s_volunteer", env.EventID),
		},
	}, nil
}

