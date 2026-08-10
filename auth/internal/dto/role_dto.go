package dto

import (
	"time"
	"auth/internal/models"
)

type CreateRoleRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=50"`
	Description string `json:"description" validate:"omitempty,max=255"`
}

type UpdateRoleRequest struct {
	Name        string `json:"name" validate:"omitempty,min=2,max=50"`
	Description string `json:"description" validate:"omitempty,max=255"`
}

type RoleResponse struct {
	RoleID      string    `json:"role_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func ToRoleResponse(r *models.Role) *RoleResponse {
	return &RoleResponse{
		RoleID:      r.RoleID,
		Name:        r.Name,
		Description: r.Description,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}
