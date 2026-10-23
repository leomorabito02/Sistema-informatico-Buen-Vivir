package services

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"buenvivir/human-promotion/internal/models"
	"buenvivir/human-promotion/internal/repositories"
)

// MaxRegistrosAuditoria acota cada consulta del log.
const MaxRegistrosAuditoria = 500

const fechaAuditoriaLayout = "2006-01-02"

// zonaCentro es la hora de Argentina (UTC-3, sin horario de verano):
// "desde" y "hasta" son días del calendario del centro.
var zonaCentro = time.FixedZone("UTC-3", -3*60*60)

// FiltroAuditoria son los filtros de la consulta. Desde y Hasta van en
// AAAA-MM-DD y los dos días se incluyen; vacíos = últimos 30 días.
type FiltroAuditoria struct {
	Desde   string
	Hasta   string
	Accion  string
	Recurso string
}

// AuditoriaService es la consulta de solo lectura del log de auditoría
// de fichas e hitos (HU-30 y HU-35).
type AuditoriaService interface {
	List(ctx context.Context, f FiltroAuditoria) ([]models.RegistroAuditoria, error)
}

type auditoriaService struct {
	repo repositories.AuditoriaRepository
	auth repositories.AuthDirectory
	now  func() time.Time
}

func NewAuditoriaService(repo repositories.AuditoriaRepository, auth repositories.AuthDirectory) AuditoriaService {
	return &auditoriaService{repo: repo, auth: auth, now: time.Now}
}

func (s *auditoriaService) List(ctx context.Context, f FiltroAuditoria) ([]models.RegistroAuditoria, error) {
	desde, hasta, err := RangoAuditoria(f.Desde, f.Hasta, s.now())
	if err != nil {
		return nil, err
	}
	registros, err := s.repo.List(ctx, desde, hasta, MaxRegistrosAuditoria)
	if err != nil {
		return nil, err
	}

	out := make([]models.RegistroAuditoria, 0, len(registros))
	ids := map[string]bool{}
	for _, r := range registros {
		if (f.Accion == "" || r.Accion == f.Accion) && (f.Recurso == "" || r.Recurso == f.Recurso) {
			out = append(out, r)
			if r.UsuarioID != "" {
				ids[r.UsuarioID] = true
			}
		}
	}
	if len(ids) == 0 {
		return out, nil
	}

	// Quién hizo cada cambio: si Auth no responde, el log sale igual con
	// los ids.
	userIDs := make([]string, 0, len(ids))
	for id := range ids {
		userIDs = append(userIDs, id)
	}
	nombres, err := s.auth.GetNombresUsuarios(ctx, userIDs)
	if err != nil {
		slog.WarnContext(ctx, "auditoría sin nombres de usuario", slog.String("error", err.Error()))
		return out, nil
	}
	for i := range out {
		out[i].UsuarioNombre = nombres[out[i].UsuarioID]
	}
	return out, nil
}

// RangoAuditoria convierte desde/hasta (AAAA-MM-DD, días del centro) en
// el intervalo [desde 00:00, día siguiente a hasta 00:00). Vacíos: los
// últimos 30 días hasta hoy. Hasta un año por consulta.
func RangoAuditoria(desde, hasta string, now time.Time) (time.Time, time.Time, error) {
	hoy := now.In(zonaCentro)
	fin := time.Date(hoy.Year(), hoy.Month(), hoy.Day(), 0, 0, 0, 0, zonaCentro)
	if strings.TrimSpace(hasta) != "" {
		t, err := time.ParseInLocation(fechaAuditoriaLayout, strings.TrimSpace(hasta), zonaCentro)
		if err != nil {
			return time.Time{}, time.Time{}, NewValidationError("Fecha hasta inválida: se espera AAAA-MM-DD")
		}
		fin = t
	}
	inicio := fin.AddDate(0, 0, -29)
	if strings.TrimSpace(desde) != "" {
		t, err := time.ParseInLocation(fechaAuditoriaLayout, strings.TrimSpace(desde), zonaCentro)
		if err != nil {
			return time.Time{}, time.Time{}, NewValidationError("Fecha desde inválida: se espera AAAA-MM-DD")
		}
		inicio = t
	}
	if inicio.After(fin) {
		return time.Time{}, time.Time{}, NewValidationError("La fecha desde no puede ser posterior a la fecha hasta")
	}
	if fin.Sub(inicio) > 366*24*time.Hour {
		return time.Time{}, time.Time{}, NewValidationError("El rango puede ser de hasta un año")
	}
	return inicio.UTC(), fin.AddDate(0, 0, 1).UTC(), nil
}
