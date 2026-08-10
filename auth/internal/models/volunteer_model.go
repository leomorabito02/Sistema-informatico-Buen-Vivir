package models

import (
	"time"

	"gorm.io/gorm"
)

type Volunteer struct {
	VolunteerID string    `json:"volunteer_id" gorm:"primaryKey"`
	UserID        string    `json:"user_id" gorm:"not null;unique"`
	User          *User     `json:"user,omitempty"`
	ApprovalStatus string     `json:"approval_status" gorm:"not null"`
	ApprovalDate   *time.Time `json:"approval_date"`
	CreatedAt      time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
	ApprovedBy     *string    `json:"approved_by" gorm:"index"`
	Approver       *User      `gorm:"foreignKey:ApprovedBy" json:"approver,omitempty"`
}

