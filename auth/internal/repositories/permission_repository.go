package repositories

import (
	"context"
	"errors"

	"auth/internal/models"
	"gorm.io/gorm"
)

type PermissionRepository interface {
	FindByID(ctx context.Context, id string) (*models.Permission, error)
	FindAll(ctx context.Context) ([]models.Permission, error)
	Create(ctx context.Context, permission *models.Permission) error
	Update(ctx context.Context, permission *models.Permission) error
	Delete(ctx context.Context, id string) error
	CheckRolePermission(ctx context.Context, roleID, action, resource string, count *int64) error
}

type permissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) PermissionRepository {
	return &permissionRepository{db: db}
}

func (r *permissionRepository) FindByID(ctx context.Context, id string) (*models.Permission, error) {
	var permission models.Permission
	if err := r.db.WithContext(ctx).First(&permission, "permission_id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("permiso no encontrado")
		}
		return nil, err
	}
	return &permission, nil
}

func (r *permissionRepository) FindAll(ctx context.Context) ([]models.Permission, error) {
	var permissions []models.Permission
	if err := r.db.WithContext(ctx).Find(&permissions).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}

func (r *permissionRepository) Create(ctx context.Context, permission *models.Permission) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()
		return tx.Create(permission).Error
	})
}

func (r *permissionRepository) Update(ctx context.Context, permission *models.Permission) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()
		return tx.Save(permission).Error
	})
}

func (r *permissionRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()
		return tx.Where("permission_id = ?", id).Delete(&models.Permission{}).Error
	})
}

func (r *permissionRepository) CheckRolePermission(ctx context.Context, roleID, action, resource string, count *int64) error {
	return r.db.WithContext(ctx).Table("role_permissions").
		Joins("JOIN permissions ON permissions.permission_id = role_permissions.permission_id").
		Where("role_permissions.role_id = ? AND permissions.action = ? AND permissions.resource = ?", roleID, action, resource).
		Count(count).Error
}
