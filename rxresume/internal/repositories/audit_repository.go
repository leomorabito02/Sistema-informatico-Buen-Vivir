package repositories

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"buenvivir/rxresume/internal/models"
)

// AuditFilter filtra la lectura (Desde inclusive, Hasta exclusivo).
type AuditFilter struct {
	Desde, Hasta time.Time
	Accion       string
}

// AuditRepository solo agrega y lee: no hay edición ni borrado.
type AuditRepository interface {
	Record(ctx context.Context, e models.AuditEntry) error
	List(ctx context.Context, f AuditFilter) ([]models.AuditEntry, error)
}

type gormAuditRepository struct{ db *gorm.DB }

// NewAuditRepository arma el repositorio. Con db nil (sin base) devuelve nil.
func NewAuditRepository(db *gorm.DB) AuditRepository {
	if db == nil {
		return nil
	}
	return &gormAuditRepository{db: db}
}

var errSinBase = errors.New("auditoría sin base de datos")

func (r *gormAuditRepository) Record(ctx context.Context, e models.AuditEntry) error {
	if r == nil || r.db == nil {
		return errSinBase
	}
	if e.Fecha.IsZero() {
		e.Fecha = time.Now().UTC()
	}
	return r.db.WithContext(ctx).Create(&e).Error
}

func (r *gormAuditRepository) List(ctx context.Context, f AuditFilter) ([]models.AuditEntry, error) {
	if r == nil || r.db == nil {
		return nil, errSinBase
	}
	q := r.db.WithContext(ctx).Model(&models.AuditEntry{})
	if !f.Desde.IsZero() {
		q = q.Where("fecha >= ?", f.Desde)
	}
	if !f.Hasta.IsZero() {
		q = q.Where("fecha < ?", f.Hasta)
	}
	if f.Accion != "" {
		q = q.Where("accion = ?", f.Accion)
	}
	out := []models.AuditEntry{}
	err := q.Order("fecha DESC").Limit(500).Find(&out).Error
	return out, err
}
