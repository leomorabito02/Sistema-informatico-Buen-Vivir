package dto

import "auth/internal/models"

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
	Role     string `json:"role" validate:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
	Role     string `json:"role" validate:"required"`
}

type GoogleLoginRequest struct {
	IDToken  string `json:"id_token" validate:"required"`
	GoogleID string `json:"google_id" validate:"required"`
	Role     string `json:"role" validate:"required"`
}

type LoginResponse struct {
	Token        string        `json:"token"`
	RefreshToken string        `json:"refresh_token,omitempty"`
	User         *models.User  `json:"user"`
}
