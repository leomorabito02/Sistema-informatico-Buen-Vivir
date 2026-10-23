package services

import (
	"context"
	"strings"
	"time"

	"buenvivir/education/internal/models"
	"buenvivir/education/internal/repositories"
)

// MaxRegistrosAuditoria acota cada consulta del log.
const MaxRegistrosAuditoria = 500

// zonaCentro es la hora de Argentina (UTC-3, sin horario de verano):
// "desde" y "hasta" son días del calendario del centro.
var zonaCentro = time.FixedZone("UTC-3", -3*60*60)

// FiltroAuditoria son los filtros de GET /api/auditoria. Desde y Hasta
// van en YYYY-MM-DD y los dos días se incluyen; vacíos = últimos 30 días.
type FiltroAuditoria struct {
	Desde   string
	Hasta   string
	Accion  string
	Recurso string
}

// AuditoriaService es la consulta (solo lectura) del registro de
// auditoría de Educación, para administración.
type AuditoriaService interface {
	List(ctx context.Context, f FiltroAuditoria, bearerToken string) ([]models.RegistroAuditoria, error)
}

type auditoriaService struct {
	repo       repositories.AuditRepository
	authClient *repositories.AuthServiceClient
	now        func() time.Time
}

func NewAuditoriaService(repo repositories.AuditRepository, authClient *repositories.AuthServiceClient) AuditoriaService {
	return &auditoriaService{repo: repo, authClient: authClient, now: time.Now}
}

func (s *auditoriaService) List(ctx context.Context, f FiltroAuditoria, bearerToken string) ([]models.RegistroAuditoria, error) {
	desde, hasta, err := RangoAuditoria(f.Desde, f.Hasta, s.now())
	if err != nil {
		return nil, err
	}
	registros, err := s.repo.List(ctx, desde, hasta, MaxRegistrosAuditoria)
	if err != nil {
		return nil, err
	}

	out := make([]models.RegistroAuditoria, 0, len(registros))
	for _, r := range registros {
		if (f.Accion == "" || r.Accion == f.Accion) && (f.Recurso == "" || r.Recurso == f.Recurso) {
			out = append(out, r)
		}
	}
	// Quién hizo cada cambio: no fatal si auth-service no responde.
	if len(out) > 0 {
		if nombres, err := nombresDeUsuarios(ctx, s.authClient, bearerToken); err == nil {
			for i := range out {
				out[i].UsuarioNombre = nombres[out[i].UsuarioID]
			}
		}
	}
	return out, nil
}

// RangoAuditoria convierte desde/hasta (YYYY-MM-DD, días del centro) en
// el intervalo [desde 00:00, día siguiente a hasta 00:00). Vacíos: los
// últimos 30 días hasta hoy. Hasta un año por consulta.
func RangoAuditoria(desde, hasta string, now time.Time) (time.Time, time.Time, error) {
	hoy := now.In(zonaCentro)
	fin := time.Date(hoy.Year(), hoy.Month(), hoy.Day(), 0, 0, 0, 0, zonaCentro)
	if strings.TrimSpace(hasta) != "" {
		t, err := time.ParseInLocation(fechaLayout, strings.TrimSpace(hasta), zonaCentro)
		if err != nil {
			return time.Time{}, time.Time{}, NewValidationError("hasta inválida: se espera el formato YYYY-MM-DD")
		}
		fin = t
	}
	inicio := fin.AddDate(0, 0, -29)
	if strings.TrimSpace(desde) != "" {
		t, err := time.ParseInLocation(fechaLayout, strings.TrimSpace(desde), zonaCentro)
		if err != nil {
			return time.Time{}, time.Time{}, NewValidationError("desde inválida: se espera el formato YYYY-MM-DD")
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
