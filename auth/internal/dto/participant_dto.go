package dto

import (
	"time"

	"auth/internal/models"
)

type CreateParticipantRequest struct {
	UserID         string    `json:"user_id" validate:"required"`
	Birthdate      time.Time `json:"birthdate" validate:"required"`
	DNI            string    `json:"dni" validate:"required,dni"`
	Address        string    `json:"address" validate:"required,min=3"`
	Gender         string    `json:"gender" validate:"required"`
	EmergencyContactName string `json:"emergency_contact_name" validate:"omitempty"`
	EmergencyContactPhone string `json:"emergency_contact_phone" validate:"omitempty"`
	ImageURL       string    `json:"image_url" validate:"omitempty,url_or_relative"`
	ApprovalStatus string    `json:"approval_status" validate:"omitempty,oneof=pending approved"`
}

type UpdateParticipantRequest struct {
	Birthdate      *time.Time `json:"birthdate" validate:"omitempty"`
	DNI            string     `json:"dni" validate:"omitempty"`
	Address        string     `json:"address" validate:"omitempty"`
	Gender         string     `json:"gender" validate:"omitempty"`
	EmergencyContactName string `json:"emergency_contact_name" validate:"omitempty"`
	EmergencyContactPhone string `json:"emergency_contact_phone" validate:"omitempty"`
	ImageURL       string     `json:"image_url" validate:"omitempty,url_or_relative"`
	ApprovalStatus string `json:"approval_status" validate:"omitempty,oneof=pending approved"`
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
	Documents      []ParticipantDocumentResponse `json:"documents,omitempty"`
	ApprovalStatus string     `json:"approval_status"`
	ApprovalDate   *time.Time `json:"approval_date,omitempty"`
	ApprovedBy     *string    `json:"approved_by,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type ParticipantDocumentResponse struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	DocURL        string    `json:"doc_url"`
	FileType      string    `json:"file_type"`
	FileSize      int64     `json:"file_size"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func ToParticipantResponse(p *models.Participant) *ParticipantResponse {
	docs := make([]ParticipantDocumentResponse, 0, len(p.Documents))
	for _, doc := range p.Documents {
		docs = append(docs, ParticipantDocumentResponse{
			ID:        doc.ID,
			Name:      doc.Name,
			DocURL:    doc.DocURL,
			FileType:  doc.FileType,
			FileSize:  doc.FileSize,
			CreatedAt: doc.CreatedAt,
			UpdatedAt: doc.UpdatedAt,
		})
	}

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
		Documents:      docs,
		ApprovalStatus: p.ApprovalStatus,
		ApprovalDate:   p.ApprovalDate,
		ApprovedBy:     p.ApprovedBy,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}
