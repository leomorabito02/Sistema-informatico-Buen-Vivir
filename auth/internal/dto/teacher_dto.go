package dto

import (
	"time"

	"auth/internal/models"
)

type CreateTeacherRequest struct {
	UserID         string `json:"user_id" validate:"required"`
	Specialization string `json:"specialization" validate:"required,min=2,max=100"`
	IsExternal     bool   `json:"is_external"`
	ApprovalStatus string `json:"approval_status" validate:"omitempty,oneof=pending approved rejected"`
}

type UpdateTeacherRequest struct {
	Specialization string `json:"specialization" validate:"omitempty,min=2,max=100"`
	IsExternal     *bool  `json:"is_external" validate:"omitempty"`
	ApprovalStatus string `json:"approval_status" validate:"omitempty,oneof=pending approved rejected"`
	ApprovedBy     string `json:"approved_by" validate:"omitempty"`
}

type TeacherResponse struct {
	TeacherID      string     `json:"teacher_id"`
	UserID         string     `json:"user_id"`
	Specialization string     `json:"specialization"`
	IsExternal     bool       `json:"is_external"`
	ApprovalStatus string     `json:"approval_status"`
	ApprovalDate   *time.Time `json:"approval_date,omitempty"`
	ApprovedBy     *string    `json:"approved_by,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func ToTeacherResponse(t *models.Teacher) *TeacherResponse {
	return &TeacherResponse{
		TeacherID:      t.TeacherID,
		UserID:         t.UserID,
		Specialization: t.Specialization,
		IsExternal:     t.IsExternal,
		ApprovalStatus: t.ApprovalStatus,
		ApprovalDate:   t.ApprovalDate,
		ApprovedBy:     t.ApprovedBy,
		CreatedAt:      t.CreatedAt,
		UpdatedAt:      t.UpdatedAt,
	}
}
