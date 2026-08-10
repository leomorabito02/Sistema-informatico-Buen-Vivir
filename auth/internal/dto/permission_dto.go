package dto

import (
	"time"

	"auth/internal/models"
)

type CreatePermissionRequest struct {
	Action      string `json:"action" validate:"required,min=2,max=50"`
	Resource    string `json:"resource" validate:"required,min=2,max=50"`
	Description string `json:"description" validate:"omitempty,max=255"`
}

type UpdatePermissionRequest struct {
	Action      string `json:"action" validate:"omitempty,min=2,max=50"`
	Resource    string `json:"resource" validate:"omitempty,min=2,max=50"`
	Description string `json:"description" validate:"omitempty,max=255"`
}

type PermissionResponse struct {
	PermissionID string    `json:"permission_id"`
	Action       string    `json:"action"`
	Resource     string    `json:"resource"`
	Description  string    `json:"description"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func ToPermissionResponse(p *models.Permission) *PermissionResponse {
	return &PermissionResponse{
		PermissionID: p.PermissionID,
		Action:       p.Action,
		Resource:     p.Resource,
		Description:  p.Description,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}
}
