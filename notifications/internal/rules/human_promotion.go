package rules

import (
	"context"
	"fmt"
	"net/url"

	"buenvivir/notifications/internal/events"
	"buenvivir/notifications/internal/models"
)

// fichaURL es la pantalla de la ficha de un participante en el frontend (solo admin y
// voluntario).
func fichaURL(participanteID string) string {
	if participanteID == "" {
		return "/dashboard/promocion-humana"
	}
	return "/dashboard/promocion-humana/participantes/" + url.PathEscape(participanteID)
}

// ── P1: FichaCreadaRule ───────────────────────────────────────────────────

type FichaCreadaRule struct{}

func (r *FichaCreadaRule) EventType() string {
	return events.FichaCreada
}

func (r *FichaCreadaRule) Build(_ context.Context, env events.Envelope) ([]NotificationPlan, error) {
	p, err := events.DecodePayload[events.FichaPayload](env)
	if err != nil {
		return nil, err
	}

	adminRole := "role_admin"
	actionURL := fichaURL(p.ParticipanteID)
	return []NotificationPlan{
		{
			SourceService: "human-promotion",
			EventType:     env.EventType,
			Category:      models.CategoryHumanPromotion,
			Priority:      models.PriorityLow,
			AudienceType:  AudienceRole,
			TargetRole:    &adminRole,
			Title:         "Nueva ficha de proceso personal",
			Message:       "Se ha creado una nueva ficha de acompañamiento personal en el centro.",
			ActionURL:     &actionURL,
			SendInApp:     true,
			DedupeKey:     fmt.Sprintf("%s_admin", env.EventID),
		},
	}, nil
}

// ── P2: FichaActualizadaRule ───────────────────────────────────────────────

type FichaActualizadaRule struct{}

func (r *FichaActualizadaRule) EventType() string {
	return events.FichaActualizada
}

func (r *FichaActualizadaRule) Build(_ context.Context, env events.Envelope) ([]NotificationPlan, error) {
	p, err := events.DecodePayload[events.FichaPayload](env)
	if err != nil {
		return nil, err
	}

	// Si fue actualizada por alguien distinto del creador original, avisarle al creador
	if p.CreadoPor != "" && env.ActorID != "" && p.CreadoPor != env.ActorID {
		actionURL := fichaURL(p.ParticipanteID)
		return []NotificationPlan{
			{
				SourceService:    "human-promotion",
				EventType:        env.EventType,
				Category:         models.CategoryHumanPromotion,
				Priority:         models.PriorityLow,
				AudienceType:     AudienceIndividual,
				RecipientUserIDs: []string{p.CreadoPor},
				Title:            "Actualización en ficha de seguimiento",
				Message:          "Un miembro del equipo ha realizado modificaciones en una ficha iniciada por ti.",
				ActionURL:        &actionURL,
				SendInApp:        true,
				DedupeKey:        fmt.Sprintf("%s_creator_%s", env.EventID, p.CreadoPor),
			},
		}, nil
	}

	return nil, nil
}

// ── P3 & P4: HitoRegistradoRule ────────────────────────────────────────────

type HitoRegistradoRule struct{}

func (r *HitoRegistradoRule) EventType() string {
	return events.HitoRegistrado
}

func (r *HitoRegistradoRule) Build(_ context.Context, env events.Envelope) ([]NotificationPlan, error) {
	p, err := events.DecodePayload[events.HitoPayload](env)
	if err != nil {
		return nil, err
	}

	plans := make([]NotificationPlan, 0, 2)

	// P3: Notificación al participante si el hito está marcado como visible
	if p.VisibleParticipante {
		actionURL := "/dashboard/mi-proceso"
		plans = append(plans, NotificationPlan{
			SourceService: "human-promotion",
			EventType:     env.EventType,
			Category:      models.CategoryHumanPromotion,
			Priority:      models.PriorityNormal,
			AudienceType:  AudienceParticipant,
			ParticipantID: p.ParticipanteID,
			Title:         "Nuevo hito en tu proceso",
			Message:       "Se ha registrado un nuevo logro o avance en tu línea de tiempo.",
			ActionURL:     &actionURL,
			SendPush:      true,
			SendInApp:     true,
			DedupeKey:     fmt.Sprintf("%s_participant_%s", env.EventID, p.ParticipanteID),
		})
	}

	// P4: Notificación al voluntario responsable asignado (si es distinto del actor que lo creó)
	if p.ResponsableID != "" && p.ResponsableID != env.ActorID {
		actionURL := fichaURL(p.ParticipanteID)
		plans = append(plans, NotificationPlan{
			SourceService:    "human-promotion",
			EventType:        env.EventType,
			Category:         models.CategoryHumanPromotion,
			Priority:         models.PriorityNormal,
			AudienceType:     AudienceIndividual,
			RecipientUserIDs: []string{p.ResponsableID},
			Title:            "Asignación de hito en ficha",
			Message:          "Has sido asignado como responsable de un nuevo hito de seguimiento.",
			ActionURL:        &actionURL,
			SendEmail:        true,
			SendInApp:        true,
			DedupeKey:        fmt.Sprintf("%s_responsible_%s", env.EventID, p.ResponsableID),
		})
	}

	return plans, nil
}

// ── P5: HitoActualizadoRule ────────────────────────────────────────────────

type HitoActualizadoRule struct{}

func (r *HitoActualizadoRule) EventType() string {
	return events.HitoActualizado
}

func (r *HitoActualizadoRule) Build(_ context.Context, env events.Envelope) ([]NotificationPlan, error) {
	p, err := events.DecodePayload[events.HitoPayload](env)
	if err != nil {
		return nil, err
	}

	if p.VisibleParticipante {
		actionURL := "/dashboard/mi-proceso"
		return []NotificationPlan{
			{
				SourceService: "human-promotion",
				EventType:     env.EventType,
				Category:      models.CategoryHumanPromotion,
				Priority:      models.PriorityLow,
				AudienceType:  AudienceParticipant,
				ParticipantID: p.ParticipanteID,
				Title:         "Actualización en tu línea de tiempo",
				Message:       "Se ha actualizado información sobre un avance en tu proceso.",
				ActionURL:     &actionURL,
				SendInApp:     true,
				DedupeKey:     fmt.Sprintf("%s_participant_%s", env.EventID, p.ParticipanteID),
			},
		}, nil
	}

	return nil, nil
}

// ── P6: ParticipanteEliminadoRule ──────────────────────────────────────────

type ParticipanteEliminadoRule struct{}

func (r *ParticipanteEliminadoRule) EventType() string {
	return events.ParticipanteEliminado
}

func (r *ParticipanteEliminadoRule) Build(_ context.Context, env events.Envelope) ([]NotificationPlan, error) {
	adminRole := "role_admin"
	return []NotificationPlan{
		{
			SourceService: "human-promotion",
			EventType:     env.EventType,
			Category:      models.CategoryHumanPromotion,
			Priority:      models.PriorityNormal,
			AudienceType:  AudienceRole,
			TargetRole:    &adminRole,
			Title:         "Baja de participante registrada",
			Message:       "Un participante ha sido dado de baja. Sus fichas e historial pasan al estado baja de forma automática.",
			SendInApp:     true,
			DedupeKey:     fmt.Sprintf("%s_admin", env.EventID),
		},
	}, nil
}

