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
}

type UpdateUserRequest struct {
	FirstName string `json:"first_name" validate:"omitempty,min=2,max=50"`
	LastName  string `json:"last_name" validate:"omitempty,min=2,max=50"`
	Phone     string `json:"phone" validate:"omitempty"`
	RoleID    string `json:"role_id" validate:"omitempty"`
	Status    string `json:"status" validate:"omitempty,oneof=active suspended"`
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
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
