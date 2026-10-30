package services

import (
	"context"
	"strings"
	"time"

	"buenvivir/education/internal/models"
	"buenvivir/education/internal/repositories"
)

// fechaLayout es el formato YYYY-MM-DD que usan tanto el id
// determinístico de Firestore como el query param `fecha` de la API.
const fechaLayout = "2006-01-02"

// GeneralAttendanceService valida las reglas de negocio de la
// asistencia general del centro (HU-19bis/RF-64).
type GeneralAttendanceService interface {
	// ListByDate arma el listado combinado: todos los participantes
	// aprobados del centro (auth-service, vía AuthServiceClient), con su
	// nombre/apellido resuelto (auth-service también, GET /users) y su
	// asistencia para `fecha`, si ya se cargó. bearerToken es el JWT
	// crudo del usuario autenticado en education (ver
	// middlewares.TokenKey), que se reenvía tal cual a auth-service.
	ListByDate(ctx context.Context, fecha, bearerToken string) ([]models.GeneralAttendanceView, error)
	// RegisterBatch valida, por cada item del lote, que el participante
	// exista en auth-service y esté aprobado, antes de guardar.
	RegisterBatch(ctx context.Context, fecha string, inputs []models.GeneralAttendanceInput, bearerToken string) error
}

type generalAttendanceService struct {
	repo       repositories.GeneralAttendanceRepository
	authClient *repositories.AuthServiceClient
}

func NewGeneralAttendanceService(repo repositories.GeneralAttendanceRepository, authClient *repositories.AuthServiceClient) GeneralAttendanceService {
	return &generalAttendanceService{repo: repo, authClient: authClient}
}

func (s *generalAttendanceService) ListByDate(ctx context.Context, fecha, bearerToken string) ([]models.GeneralAttendanceView, error) {
	fecha = strings.TrimSpace(fecha)
	if _, err := time.Parse(fechaLayout, fecha); err != nil {
		return nil, NewValidationError("fecha inválida: se espera el formato YYYY-MM-DD")
	}

	participants, err := s.authClient.ListParticipants(ctx, bearerToken)
	if err != nil {
		return nil, err
	}

	// Participante no tiene nombre propio (vive en Usuario, ver DER).
	// Resolverlo es una segunda llamada a auth-service, deliberadamente
	// NO fatal: si falla (auth-service lento/caído en ese instante), se
	// sigue devolviendo la asistencia general con los ids nomás, en vez
	// de romper la pantalla completa por un dato secundario.
	usersByID := make(map[string]models.AuthUser)
	if users, err := s.authClient.ListUsers(ctx, bearerToken); err == nil {
		for _, u := range users {
			usersByID[u.UserID] = u
		}
	}

	records, err := s.repo.ListByDate(ctx, fecha)
	if err != nil {
		return nil, err
	}
	byParticipant := make(map[string]models.GeneralAttendance, len(records))
	for _, rec := range records {
		byParticipant[rec.ParticipanteID] = rec
	}

	views := make([]models.GeneralAttendanceView, 0, len(participants))
	for _, p := range participants {
		// "participantes activos del centro" (HU-19bis) = participantes
		// aprobados. approval_status vacío se tolera (defensivo, no
		// debería pasar) para no ocultar participantes por un dato mal
		// formado en vez de fallar en silencio del otro lado.
		if !isParticipantApproved(p.ApprovalStatus) {
			continue
		}

		view := models.GeneralAttendanceView{ParticipanteID: p.ParticipantID}
		if u, ok := usersByID[p.UserID]; ok {
			view.Nombre = u.FirstName
			view.Apellido = u.LastName
		}
		if rec, ok := byParticipant[p.ParticipantID]; ok {
			presente := rec.Presente
			view.Presente = &presente
			view.Observacion = rec.Observacion
		}
		views = append(views, view)
	}
	return views, nil
}

func (s *generalAttendanceService) RegisterBatch(ctx context.Context, fecha string, inputs []models.GeneralAttendanceInput, bearerToken string) error {
	fecha = strings.TrimSpace(fecha)
	if _, err := time.Parse(fechaLayout, fecha); err != nil {
		return NewValidationError("fecha inválida: se espera el formato YYYY-MM-DD")
	}
	if len(inputs) == 0 {
		return NewValidationError("se requiere al menos un registro de asistencia")
	}
	for _, in := range inputs {
		if strings.TrimSpace(in.ParticipanteID) == "" {
			return NewValidationError("cada registro de asistencia requiere 'participanteId'")
		}
	}

	for _, in := range inputs {
		participant, err := s.authClient.GetParticipant(ctx, in.ParticipanteID, bearerToken)
		if err != nil {
			return err
		}
		if participant == nil {
			return NewValidationError("El participante " + in.ParticipanteID + " no existe")
		}
		if !isParticipantApproved(participant.ApprovalStatus) {
			return NewValidationError("El participante " + in.ParticipanteID + " no está activo")
		}
	}

	return s.repo.RegisterBatch(ctx, fecha, inputs)
}
