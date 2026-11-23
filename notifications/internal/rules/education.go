package rules

import (
	"context"
	"fmt"
	"strings"

	"buenvivir/notifications/internal/events"
	"buenvivir/notifications/internal/models"
)

// ── E1 & E2: InscripcionCreadaRule ──────────────────────────────────────────

type InscripcionCreadaRule struct{}

func (r *InscripcionCreadaRule) EventType() string {
	return events.InscripcionCreada
}

func (r *InscripcionCreadaRule) Build(_ context.Context, env events.Envelope) ([]NotificationPlan, error) {
	p, err := events.DecodePayload[events.InscripcionCreadaPayload](env)
	if err != nil {
		return nil, err
	}

	plans := make([]NotificationPlan, 0, 2)

	// E1: Notificación al participante
	actionURL := "/dashboard/cursos/" + p.CursoID
	title := "Inscripción confirmada"
	msg := "Te has inscripto correctamente al curso formativo."
	if p.Titulo != "" {
		title = fmt.Sprintf("Inscripción confirmada: %s", p.Titulo)
		msg = fmt.Sprintf("Tu inscripción al curso «%s» ha sido registrada con éxito. Podés acceder al contenido desde la plataforma.", p.Titulo)
	}
	plans = append(plans, NotificationPlan{
		SourceService: "education",
		EventType:     env.EventType,
		Category:      models.CategoryEducation,
		Priority:      models.PriorityNormal,
		AudienceType:  AudienceParticipant,
		ParticipantID: p.ParticipanteID,
		Title:         title,
		Message:       msg,
		ActionURL:     &actionURL,
		SendEmail:     true,
		SendPush:      true,
		SendInApp:     true,
		DedupeKey:     fmt.Sprintf("%s_participant_%s", env.EventID, p.ParticipanteID),
	})

	// E2: Notificación al docente si está presente
	if p.DocenteID != "" {
		teacherURL := "/dashboard/cursos/" + p.CursoID
		docenteTitle := "Nueva inscripción en tu curso"
		docenteMsg := "Un participante se ha inscripto a tu curso."
		if p.Titulo != "" {
			docenteTitle = fmt.Sprintf("Nueva inscripción: %s", p.Titulo)
			docenteMsg = fmt.Sprintf("Un participante se inscribió a tu curso «%s».", p.Titulo)
		}
		plans = append(plans, NotificationPlan{
			SourceService:    "education",
			EventType:        env.EventType,
			Category:         models.CategoryEducation,
			Priority:         models.PriorityLow,
			AudienceType:     AudienceIndividual,
			RecipientUserIDs: []string{p.DocenteID},
			Title:            docenteTitle,
			Message:          docenteMsg,
			ActionURL:        &teacherURL,
			SendInApp:        true,
			DedupeKey:        fmt.Sprintf("%s_teacher_%s", env.EventID, p.DocenteID),
		})
	}

	return plans, nil
}

// ── E3: ContenidoDocenteSubidoRule ─────────────────────────────────────────

type ContenidoDocenteSubidoRule struct{}

func (r *ContenidoDocenteSubidoRule) EventType() string {
	return events.ContenidoDocenteSubido
}

func (r *ContenidoDocenteSubidoRule) Build(_ context.Context, env events.Envelope) ([]NotificationPlan, error) {
	p, err := events.DecodePayload[events.ContenidoDocenteSubidoPayload](env)
	if err != nil {
		return nil, err
	}

	adminRole := "role_admin"
	actionURL := "/dashboard/contenido-docente"
	title := "Nuevo contenido pendiente de moderación"
	message := "Un docente ha subido nuevo material educativo para revisión."
	// Curso creado por un docente: queda en borrador hasta que el admin lo aprueba y
	// publica desde Gestión de cursos.
	if p.Tipo == events.ContenidoTipoCurso {
		actionURL = "/dashboard/gestion-cursos"
		title = "Nuevo curso pendiente de aprobación"
		message = "Un docente creó un curso. Revisalo y publicalo desde Gestión de cursos."
	}
	return []NotificationPlan{
		{
			SourceService: "education",
			EventType:     env.EventType,
			Category:      models.CategoryEducation,
			Priority:      models.PriorityNormal,
			AudienceType:  AudienceRole,
			TargetRole:    &adminRole,
			Title:         title,
			Message:       message,
			ActionURL:     &actionURL,
			SendEmail:     true,
			SendInApp:     true,
			DedupeKey:     fmt.Sprintf("%s_admin", env.EventID),
		},
	}, nil
}

