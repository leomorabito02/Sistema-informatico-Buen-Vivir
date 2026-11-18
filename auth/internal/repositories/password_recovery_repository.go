package repositories

import (
	"context"
	"errors"
	"time"

	"auth/internal/models"
	"gorm.io/gorm"
)

type PasswordRecoveryRepository interface {
	Create(ctx context.Context, v *models.PasswordRecovery) error
	FindByTokenHash(ctx context.Context, tokenHash string) (*models.PasswordRecovery, error)
	MarkAsUsed(ctx context.Context, id string) error
	DeleteByUserID(ctx context.Context, userID string) error
	DeleteExpired(ctx context.Context) error
}

type passwordRecoveryRepository struct {
	db *gorm.DB
}

func NewPasswordRecoveryRepository(db *gorm.DB) PasswordRecoveryRepository {
	return &passwordRecoveryRepository{db: db}
}

func (r *passwordRecoveryRepository) Create(ctx context.Context, v *models.PasswordRecovery) error {
	return r.db.WithContext(ctx).Create(v).Error
}

func (r *passwordRecoveryRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*models.PasswordRecovery, error) {
	var pr models.PasswordRecovery
	if err := r.db.WithContext(ctx).Where("token_hash = ?", tokenHash).First(&pr).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("token no encontrado")
		}
		return nil, err
	}
	return &pr, nil
}

func (r *passwordRecoveryRepository) MarkAsUsed(ctx context.Context, id string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&models.PasswordRecovery{}).
		Where("password_recovery_id = ?", id).
		Update("used_at", &now).Error
}

func (r *passwordRecoveryRepository) DeleteByUserID(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&models.PasswordRecovery{}).Error
}

func (r *passwordRecoveryRepository) DeleteExpired(ctx context.Context) error {
	now := time.Now()
	return r.db.WithContext(ctx).Where("expires_at < ? OR used_at IS NOT NULL", now).Delete(&models.PasswordRecovery{}).Error
}
