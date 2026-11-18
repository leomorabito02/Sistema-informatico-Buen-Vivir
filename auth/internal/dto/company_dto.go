package dto

import (
	"time"
	"auth/internal/models"
)

type CreateCompanyRequest struct {
	UserID         string `json:"user_id" validate:"required"`
	BusinessName   string `json:"business_name" validate:"required,min=2,max=100"`
	Cuit           string `json:"cuit" validate:"required"`
	ApprovalStatus string `json:"approval_status" validate:"omitempty"` // e.g., "pending", "approved", "rejected"
}

type UpdateCompanyRequest struct {
	BusinessName   string `json:"business_name" validate:"omitempty,min=2,max=100"`
	ApprovalStatus string `json:"approval_status" validate:"omitempty,oneof=pending approved"`
	ApprovedBy     string `json:"approved_by" validate:"omitempty"`
}

type CompanyResponse struct {
	CompanyID      string     `json:"company_id"`
	UserID         string     `json:"user_id"`
	BusinessName   string     `json:"business_name"`
	Cuit           string     `json:"cuit"`
	ApprovalStatus string     `json:"approval_status"`
	ApprovalDate   *time.Time `json:"approval_date,omitempty"`
	ApprovedBy     *string    `json:"approved_by,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func ToCompanyResponse(c *models.Company) *CompanyResponse {
	return &CompanyResponse{
		CompanyID:      c.CompanyID,
		UserID:         c.UserID,
		BusinessName:   c.BusinessName,
		Cuit:           c.Cuit,
		ApprovalStatus: c.ApprovalStatus,
		ApprovalDate:   c.ApprovalDate,
		ApprovedBy:     c.ApprovedBy,
		CreatedAt:      c.CreatedAt,
		UpdatedAt:      c.UpdatedAt,
	}
}