// ── E4: ContenidoDocenteModeradoRule ───────────────────────────────────────

type ContenidoDocenteModeradoRule struct{}

func (r *ContenidoDocenteModeradoRule) EventType() string {
	return events.ContenidoDocenteModerado
}

func (r *ContenidoDocenteModeradoRule) Build(_ context.Context, env events.Envelope) ([]NotificationPlan, error) {
	p, err := events.DecodePayload[events.ContenidoDocenteModeradoPayload](env)
	if err != nil {
		return nil, err
	}

	if p.DocenteID == "" {
		return nil, nil
	}

	actionURL := "/dashboard/mis-contenidos"
	msg := fmt.Sprintf("Tu contenido docente ha sido %s por la administración.", p.Resultado)

	return []NotificationPlan{
		{
			SourceService:    "education",
			EventType:        env.EventType,
			Category:         models.CategoryEducation,
			Priority:         models.PriorityHigh,
			AudienceType:     AudienceIndividual,
			RecipientUserIDs: []string{p.DocenteID},
			Title:            "Resultado de moderación de contenido",
			Message:          msg,
			ActionURL:        &actionURL,
			SendEmail:        true,
			SendPush:         true,
			SendInApp:        true,
			DedupeKey:        fmt.Sprintf("%s_docente_%s", env.EventID, p.DocenteID),
		},
	}, nil
}

// ── E5 & E6: CursoCompletadoRule ───────────────────────────────────────────

type CursoCompletadoRule struct{}

func (r *CursoCompletadoRule) EventType() string {
	return events.CursoCompletadoParticipante
}

func (r *CursoCompletadoRule) Build(_ context.Context, env events.Envelope) ([]NotificationPlan, error) {
	p, err := events.DecodePayload[events.CursoCompletadoPayload](env)
	if err != nil {
		return nil, err
	}

	actionURL := "/dashboard/cursos"
	return []NotificationPlan{
		{
			SourceService: "education",
			EventType:     env.EventType,
			Category:      models.CategoryEducation,
			Priority:      models.PriorityNormal,
			AudienceType:  AudienceParticipant,
			ParticipantID: p.ParticipanteID,
			Title:         "¡Felicitaciones! Curso completado",
			Message:       "Has completado todas las actividades del curso satisfactoriamente.",
			ActionURL:     &actionURL,
			SendPush:      true,
			SendInApp:     true,
			DedupeKey:     fmt.Sprintf("%s_completed_%s", env.EventID, p.ParticipanteID),
		},
	}, nil
}

// ── E7: HabilidadOtorgadaRule ──────────────────────────────────────────────

type HabilidadOtorgadaRule struct{}

func (r *HabilidadOtorgadaRule) EventType() string {
	return events.HabilidadOtorgada
}

func (r *HabilidadOtorgadaRule) Build(_ context.Context, env events.Envelope) ([]NotificationPlan, error) {
	p, err := events.DecodePayload[events.HabilidadOtorgadaPayload](env)
	if err != nil {
		return nil, err
	}

	actionURL := "/dashboard/mi-proceso"
	// Education manda el nombre de la habilidad; los eventos viejos (sin nombre) usan el
	// texto genérico.
	message := "Obtuviste una nueva habilidad por completar una actividad. Mirala en Mi proceso."
	if nombre := strings.TrimSpace(p.HabilidadNombre); nombre != "" {
		message = fmt.Sprintf("Obtuviste la habilidad «%s» por completar una actividad. Ya figura en Mi proceso y en tu CV.", nombre)
	}
	return []NotificationPlan{
		{
			SourceService: "education",
			EventType:     env.EventType,
			Category:      models.CategoryEducation,
			Priority:      models.PriorityLow,
			AudienceType:  AudienceParticipant,
			ParticipantID: p.ParticipanteID,
			Title:         "¡Nueva habilidad obtenida!",
			Message:       message,
			ActionURL:     &actionURL,
			SendPush:      true,
			SendInApp:     true,
			DedupeKey:     fmt.Sprintf("%s_skill_%s", env.EventID, p.HabilidadID),
		},
	}, nil
}

