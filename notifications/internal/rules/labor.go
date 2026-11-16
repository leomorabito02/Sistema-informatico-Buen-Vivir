package rules

import (
	"context"
	"fmt"
	"strings"

	"buenvivir/notifications/internal/events"
	"buenvivir/notifications/internal/models"
)

// ── L1: OfertaCreadaRule ─────────────────────────────────────────────────

type OfertaCreadaRule struct{}

func (r *OfertaCreadaRule) EventType() string {
	return events.OfertaCreada
}

func (r *OfertaCreadaRule) Build(_ context.Context, env events.Envelope) ([]NotificationPlan, error) {
	p, err := events.DecodePayload[events.OfertaCreadaPayload](env)
	if err != nil {
		return nil, err
	}

	adminRole := "role_admin"
	volRole := "role_volunteer"
	actionURL := "/dashboard/ofertas"

	var message string
	if strings.EqualFold(p.Estado, "activa") {
		message = "Se ha publicado una nueva oferta laboral."
	} else {
		message = "Se ha registrado una nueva oferta laboral pendiente de revisión."
	}

	return []NotificationPlan{
		{
			SourceService: "labor-integration",
			EventType:     env.EventType,
			Category:      models.CategoryLabor,
			Priority:      models.PriorityNormal,
			AudienceType:  AudienceRole,
			TargetRole:    &adminRole,
			Title:         "Nueva oferta laboral registrada",
			Message:       message,
			ActionURL:     &actionURL,
			SendInApp:     true,
			SendEmail:     true,
			SendPush:      true,
			DedupeKey:     fmt.Sprintf("%s_admin", env.EventID),
		},
		{
			SourceService: "labor-integration",
			EventType:     env.EventType,
			Category:      models.CategoryLabor,
			Priority:      models.PriorityNormal,
			AudienceType:  AudienceRole,
			TargetRole:    &volRole,
			Title:         "Nueva oferta laboral registrada",
			Message:       message,
			ActionURL:     &actionURL,
			SendInApp:     true,
			SendEmail:     true,
			SendPush:      true,
			DedupeKey:     fmt.Sprintf("%s_volunteer", env.EventID),
		},
	}, nil
}

// ── L2: OfertaRevisadaRule ────────────────────────────────────────────────

type OfertaRevisadaRule struct{}

func (r *OfertaRevisadaRule) EventType() string {
	return events.OfertaRevisada
}

func (r *OfertaRevisadaRule) Build(_ context.Context, env events.Envelope) ([]NotificationPlan, error) {
	p, err := events.DecodePayload[events.OfertaRevisadaPayload](env)
	if err != nil {
		return nil, err
	}

	if p.EmpresaID == "" {
		return nil, nil
	}

	actionURL := "/dashboard/ofertas"
	var title, message string
	if p.Aprobada {
		title = "Oferta laboral aprobada"
		message = "Tu oferta laboral ha sido aprobada y ya se encuentra activa."
	} else {
		title = "Oferta laboral rechazada"
		message = "Tu oferta laboral ha sido rechazada."
	}

	return []NotificationPlan{
		{
			SourceService:    "labor-integration",
			EventType:        env.EventType,
			Category:         models.CategoryLabor,
			Priority:         models.PriorityNormal,
			AudienceType:     AudienceIndividual,
			RecipientUserIDs: []string{p.EmpresaID},
			Title:            title,
			Message:          message,
			ActionURL:        &actionURL,
			SendInApp:        true,
			SendEmail:        true,
			SendPush:         true,
			DedupeKey:        fmt.Sprintf("%s_company_%s", env.EventID, p.EmpresaID),
		},
	}, nil
}

// ── L3: PostulacionCreadaRule ─────────────────────────────────────────────

type PostulacionCreadaRule struct{}

func (r *PostulacionCreadaRule) EventType() string {
	return events.PostulacionCreada
}

func (r *PostulacionCreadaRule) Build(_ context.Context, env events.Envelope) ([]NotificationPlan, error) {
	p, err := events.DecodePayload[events.PostulacionCreadaPayload](env)
	if err != nil {
		return nil, err
	}

	actionURL := "/dashboard/ofertas"
	var plans []NotificationPlan

	if p.EmpresaID != "" {
		plans = append(plans, NotificationPlan{
			SourceService:    "labor-integration",
			EventType:        env.EventType,
			Category:         models.CategoryLabor,
			Priority:         models.PriorityNormal,
			AudienceType:     AudienceIndividual,
			RecipientUserIDs: []string{p.EmpresaID},
			Title:            "Nueva postulación recibida",
			Message:          "Un participante se postuló a tu oferta laboral.",
			ActionURL:        &actionURL,
			SendInApp:        true,
			SendEmail:        true,
			SendPush:         true,
			DedupeKey:        fmt.Sprintf("%s_company_%s", env.EventID, p.EmpresaID),
		})
	}

	if p.ParticipanteID != "" {
		plans = append(plans, NotificationPlan{
			SourceService: "labor-integration",
			EventType:     env.EventType,
			Category:      models.CategoryLabor,
			Priority:      models.PriorityNormal,
			AudienceType:  AudienceParticipant,
			ParticipantID: p.ParticipanteID,
			Title:         "Postulación registrada",
			Message:       "Te has postulado con éxito a la oferta laboral.",
			ActionURL:     &actionURL,
			SendInApp:     true,
			SendEmail:     true,
			SendPush:      true,
			DedupeKey:     fmt.Sprintf("%s_part_%s", env.EventID, p.ParticipanteID),
		})
	}

	return plans, nil
}

// ── L4: PostulacionDecididaRule ───────────────────────────────────────────

type PostulacionDecididaRule struct{}

