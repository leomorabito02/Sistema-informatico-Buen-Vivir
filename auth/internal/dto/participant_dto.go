package dto

import (
	"time"

	"auth/internal/models"
)

type CreateParticipantRequest struct {
	UserID         string    `json:"user_id" validate:"required"`
	Birthdate      time.Time `json:"birthdate" validate:"required"`
	DNI            string    `json:"dni" validate:"omitempty"`
	Address        string    `json:"address" validate:"omitempty"`
	Gender         string    `json:"gender" validate:"required"`
	EmergencyContactName string `json:"emergency_contact_name" validate:"omitempty"`
	EmergencyContactPhone string `json:"emergency_contact_phone" validate:"omitempty"`
	ImageURL       string    `json:"image_url" validate:"omitempty,url"`
	DocURL         string    `json:"doc_url" validate:"omitempty,url"`
	ApprovalStatus string    `json:"approval_status" validate:"omitempty,oneof=pending approved rejected"`
}

type UpdateParticipantRequest struct {
	Birthdate      *time.Time `json:"birthdate" validate:"omitempty"`
	DNI            string     `json:"dni" validate:"omitempty"`
	Address        string     `json:"address" validate:"omitempty"`
	Gender         string     `json:"gender" validate:"omitempty"`
	EmergencyContactName string `json:"emergency_contact_name" validate:"omitempty"`
	EmergencyContactPhone string `json:"emergency_contact_phone" validate:"omitempty"`
	ImageURL       string     `json:"image_url" validate:"omitempty,url"`
	DocURL         string     `json:"doc_url" validate:"omitempty,url"`
	ApprovalStatus string     `json:"approval_status" validate:"omitempty,oneof=pending approved rejected"`
	ApprovedBy     string     `json:"approved_by" validate:"omitempty"`
}

type ParticipantResponse struct {
	ParticipantID  string     `json:"participant_id"`
	UserID         string     `json:"user_id"`
	Birthdate      time.Time  `json:"birthdate"`
	DNI            string     `json:"dni,omitempty"`
	Address        string     `json:"address,omitempty"`
	Gender         string     `json:"gender"`
	EmergencyContactName string `json:"emergency_contact_name,omitempty"`
	EmergencyContactPhone string `json:"emergency_contact_phone,omitempty"`
	ImageURL       string     `json:"image_url"`
	DocURL         string     `json:"doc_url"`
	ApprovalStatus string     `json:"approval_status"`
	ApprovalDate   *time.Time `json:"approval_date,omitempty"`
	ApprovedBy     *string    `json:"approved_by,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func ToParticipantResponse(p *models.Participant) *ParticipantResponse {
	return &ParticipantResponse{
		ParticipantID:  p.ParticipantID,
		UserID:         p.UserID,
		Birthdate:      p.Birthdate,
		DNI:            p.DNI,
		Address:        p.Address,
		Gender:         p.Gender,
		EmergencyContactName: p.EmergencyContactName,
		EmergencyContactPhone: p.EmergencyContactPhone,
		ImageURL:       p.ImageURL,
		DocURL:         p.DocURL,
		ApprovalStatus: p.ApprovalStatus,
		ApprovalDate:   p.ApprovalDate,
		ApprovedBy:     p.ApprovedBy,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}