// ── E8: CursoEliminadoRule ─────────────────────────────────────────────────

type CursoEliminadoRule struct{}

func (r *CursoEliminadoRule) EventType() string {
	return events.CursoEliminado
}

func (r *CursoEliminadoRule) Build(_ context.Context, env events.Envelope) ([]NotificationPlan, error) {
	adminRole := "role_admin"
	return []NotificationPlan{
		{
			SourceService: "education",
			EventType:     env.EventType,
			Category:      models.CategoryEducation,
			Priority:      models.PriorityLow,
			AudienceType:  AudienceRole,
			TargetRole:    &adminRole,
			Title:         "Curso dado de baja",
			Message:       "Un curso formativo ha sido dado de baja en el catálogo general.",
			SendInApp:     true,
			DedupeKey:     fmt.Sprintf("%s_admin", env.EventID),
		},
	}, nil
}

// ── E9: CursoCreadoRule ───────────────────────────────────────────────────

type CursoCreadoRule struct{}

func (r *CursoCreadoRule) EventType() string {
	return events.CursoCreado
}

func (r *CursoCreadoRule) Build(_ context.Context, env events.Envelope) ([]NotificationPlan, error) {
	adminRole := "role_admin"
	actionURL := "/dashboard/cursos"
	return []NotificationPlan{
		{
			SourceService: "education",
			EventType:     env.EventType,
			Category:      models.CategoryEducation,
			Priority:      models.PriorityLow,
			AudienceType:  AudienceRole,
			TargetRole:    &adminRole,
			Title:         "Nuevo curso creado",
			Message:       "Se ha publicado un nuevo curso formativo en el catálogo.",
			ActionURL:     &actionURL,
			SendInApp:     true,
			DedupeKey:     fmt.Sprintf("%s_admin", env.EventID),
		},
	}, nil
}

// ── E10: ActividadRecordatorioRule ──────────────────────────────────────────

type ActividadRecordatorioRule struct{}

func (r *ActividadRecordatorioRule) EventType() string {
	return events.ActividadRecordatorio
}

func (r *ActividadRecordatorioRule) Build(_ context.Context, env events.Envelope) ([]NotificationPlan, error) {
	p, err := events.DecodePayload[events.ActividadRecordatorioPayload](env)
	if err != nil {
		return nil, err
	}

	actionURL := "/dashboard/cursos/" + p.CursoID
	title := "Recordatorio de actividad"
	msg := fmt.Sprintf("Tenés una actividad programada para la fecha %s.", p.Fecha)
	if p.Titulo != "" {
		title = fmt.Sprintf("Recordatorio: %s", p.Titulo)
		msg = fmt.Sprintf("Tenés una actividad programada para mañana (%s): «%s».", p.Fecha, p.Titulo)
	}

	return []NotificationPlan{
		{
			SourceService: "education",
			EventType:     env.EventType,
			Category:      models.CategoryEducation,
			Priority:      models.PriorityNormal,
			AudienceType:  AudienceParticipant,
			ParticipantID: p.ParticipanteID,
			Title:         title,
			Message:       msg,
			ActionURL:     &actionURL,
			SendEmail:     true,
			SendPush:      true,
			SendInApp:     true,
			DedupeKey:     fmt.Sprintf("%s_reminder_%s", env.EventID, p.ParticipanteID),
		},
	}, nil
}
