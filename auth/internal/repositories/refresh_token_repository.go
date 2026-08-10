package repositories

import (
	"context"
	"errors"
	"time"

	"auth/internal/models"
	"gorm.io/gorm"
)

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *models.RefreshToken) error
	FindByTokenHash(ctx context.Context, hash string) (*models.RefreshToken, error)
	Revoke(ctx context.Context, id string) error
	RevokeAllForUser(ctx context.Context, userID string) error
	DeleteExpired(ctx context.Context) error
}

type refreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(ctx context.Context, token *models.RefreshToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *refreshTokenRepository) FindByTokenHash(ctx context.Context, hash string) (*models.RefreshToken, error) {
	var token models.RefreshToken
	if err := r.db.WithContext(ctx).Where("token_hash = ?", hash).First(&token).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("refresh token not found")
		}
		return nil, err
	}
	return &token, nil
}

func (r *refreshTokenRepository) Revoke(ctx context.Context, id string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&models.RefreshToken{}).
		Where("refresh_token_id = ?", id).
		Update("revoked_at", &now).Error
}

func (r *refreshTokenRepository) RevokeAllForUser(ctx context.Context, userID string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&models.RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", &now).Error
}

func (r *refreshTokenRepository) DeleteExpired(ctx context.Context) error {
	now := time.Now()
	return r.db.WithContext(ctx).Where("expires_at < ? OR revoked_at IS NOT NULL", now).Delete(&models.RefreshToken{}).Error
}
