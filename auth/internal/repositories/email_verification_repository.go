package repositories

import (
	"context"
	"errors"
	"time"

	"auth/internal/models"
	"gorm.io/gorm"
)

type EmailVerificationRepository interface {
	Create(ctx context.Context, v *models.EmailVerification) error
	FindByTokenHash(ctx context.Context, tokenHash string) (*models.EmailVerification, error)
	MarkAsUsed(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
	DeleteExpired(ctx context.Context) error
}

type emailVerificationRepository struct {
	db *gorm.DB
}

func NewEmailVerificationRepository(db *gorm.DB) EmailVerificationRepository {
	return &emailVerificationRepository{db: db}
}

func (r *emailVerificationRepository) Create(ctx context.Context, v *models.EmailVerification) error {
	return r.db.WithContext(ctx).Create(v).Error
}

func (r *emailVerificationRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*models.EmailVerification, error) {
	var ev models.EmailVerification
	if err := r.db.WithContext(ctx).Where("token_hash = ?", tokenHash).First(&ev).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("token no encontrado")
		}
		return nil, err
	}
	return &ev, nil
}

func (r *emailVerificationRepository) MarkAsUsed(ctx context.Context, id string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&models.EmailVerification{}).
		Where("email_verification_id = ?", id).
		Update("used_at", &now).Error
}

func (r *emailVerificationRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("email_verification_id = ?", id).Delete(&models.EmailVerification{}).Error
}

func (r *emailVerificationRepository) DeleteExpired(ctx context.Context) error {
	now := time.Now()
	return r.db.WithContext(ctx).Where("expires_at < ? OR used_at IS NOT NULL", now).Delete(&models.EmailVerification{}).Error
}
