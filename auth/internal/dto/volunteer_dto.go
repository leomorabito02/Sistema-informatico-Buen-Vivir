package dto

import (
	"time"

	"auth/internal/models"
)

type CreateVolunteerRequest struct {
	UserID         string `json:"user_id" validate:"required"`
	ApprovalStatus string `json:"approval_status" validate:"omitempty,oneof=pending approved"`
}

type UpdateVolunteerRequest struct {
	ApprovalStatus string `json:"approval_status" validate:"omitempty,oneof=pending approved"`
	ApprovedBy     string `json:"approved_by" validate:"omitempty"`
}

type VolunteerResponse struct {
	VolunteerID    string     `json:"volunteer_id"`
	UserID         string     `json:"user_id"`
	ApprovalStatus string     `json:"approval_status"`
	ApprovalDate   *time.Time `json:"approval_date,omitempty"`
	ApprovedBy     *string    `json:"approved_by,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func ToVolunteerResponse(v *models.Volunteer) *VolunteerResponse {
	return &VolunteerResponse{
		VolunteerID:    v.VolunteerID,
		UserID:         v.UserID,
		ApprovalStatus: v.ApprovalStatus,
		ApprovalDate:   v.ApprovalDate,
		ApprovedBy:     v.ApprovedBy,
		CreatedAt:      v.CreatedAt,
		UpdatedAt:      v.UpdatedAt,
	}
}
