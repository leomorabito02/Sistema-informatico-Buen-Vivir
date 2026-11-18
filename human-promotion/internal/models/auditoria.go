package models

import "time"

// RegistroAuditoria es una fila de GET /api/v1/human-promotion/auditoria
// (HU-35, solo administración). Dice quién tocó qué ficha o hito y
// cuándo, nunca el contenido. UsuarioNombre se resuelve contra Auth y
// queda vacío si no se pudo.
type RegistroAuditoria struct {
	ID            string    `json:"id"`
	UsuarioID     string    `json:"usuario_id"`
	UsuarioNombre string    `json:"usuario_nombre,omitempty"`
	Accion        string    `json:"accion"`
	Recurso       string    `json:"recurso"`
	RecursoID     string    `json:"recurso_id"`
	Fecha         time.Time `json:"fecha"`
	IP            string    `json:"ip,omitempty"`
}
