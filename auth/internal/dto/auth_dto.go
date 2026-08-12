package dto

import "auth/internal/models"

type RegisterRequest struct {
	FirstName string `json:"first_name" validate:"required,min=2"`
	LastName  string `json:"last_name" validate:"required,min=2"`
	Email     string `json:"email" validate:"required,email"`
	Phone     string `json:"phone" validate:"required"`
	Password  string `json:"password" validate:"required,min=8"`
	Role      string `json:"role" validate:"required,oneof=admin volunteer participant teacher company"`
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
	Token          string        `json:"token"`
	RefreshToken   string        `json:"refresh_token,omitempty"`
	User           *models.User  `json:"user"`
	ApprovalStatus string        `json:"approval_status,omitempty"`
}