func (r *PostulacionDecididaRule) EventType() string {
	return events.PostulacionDecidida
}

func (r *PostulacionDecididaRule) Build(_ context.Context, env events.Envelope) ([]NotificationPlan, error) {
	p, err := events.DecodePayload[events.PostulacionDecididaPayload](env)
	if err != nil {
		return nil, err
	}

	actionURL := "/dashboard/ofertas"
	lower := strings.ToLower(strings.TrimSpace(p.Resultado))
	isAccepted := lower == "aceptado" || lower == "aceptada" || lower == "accepted" || lower == "aprobada" || lower == "aprobado" || lower == "seleccionado" || lower == "contratado"

	var plans []NotificationPlan

	// 1. Notificación al participante
	if p.ParticipanteID != "" {
		var title, message string
		if isAccepted {
			title = "Postulación aceptada"
			message = "¡Felicitaciones! Tu postulación para la oferta laboral ha sido aceptada."
		} else {
			title = "Actualización de postulación"
			message = "Tu postulación para la oferta laboral no ha sido seleccionada en esta oportunidad."
		}

		plans = append(plans, NotificationPlan{
			SourceService: "labor-integration",
			EventType:     env.EventType,
			Category:      models.CategoryLabor,
			Priority:      models.PriorityNormal,
			AudienceType:  AudienceParticipant,
			ParticipantID: p.ParticipanteID,
			Title:         title,
			Message:       message,
			ActionURL:     &actionURL,
			SendInApp:     true,
			SendEmail:     true,
			SendPush:      true,
			DedupeKey:     fmt.Sprintf("%s_part_%s", env.EventID, p.ParticipanteID),
		})
	}

	// 2. Notificación al staff
	adminRole := "role_admin"
	volRole := "role_volunteer"
	staffTitle := "Resolución de postulación laboral"
	staffMsg := "Se ha resuelto una postulación en una oferta laboral."

	plans = append(plans,
		NotificationPlan{
			SourceService: "labor-integration",
			EventType:     env.EventType,
			Category:      models.CategoryLabor,
			Priority:      models.PriorityNormal,
			AudienceType:  AudienceRole,
			TargetRole:    &adminRole,
			Title:         staffTitle,
			Message:       staffMsg,
			ActionURL:     &actionURL,
			SendInApp:     true,
			DedupeKey:     fmt.Sprintf("%s_admin", env.EventID),
		},
		NotificationPlan{
			SourceService: "labor-integration",
			EventType:     env.EventType,
			Category:      models.CategoryLabor,
			Priority:      models.PriorityNormal,
			AudienceType:  AudienceRole,
			TargetRole:    &volRole,
			Title:         staffTitle,
			Message:       staffMsg,
			ActionURL:     &actionURL,
			SendInApp:     true,
			DedupeKey:     fmt.Sprintf("%s_volunteer", env.EventID),
		},
	)

	return plans, nil
}


// ── L5: OfertaEliminadaRule (HU-41) ───────────────────────────────────────

type OfertaEliminadaRule struct{}

func (r *OfertaEliminadaRule) EventType() string {
	return events.OfertaEliminada
}

// Build avisa a la empresa cuando la gestión del centro borra su oferta.
// Si la borró la propia empresa no hay aviso.
func (r *OfertaEliminadaRule) Build(_ context.Context, env events.Envelope) ([]NotificationPlan, error) {
	p, err := events.DecodePayload[events.OfertaEliminadaPayload](env)
	if err != nil {
		return nil, err
	}
	if p.EmpresaID == "" || !p.PorGestion {
		return nil, nil
	}
	actionURL := "/dashboard/ofertas"
	return []NotificationPlan{{
		SourceService:    "labor-integration",
		EventType:        env.EventType,
		Category:         models.CategoryLabor,
		Priority:         models.PriorityNormal,
		AudienceType:     AudienceIndividual,
		RecipientUserIDs: []string{p.EmpresaID},
		Title:            "Oferta laboral eliminada",
		Message:          "El equipo del centro eliminó una de tus ofertas laborales.",
		ActionURL:        &actionURL,
		SendInApp:        true,
		SendEmail:        true,
		SendPush:         true,
		DedupeKey:        fmt.Sprintf("%s_company_%s", env.EventID, p.EmpresaID),
	}}, nil
}

// ── L6: OfertaCompatibleRule (HU-44) ──────────────────────────────────────

type OfertaCompatibleRule struct{}

func (r *OfertaCompatibleRule) EventType() string {
	return events.OfertaCompatible
}

// Build avisa a cada participante con habilidades que pide la oferta.
func (r *OfertaCompatibleRule) Build(_ context.Context, env events.Envelope) ([]NotificationPlan, error) {
	p, err := events.DecodePayload[events.OfertaCompatiblePayload](env)
	if err != nil {
		return nil, err
	}
	actionURL := "/dashboard/ofertas"
	plans := make([]NotificationPlan, 0, len(p.ParticipanteIDs))
	seen := map[string]bool{}
	for _, id := range p.ParticipanteIDs {
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		plans = append(plans, NotificationPlan{
			SourceService: "labor-integration",
			EventType:     env.EventType,
			Category:      models.CategoryLabor,
			Priority:      models.PriorityNormal,
			AudienceType:  AudienceParticipant,
			ParticipantID: id,
			Title:         "Hay una oferta laboral para vos",
			Message:       "Se publicó una oferta laboral que pide habilidades que ya tenés. Mirala en Ofertas.",
			ActionURL:     &actionURL,
			SendInApp:     true,
			SendEmail:     true,
			SendPush:      true,
			DedupeKey:     fmt.Sprintf("%s_part_%s", env.EventID, id),
		})
	}
	return plans, nil
}
