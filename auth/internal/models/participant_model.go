package models

import (
	"time"

	"gorm.io/gorm"
)


type Participant struct {
	ParticipantID string    `json:"participant_id" gorm:"primaryKey"`
	UserID         string    `json:"user_id" gorm:"not null;unique"`
	User           *User     `json:"user,omitempty"`
	Birthdate      time.Time `json:"birthdate"`
	DNI            string    `json:"dni"`
	Address        string    `json:"address"`
	Gender         string    `json:"gender"`
	ImageURL       string    `json:"image_url"`
	Documents      []ParticipantDocument `json:"documents,omitempty" gorm:"foreignKey:ParticipantID;references:ParticipantID"`
	EmergencyContactName string `json:"emergency_contact_name"`
	EmergencyContactPhone string `json:"emergency_contact_phone"`
	ApprovalStatus string     `json:"approval_status" gorm:"not null"`
	ApprovalDate   *time.Time `json:"approval_date"`
	CreatedAt      time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
	ApprovedBy     *string    `json:"approved_by" gorm:"index"`
	Approver       *User      `gorm:"foreignKey:ApprovedBy" json:"approver,omitempty"`
}