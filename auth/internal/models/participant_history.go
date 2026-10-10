package models

import "time"

// ParticipantHistory es la historia de alta y baja de un participante,
// incluidos los dados de baja (soft delete). La usa el Dashboard para
// saber cuántos participantes estaban activos en cada mes.
type ParticipantHistory struct {
	UserID         string     `gorm:"column:user_id"`
	Status         string     `gorm:"column:status"`
	ApprovalStatus string     `gorm:"column:approval_status"`
	ApprovalDate   *time.Time `gorm:"column:approval_date"`
	CreatedAt      time.Time  `gorm:"column:created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at"`
	DeletedAt      *time.Time `gorm:"column:deleted_at"`
}
