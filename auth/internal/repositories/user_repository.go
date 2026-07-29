package repositories

import (
	"context"
	"errors"
	"time"

	"buenvivir/auth/internal/models"
)

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	Create(ctx context.Context, user *models.User) error
}

type userRepository struct {
	// db *sql.DB / gorm.DB
}

func NewUserRepository() UserRepository {
	return &userRepository{}
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	// Mock temporal para estructura inicial
	if email == "admin@buenvivir.org" {
		return &models.User{
			ID:        "usr-1",
			Email:     email,
			Password:  "$2a$10$7EqJtq986P27GkS27EqJtO", // Mock hash
			Name:      "Administrador Buen Vivir",
			Role:      "ADMIN",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}, nil
	}
	return nil, errors.New("usuario no encontrado")
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	return nil
}
