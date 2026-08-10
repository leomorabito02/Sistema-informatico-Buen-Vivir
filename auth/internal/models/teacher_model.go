package models

import (
	"time"

	"gorm.io/gorm"
)

type Teacher struct {
	TeacherID      string    `json:"teacher_id" gorm:"primaryKey"`
	UserID         string    `gorm:"not null" json:"user_id"`
	User           *User     `json:"user,omitempty"`
	Specialization string    `json:"specialization"`
	IsExternal     bool      `json:"is_external"`
	ApprovalStatus string     `json:"approval_status" gorm:"not null"`
	ApprovalDate   *time.Time `json:"approval_date"`
	CreatedAt      time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
	ApprovedBy     *string    `json:"approved_by" gorm:"index"`
	Approver       *User      `gorm:"foreignKey:ApprovedBy" json:"approver,omitempty"`
}
