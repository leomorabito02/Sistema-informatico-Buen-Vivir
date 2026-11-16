package dto

import (
	"time"
	"auth/internal/models"
)

type CreateUserRequest struct {
	FirstName string `json:"first_name" validate:"required,min=2,max=50"`
	LastName  string `json:"last_name" validate:"required,min=2,max=50"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,password"`
	Phone     string `json:"phone" validate:"required"`
	RoleID    string `json:"role_id" validate:"required"`
	// DNI y Address son obligatorios cuando role_id es role_participant
	// (HU-07); se validan en el servicio porque dependen del rol.
	DNI     string `json:"dni" validate:"omitempty"`
	Address string `json:"address" validate:"omitempty"`
}

type UpdateUserRequest struct {
	FirstName string `json:"first_name" validate:"omitempty,min=2,max=50"`
	LastName  string `json:"last_name" validate:"omitempty,min=2,max=50"`
	Phone     string `json:"phone" validate:"omitempty"`
	RoleID    string `json:"role_id" validate:"omitempty"`
	Status    string `json:"status" validate:"omitempty,oneof=active suspended"`
}

// ParticipantHistoryResponse es un participante en el historial que usa
// el Dashboard (GET /api/v1/users/participants/history). Incluye a los
// dados de baja: deleted_at es la fecha de la baja lógica.
type ParticipantHistoryResponse struct {
	UserID         string     `json:"user_id"`
	Status         string     `json:"status"`
	ApprovalStatus string     `json:"approval_status"`
	ApprovalDate   *time.Time `json:"approval_date"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at"`
}

type UserResponse struct {
	UserID    string    `json:"user_id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Status    string    `json:"status"`
	RoleID    string    `json:"role_id"`
	RoleApprovalStatus string `json:"role_approval_status,omitempty"`
	ImageURL           string `json:"image_url,omitempty"`
	IsEmailVerified    bool   `json:"is_email_verified"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func ToUserResponse(u *models.User) *UserResponse {
	return &UserResponse{
		UserID:    u.UserID,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
		Phone:     u.Phone,
		Status:    u.Status,
		RoleID:    u.RoleID,
		RoleApprovalStatus: u.RoleApprovalStatus,
		ImageURL:  u.ImageURL,
		IsEmailVerified:    u.IsEmailVerified,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
