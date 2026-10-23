package models

import "time"

// RegistroAuditoria es una fila de GET /api/auditoria (solo
// administración). UsuarioNombre se resuelve contra auth-service y queda
// vacío si no se pudo.
type RegistroAuditoria struct {
	ID            string    `json:"id"`
	UsuarioID     string    `json:"usuarioId"`
	UsuarioNombre string    `json:"usuarioNombre,omitempty"`
	Accion        string    `json:"accion"`
	Recurso       string    `json:"recurso"`
	RecursoID     string    `json:"recursoId"`
	Fecha         time.Time `json:"fecha"`
	IP            string    `json:"ip,omitempty"`
}
