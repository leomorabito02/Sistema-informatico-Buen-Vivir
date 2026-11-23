package services

import (
	"context"
	"errors"
	"log/slog"
	"sort"
	"strings"
	"time"

	"buenvivir/labor-integration/internal/models"
	"buenvivir/labor-integration/internal/repositories"
)

// ErrOfferInvalid: falta un campo obligatorio de la oferta (HU-39). El
// controlador responde 400 con el mensaje.
var ErrOfferInvalid = errors.New("oferta inválida")

type offerFieldError struct{ msg string }

func (e *offerFieldError) Error() string        { return e.msg }
func (e *offerFieldError) Is(target error) bool { return target == ErrOfferInvalid }

// validateOfferFields exige todos los campos de la HU-39: título,
// descripción, ubicación, modalidad, tipo de empleo, salario, horario,
// al menos un requisito y fecha límite. checkDeadline pide además que la
// fecha límite no haya pasado (al publicar).
func validateOfferFields(o *models.JobOffer, checkDeadline bool) error {
	campos := []struct{ valor, nombre string }{
		{o.Title, "título"}, {o.Description, "descripción"}, {o.Location, "ubicación"},
		{o.Modality, "modalidad"}, {o.JobType, "tipo de empleo"}, {o.Salary, "salario"},
		{o.Schedule, "horario"},
	}
	// Se juntan todos los faltantes en un solo mensaje. El frontend usa la misma lista
	// (camposFaltantesOferta).
	var faltan []string
	for _, c := range campos {
		if strings.TrimSpace(c.valor) == "" {
			faltan = append(faltan, c.nombre)
		}
	}
	if o.Vacancies < 1 {
		faltan = append(faltan, "vacantes (al menos 1)")
	}
	tieneRequisito := false
	for _, r := range o.Requirements {
		if strings.TrimSpace(r) != "" {
			tieneRequisito = true
			break
		}
	}
	if !tieneRequisito {
		faltan = append(faltan, "requisitos (al menos uno)")
	}
	if o.Deadline.IsZero() {
		faltan = append(faltan, "fecha límite")
	}
	switch len(faltan) {
	case 0:
	case 1:
		return &offerFieldError{msg: "falta el campo obligatorio: " + faltan[0]}
	default:
		return &offerFieldError{msg: "faltan los campos obligatorios: " + strings.Join(faltan, ", ")}
	}
	if checkDeadline && o.Deadline.Before(time.Now().Add(-24*time.Hour)) {
		return &offerFieldError{msg: "la fecha límite ya pasó"}
	}
	return nil
}

// WithSkillsClient activa el cruce de ofertas compatibles (HU-44).
func WithSkillsClient(c repositories.SkillsClient) JobOfferServiceOption {
	return func(s *JobOfferService) { s.skills = c }
}

type bearerKey struct{}

// WithBearer guarda el token de quien hace el pedido (lo pone el
// controlador) para reenviarlo a Educación en el cruce de HU-44.
func WithBearer(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, bearerKey{}, token)
}

func bearerFrom(ctx context.Context) string {
	v, _ := ctx.Value(bearerKey{}).(string)
	return v
}

// MatchCompatibles devuelve, ordenados y sin repetir, los participantes con al menos una
// habilidad que coincide con un requisito de la oferta (HU-44), sin distinguir
// mayúsculas ni acentos.
func MatchCompatibles(requisitos []string, skills []repositories.GrantedSkill) []string {
	pedidos := map[string]bool{}
	for _, r := range requisitos {
		if n := strings.Join(strings.Fields(normalizar(r)), " "); n != "" {
			pedidos[n] = true
		}
	}
	set := map[string]bool{}
	for _, sk := range skills {
		if sk.ParticipanteID != "" && pedidos[strings.Join(strings.Fields(normalizar(sk.Nombre)), " ")] {
			set[sk.ParticipanteID] = true
		}
	}
	out := make([]string, 0, len(set))
	for id := range set {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}

// notifyCompatibles busca los participantes compatibles con una oferta
// recién publicada y emite insercion_laboral.oferta.compatible. Nunca
// hace fallar la aprobación: un error queda en el log.
func (s *JobOfferService) notifyCompatibles(ctx context.Context, offer *models.JobOffer) {
	if s.skills == nil || s.producer == nil || offer == nil || offer.Status != "Activa" {
		return
	}
	token := bearerFrom(ctx)
	if token == "" {
		slog.Warn("oferta compatible: sin token para consultar Educación", slog.String("offer_id", offer.ID))
		return
	}
	skills, err := s.skills.GrantedSkills(ctx, token)
	if err != nil {
		slog.Warn("oferta compatible: no se pudo consultar Educación", slog.String("offer_id", offer.ID), slog.String("error", err.Error()))
		return
	}
	ids := MatchCompatibles(offer.Requirements, skills)
	slog.Info("oferta compatible: cruce hecho", slog.String("offer_id", offer.ID), slog.Int("participantes", len(ids)))
	if len(ids) == 0 {
		return
	}
	if err := s.producer.OfertaCompatible(ctx, offer.ID, ids); err != nil {
		slog.Warn("failed to emit oferta.compatible", slog.String("error", err.Error()), slog.String("offer_id", offer.ID))
		return
	}
	s.producer.Notify()
}
