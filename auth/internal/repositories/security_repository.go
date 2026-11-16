package repositories

import (
	"context"
	"errors"
	"strings"
	"time"

	"auth/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// LoginAttemptRepository guarda los intentos fallidos de login (HU-02).
type LoginAttemptRepository interface {
	// Get devuelve el registro de la clave, o nil si no hay ninguno.
	Get(ctx context.Context, key string) (*models.LoginAttempt, error)
	// RegisterFailure suma un fallo. Al llegar a max, bloquea la clave
	// durante lock y reinicia el contador. Devuelve el estado final.
	RegisterFailure(ctx context.Context, key string, max int, lock time.Duration) (*models.LoginAttempt, error)
	// Reset borra los fallos (login correcto).
	Reset(ctx context.Context, key string) error
}

type loginAttemptRepository struct{ db *gorm.DB }

func NewLoginAttemptRepository(db *gorm.DB) LoginAttemptRepository {
	return &loginAttemptRepository{db: db}
}

func (r *loginAttemptRepository) Get(ctx context.Context, key string) (*models.LoginAttempt, error) {
	var la models.LoginAttempt
	err := r.db.WithContext(ctx).Where("email_hash = ?", key).First(&la).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &la, nil
}

func (r *loginAttemptRepository) RegisterFailure(ctx context.Context, key string, max int, lock time.Duration) (*models.LoginAttempt, error) {
	var out models.LoginAttempt
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var la models.LoginAttempt
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("email_hash = ?", key).First(&la).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			la = models.LoginAttempt{EmailHash: key}
		} else if err != nil {
			return err
		}
		now := time.Now()
		if la.LockedUntil != nil && !la.LockedUntil.After(now) {
			la.LockedUntil = nil // el bloqueo anterior ya venció
		}
		la.Failed++
		if la.Failed >= max {
			until := now.Add(lock)
			la.LockedUntil = &until
			la.Failed = 0
		}
		if err := tx.Save(&la).Error; err != nil {
			return err
		}
		out = la
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *loginAttemptRepository) Reset(ctx context.Context, key string) error {
	return r.db.WithContext(ctx).Where("email_hash = ?", key).Delete(&models.LoginAttempt{}).Error
}

// AuditFilter filtra GET /api/v1/auth/auditoria. Desde y Hasta son
// fechas YYYY-MM-DD inclusive (zona de Argentina, como el resto).
type AuditFilter struct {
	Desde   time.Time
	Hasta   time.Time
	Accion  string
	Recurso string
	Limit   int
}

// AuditRepository es el log de auditoría de Autenticación: solo se
// agregan registros y se leen; no hay métodos para editar ni borrar.
type AuditRepository interface {
	Create(ctx context.Context, entry *models.AuditLog) error
	List(ctx context.Context, f AuditFilter) ([]models.AuditLog, error)
	UserNames(ctx context.Context, ids []string) (map[string]string, error)
}

type auditRepository struct{ db *gorm.DB }

func NewAuditRepository(db *gorm.DB) AuditRepository { return &auditRepository{db: db} }

func (r *auditRepository) Create(ctx context.Context, entry *models.AuditLog) error {
	return r.db.WithContext(ctx).Create(entry).Error
}

func (r *auditRepository) List(ctx context.Context, f AuditFilter) ([]models.AuditLog, error) {
	q := r.db.WithContext(ctx).Model(&models.AuditLog{})
	if !f.Desde.IsZero() {
		q = q.Where("fecha >= ?", f.Desde)
	}
	if !f.Hasta.IsZero() {
		q = q.Where("fecha < ?", f.Hasta)
	}
	if f.Accion != "" {
		q = q.Where("accion = ?", f.Accion)
	}
	if f.Recurso != "" {
		q = q.Where("recurso = ?", f.Recurso)
	}
	limit := f.Limit
	if limit <= 0 || limit > 1000 {
		limit = 500
	}
	var out []models.AuditLog
	err := q.Order("fecha DESC").Limit(limit).Find(&out).Error
	return out, err
}

// AuditImmutabilitySQL crea un trigger que rechaza UPDATE y DELETE sobre
// la tabla auditoria (HU-10: "no puede ser modificado ni eliminado").
// Lo ejecutan cmd/migrate y el arranque del servicio; es idempotente.
const AuditImmutabilitySQL = `
CREATE OR REPLACE FUNCTION auditoria_inmutable() RETURNS trigger AS $$
BEGIN
  RAISE EXCEPTION 'el registro de auditoria no se puede modificar ni borrar';
END;
$$ LANGUAGE plpgsql;
DROP TRIGGER IF EXISTS auditoria_inmutable ON auditoria;
CREATE TRIGGER auditoria_inmutable BEFORE UPDATE OR DELETE ON auditoria
  FOR EACH ROW EXECUTE FUNCTION auditoria_inmutable();`

// UserNames devuelve "Nombre Apellido" por user_id, incluidos los dados
// de baja (para mostrar quién hizo cada acción).
func (r *auditRepository) UserNames(ctx context.Context, ids []string) (map[string]string, error) {
	out := map[string]string{}
	if len(ids) == 0 {
		return out, nil
	}
	var rows []struct {
		UserID    string
		FirstName string
		LastName  string
	}
	err := r.db.WithContext(ctx).Unscoped().Model(&models.User{}).
		Select("user_id, first_name, last_name").Where("user_id IN ?", ids).Scan(&rows).Error
	for _, row := range rows {
		out[row.UserID] = strings.TrimSpace(row.FirstName + " " + row.LastName)
	}
	return out, err
}
