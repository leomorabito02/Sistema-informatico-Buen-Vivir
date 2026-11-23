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

// AuditLog es un registro de auditoría de Autenticación (HU-10, HU-03, HU-06). Campos
// guarda solo los nombres de los campos modificados. Es inmutable: un trigger rechaza
// UPDATE y DELETE.
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

// TwofaChallengeRecord persists active 2FA challenges across stateless instances.
type TwofaChallengeRecord struct {
	Nonce       string    `gorm:"column:nonce;type:varchar(128);primaryKey"`
	Email       string    `gorm:"column:email;type:varchar(255);index;not null"`
	UserIDsJSON string    `gorm:"column:user_ids_json;type:text;not null"`
	Method      string    `gorm:"column:method;type:varchar(16);not null"`
	Code        string    `gorm:"column:code;type:varchar(64)"`
	Secret      string    `gorm:"column:secret;type:text"`
	Attempts    int       `gorm:"column:attempts;not null;default:0"`
	ExpiresAt   time.Time `gorm:"column:expires_at;index;not null"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (TwofaChallengeRecord) TableName() string { return "twofa_challenges" }
