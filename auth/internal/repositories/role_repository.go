package repositories

import (
	"context"
	"errors"

	"auth/internal/models"
	"gorm.io/gorm"
)

type RoleRepository interface {
	FindByID(ctx context.Context, id string) (*models.Role, error)
	FindAll(ctx context.Context) ([]models.Role, error)
	Create(ctx context.Context, role *models.Role) error
	Update(ctx context.Context, role *models.Role) error
	Delete(ctx context.Context, id string) error
}

type roleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) FindByID(ctx context.Context, id string) (*models.Role, error) {
	var role models.Role
	if err := r.db.WithContext(ctx).First(&role, "role_id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("rol no encontrado")
		}
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) FindAll(ctx context.Context) ([]models.Role, error) {
	var roles []models.Role
	if err := r.db.WithContext(ctx).Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *roleRepository) Create(ctx context.Context, role *models.Role) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()
		return tx.Create(role).Error
	})
}

func (r *roleRepository) Update(ctx context.Context, role *models.Role) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()
		return tx.Save(role).Error
	})
}

func (r *roleRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()
		return tx.Where("role_id = ?", id).Delete(&models.Role{}).Error
	})
}
