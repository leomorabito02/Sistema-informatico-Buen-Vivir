package repositories

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

// ParticipantProfileRow es la vista de solo lectura de un participante
// con su nombre (join participants + users) que consumen otros
// servicios por gRPC.
type ParticipantProfileRow struct {
	ParticipantID  string
	UserID         string
	FirstName      string
	LastName       string
	Birthdate      time.Time
	ApprovalStatus string
}

// UserSummaryRow es el nombre visible de un usuario.
type UserSummaryRow struct {
	UserID    string
	FirstName string
	LastName  string
}

// ErrParticipantProfileNotFound señala que no hay participante para el
// id pedido.
var ErrParticipantProfileNotFound = errors.New("participante no encontrado")

// ParticipantDirectoryRepository agrupa las consultas de solo lectura
// que necesita Promoción Humana (ficha, búsqueda y línea de tiempo). Es
// una interfaz aparte de ParticipantRepository/UserRepository para no
// ampliar contratos que ya implementan otros mocks y servicios.
type ParticipantDirectoryRepository interface {
	FindProfileByParticipantID(ctx context.Context, participantID string) (*ParticipantProfileRow, error)
	FindProfileByUserID(ctx context.Context, userID string) (*ParticipantProfileRow, error)
	// SearchProfiles busca por coincidencia parcial (sin distinguir
	// mayúsculas) en nombre, apellido o "nombre apellido".
	SearchProfiles(ctx context.Context, query string, limit int) ([]ParticipantProfileRow, error)
	FindUserSummaries(ctx context.Context, userIDs []string) ([]UserSummaryRow, error)
}

type participantDirectoryRepository struct {
	db *gorm.DB
}

func NewParticipantDirectoryRepository(db *gorm.DB) ParticipantDirectoryRepository {
	return &participantDirectoryRepository{db: db}
}

const participantProfileSelect = "participants.participant_id, participants.user_id, users.first_name, users.last_name, participants.birthdate, participants.approval_status"

func (r *participantDirectoryRepository) profileQuery(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).
		Table("participants").
		Select(participantProfileSelect).
		Joins("JOIN users ON users.user_id = participants.user_id AND users.deleted_at IS NULL").
		Where("participants.deleted_at IS NULL")
}

func (r *participantDirectoryRepository) FindProfileByParticipantID(ctx context.Context, participantID string) (*ParticipantProfileRow, error) {
	var rows []ParticipantProfileRow
	if err := r.profileQuery(ctx).Where("participants.participant_id = ?", participantID).Limit(1).Scan(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, ErrParticipantProfileNotFound
	}
	return &rows[0], nil
}

func (r *participantDirectoryRepository) FindProfileByUserID(ctx context.Context, userID string) (*ParticipantProfileRow, error) {
	var rows []ParticipantProfileRow
	if err := r.profileQuery(ctx).Where("participants.user_id = ?", userID).Limit(1).Scan(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, ErrParticipantProfileNotFound
	}
	return &rows[0], nil
}

func (r *participantDirectoryRepository) SearchProfiles(ctx context.Context, query string, limit int) ([]ParticipantProfileRow, error) {
	pattern := "%" + escapeLike(query) + "%"
	var rows []ParticipantProfileRow
	err := r.profileQuery(ctx).
		Where(
			"users.first_name ILIKE ? ESCAPE '\\' OR users.last_name ILIKE ? ESCAPE '\\' OR (users.first_name || ' ' || users.last_name) ILIKE ? ESCAPE '\\'",
			pattern, pattern, pattern,
		).
		Order("users.last_name, users.first_name").
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *participantDirectoryRepository) FindUserSummaries(ctx context.Context, userIDs []string) ([]UserSummaryRow, error) {
	if len(userIDs) == 0 {
		return []UserSummaryRow{}, nil
	}
	var rows []UserSummaryRow
	err := r.db.WithContext(ctx).
		Table("users").
		Select("user_id, first_name, last_name").
		Where("user_id IN ? AND deleted_at IS NULL", userIDs).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// escapeLike neutraliza los comodines de LIKE (% y _) que vengan en el
// texto buscado, para que se busquen literalmente.
func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}
