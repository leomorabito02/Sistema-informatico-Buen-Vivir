package services

import (
	"context"
	"errors"

	"buenvivir/auth/internal/models"
	"buenvivir/auth/internal/repositories"
)

type AuthService interface {
	Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error)
}

type authService struct {
	userRepo repositories.UserRepository
	jwtSecret string
}

func NewAuthService(userRepo repositories.UserRepository, jwtSecret string) AuthService {
	return &authService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

func (s *authService) Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, errors.New("email y contraseña son obligatorios")
	}

	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("credenciales inválidas")
	}

	// Mock token generation
	mockToken := "mock-jwt-token-buenvivir-authenticated"

	return &models.LoginResponse{
		Token: mockToken,
		User:  user,
	}, nil
}
