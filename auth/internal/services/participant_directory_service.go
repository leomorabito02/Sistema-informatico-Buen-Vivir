package services

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"auth/internal/repositories"
)

const (
	minParticipantSearchLength = 2
	maxParticipantSearchLength = 100
	defaultParticipantSearch   = 20
	maxParticipantSearch       = 50
	// maxParticipantList acota el listado completo (búsqueda vacía) que
	// muestra Promoción Humana al entrar (pedido del 17/11/2026).
	maxParticipantList = 500
	maxUserSummaryIDs  = 200
)

// ErrInvalidDirectoryRequest señala parámetros inválidos en una consulta
// del directorio (texto de búsqueda muy corto, demasiados ids, etc.).
var ErrInvalidDirectoryRequest = errors.New("solicitud inválida")

// ParticipantDirectoryService expone, por gRPC, los datos de
// participantes y usuarios que necesita Promoción Humana. Es de solo
// lectura y no devuelve datos sensibles de Auth (DNI, dirección,
// contacto de emergencia, email).
type ParticipantDirectoryService interface {
	GetProfileByParticipantID(ctx context.Context, participantID string) (*repositories.ParticipantProfileRow, error)
	GetProfileByUserID(ctx context.Context, userID string) (*repositories.ParticipantProfileRow, error)
	SearchParticipants(ctx context.Context, query string, limit int) ([]repositories.ParticipantProfileRow, error)
	GetUserSummaries(ctx context.Context, userIDs []string) ([]repositories.UserSummaryRow, error)
}

type participantDirectoryService struct {
	repo repositories.ParticipantDirectoryRepository
}

func NewParticipantDirectoryService(repo repositories.ParticipantDirectoryRepository) ParticipantDirectoryService {
	return &participantDirectoryService{repo: repo}
}

func (s *participantDirectoryService) GetProfileByParticipantID(ctx context.Context, participantID string) (*repositories.ParticipantProfileRow, error) {
	if strings.TrimSpace(participantID) == "" {
		return nil, ErrInvalidDirectoryRequest
	}
	return s.repo.FindProfileByParticipantID(ctx, participantID)
}

func (s *participantDirectoryService) GetProfileByUserID(ctx context.Context, userID string) (*repositories.ParticipantProfileRow, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrInvalidDirectoryRequest
	}
	return s.repo.FindProfileByUserID(ctx, userID)
}

func (s *participantDirectoryService) SearchParticipants(ctx context.Context, query string, limit int) ([]repositories.ParticipantProfileRow, error) {
	query = strings.Join(strings.Fields(query), " ")
	length := utf8.RuneCountInString(query)
	// Sin texto: listado completo, ordenado por apellido y nombre. Con
	// texto, se mantiene el mínimo de 2 caracteres.
	if length == 0 {
		if limit <= 0 || limit > maxParticipantList {
			limit = maxParticipantList
		}
		return s.repo.SearchProfiles(ctx, "", limit)
	}
	if length < minParticipantSearchLength || length > maxParticipantSearchLength {
		return nil, ErrInvalidDirectoryRequest
	}
	if limit <= 0 || limit > maxParticipantSearch {
		limit = defaultParticipantSearch
	}
	return s.repo.SearchProfiles(ctx, query, limit)
}

func (s *participantDirectoryService) GetUserSummaries(ctx context.Context, userIDs []string) ([]repositories.UserSummaryRow, error) {
	seen := make(map[string]bool, len(userIDs))
	unique := make([]string, 0, len(userIDs))
	for _, id := range userIDs {
		id = strings.TrimSpace(id)
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		unique = append(unique, id)
	}
	if len(unique) > maxUserSummaryIDs {
		return nil, ErrInvalidDirectoryRequest
	}
	return s.repo.FindUserSummaries(ctx, unique)
}
