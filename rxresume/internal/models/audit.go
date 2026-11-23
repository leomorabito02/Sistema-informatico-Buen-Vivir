package models

import "time"

// AuditEntry registra quién vio, editó o cambió la foto del CV de otra persona, cuándo
// y desde qué IP. No guarda el contenido del CV.
type AuditEntry struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UsuarioID string    `gorm:"size:64;index" json:"usuario_id"`
	Rol       string    `gorm:"size:32" json:"rol,omitempty"`
	Accion    string    `gorm:"size:32;index" json:"accion"`
	Recurso   string    `gorm:"size:32" json:"recurso"`
	RecursoID string    `gorm:"size:64" json:"recurso_id"`
	Fecha     time.Time `gorm:"index" json:"fecha"`
	IP        string    `gorm:"size:64" json:"ip,omitempty"`
}

// TableName: tabla propia del servicio.
func (AuditEntry) TableName() string { return "auditoria_rxresume" }
