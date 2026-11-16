package rules

import (
	"context"
	"encoding/json"
	"fmt"

	"buenvivir/notifications/internal/events"
	"buenvivir/notifications/internal/models"
)

// ── U1: UsuarioCreadoRule ────────────────────────────────────────────────

type UsuarioCreadoRule struct{}

func (r *UsuarioCreadoRule) EventType() string {
	return events.UsuarioCreado
}

func (r *UsuarioCreadoRule) Build(_ context.Context, env events.Envelope) ([]NotificationPlan, error) {
	p, err := events.DecodePayload[events.UsuarioCreadoPayload](env)
	if err != nil {
		return nil, err
	}

	if p.UserID == "" {
		return nil, nil
	}

	// Inicializa las preferencias del usuario creando un aviso de bienvenida
	actionURL := "/dashboard/mi-perfil"
	plans := []NotificationPlan{
		{
			SourceService:    "auth",
			EventType:        env.EventType,
			Category:         models.CategorySystem,
			Priority:         models.PriorityLow,
			AudienceType:     AudienceIndividual,
			RecipientUserIDs: []string{p.UserID},
			Title:            "¡Bienvenido a Buen Vivir!",
			Message:          "Tu cuenta ha sido creada exitosamente. Explora los cursos y servicios disponibles.",
			ActionURL:        &actionURL,
			SendInApp:        true,
			DedupeKey:        fmt.Sprintf("%s_welcome_%s", env.EventID, p.UserID),
		},
	}

	if p.RoleID != "" && p.RoleID != "role_admin" && p.RoleID != "role_dev" {
		adminRole := "role_admin"
		adminURL := "/dashboard/usuarios"
		plans = append(plans, NotificationPlan{
			SourceService: "auth",
			EventType:     env.EventType,
			Category:      models.CategorySystem,
			Priority:      models.PriorityNormal,
			AudienceType:  AudienceRole,
			TargetRole:    &adminRole,
			Title:         "Nuevo usuario pendiente de revisión",
			Message:       "Se ha registrado una nueva cuenta que requiere aprobación en la plataforma.",
			ActionURL:     &adminURL,
			SendInApp:     true,
			SendEmail:     true,
			DedupeKey:     fmt.Sprintf("%s_admin_review", env.EventID),
		})
	}

	return plans, nil
}

// ── U2: EmailSolicitadoRule ──────────────────────────────────────────────

type EmailSolicitadoRule struct{}

func (r *EmailSolicitadoRule) EventType() string {
	return events.EmailSolicitado
}

func (r *EmailSolicitadoRule) Build(_ context.Context, env events.Envelope) ([]NotificationPlan, error) {
	p, err := events.DecodePayload[events.EmailSolicitadoPayload](env)
	if err != nil {
		return nil, err
	}

	if p.UserID == "" {
		return nil, nil
	}

	var title, message, templateID string
	category := models.CategorySecurity
	isCritical := true

	switch p.Tipo {
	case "verificacion":
		title = "Verifica tu correo electrónico - Buen Vivir"
		message = "Por favor verifica tu correo electrónico para activar tu cuenta."
		templateID = "email_verification"
	case "cuenta_aprobada":
		title = "¡Tu cuenta ha sido aprobada! - Buen Vivir"
		message = "Tu cuenta ha sido aprobada exitosamente."
		templateID = "account_approved"
		category = models.CategorySystem
		isCritical = false
	case "cuenta_rechazada":
		// HU-12 y HU-38: aviso de rechazo. Sin enlace de acción, así
		// que no se manda request_id (no hay link que pedirle a Auth).
		title = "Tu solicitud de cuenta no fue aprobada - Buen Vivir"
		message = "El equipo del centro revisó tu solicitud de cuenta y no fue aprobada. Si creés que es un error, comunicate con el Centro Barrial La Otra Orilla."
		templateID = "general_alert"
		category = models.CategorySystem
		isCritical = false
	case "recuperacion_password":
		title = "Recuperación de contraseña - Buen Vivir"
		message = "Recibimos una solicitud para restablecer tu contraseña."
		templateID = "password_recovery"
	case "two_factor_code":
		title = "Tu código de verificación de seguridad - Buen Vivir"
		message = "Utilizá este código para completar el inicio de sesión."
		templateID = "two_factor_code"
	default:
		title = "Aviso de seguridad - Buen Vivir"
		message = "Notificación de seguridad de tu cuenta."
		templateID = p.Tipo
	}

	data := map[string]interface{}{
		"template_id": templateID,
		"request_id":  p.RequestID,
		"user_id":     p.UserID,
		"tipo":        p.Tipo,
	}
	if p.Tipo == "cuenta_rechazada" {
		delete(data, "request_id")
	}
	dataBytes, _ := json.Marshal(data)

	return []NotificationPlan{
		{
			SourceService:    "auth",
			EventType:        env.EventType,
			Category:         category,
			Priority:         models.PriorityCritical,
			AudienceType:     AudienceIndividual,
			RecipientUserIDs: []string{p.UserID},
			Title:            title,
			Message:          message,
			Data:             string(dataBytes),
			SendEmail:        true,
			IsCritical:       isCritical,
			DedupeKey:        fmt.Sprintf("%s_sec_%s", env.EventID, p.RequestID),
		},
	}, nil
}

