package models

import "time"

// LoginAttempt cuenta los intentos fallidos de contraseña por email
// (HU-02). La clave es el hash del email en minúsculas: la tabla no
// guarda el email en claro. Vive en la base (y no en memoria) porque
// Cloud Run puede tener varias instancias o reiniciarse.
type LoginAttempt struct {
	EmailHash   string     `gorm:"primaryKey;column:email_hash"`
	Failed      int        `gorm:"column:failed;not null;default:0"`
	LockedUntil *time.Time `gorm:"column:locked_until"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (LoginAttempt) TableName() string { return "login_attempts" }

// AuditLog es un registro de auditoría de Autenticación (HU-10, y los
// rechazos de HU-03 y HU-06). Mismo formato que el de Educación y
// Promoción Humana: quién, qué acción, sobre qué recurso, cuándo y desde
// qué IP. Campos lista solo los NOMBRES de los campos modificados, nunca
// sus valores (son datos personales). Es inmutable: el repositorio solo
// crea y lista, y la migración agrega un trigger que rechaza UPDATE y
// DELETE.
type AuditLog struct {
	ID        uint      `gorm:"primaryKey;autoIncrement;column:id"`
	UsuarioID string    `gorm:"column:usuario_id;index"`
	Rol       string    `gorm:"column:rol"`
	Accion    string    `gorm:"column:accion;index;not null"`
	Recurso   string    `gorm:"column:recurso;index;not null"`
	RecursoID string    `gorm:"column:recurso_id"`
	Campos    string    `gorm:"column:campos"`
	Resultado string    `gorm:"column:resultado;not null"`
	IP        string    `gorm:"column:ip"`
	Fecha     time.Time `gorm:"column:fecha;index;not null"`
}

func (AuditLog) TableName() string { return "auditoria" }
