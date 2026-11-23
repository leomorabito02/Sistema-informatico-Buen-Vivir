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
	// ListByDate devuelve los participantes aprobados del centro con su nombre y su
	// asistencia de esa fecha. bearerToken se reenvía a auth-service.
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

	// Nombre y apellido viven en Usuario: si esta llamada falla no se corta la pantalla.
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
		// Solo participantes aprobados (HU-19bis).
		if !isParticipantApproved(p.ApprovalStatus) {
			continue
		}

		view := models.GeneralAttendanceView{ParticipanteID: p.ParticipantID}
		if u, ok := usersByID[p.UserID]; ok {
			view.Nombre = u.FirstName
			view.Apellido = u.LastName
		} else if u, err := s.authClient.GetParticipantName(ctx, p.ParticipantID, bearerToken); err == nil {
			// El docente no puede listar todos los usuarios: se pide uno por uno.
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
