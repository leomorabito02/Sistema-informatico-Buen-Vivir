package models

import "time"

// GeneralAttendance es un documento de "asistencia_general" (HU-19bis). Id:
// {fecha}_{participanteId}, así recargar el mismo día actualiza en vez de duplicar.
// Fecha es "YYYY-MM-DD" para evitar problemas de huso horario.
type GeneralAttendance struct {
	ID             string    `json:"id" firestore:"-"`
	ParticipanteID string    `json:"participanteId" firestore:"participante_id"`
	Fecha          string    `json:"fecha" firestore:"fecha"`
	Presente       bool      `json:"presente" firestore:"presente"`
	Observacion    string    `json:"observacion,omitempty" firestore:"observacion"`
	UpdatedAt      time.Time `json:"updatedAt" firestore:"updated_at"`
}

// GeneralAttendanceInput es un ítem del array que recibe
// POST /api/asistencia-general.
type GeneralAttendanceInput struct {
	ParticipanteID string `json:"participanteId"`
	Presente       bool   `json:"presente"`
	Observacion    string `json:"observacion,omitempty"`
}

// GeneralAttendanceView es una fila de GET /api/asistencia-general. Presente es nil
// si todavía no se cargó (false es "ausente"). Nombre y Apellido pueden venir vacíos.
type GeneralAttendanceView struct {
	ParticipanteID string `json:"participanteId"`
	Nombre         string `json:"nombre,omitempty"`
	Apellido       string `json:"apellido,omitempty"`
	Presente       *bool  `json:"presente"`
	Observacion    string `json:"observacion,omitempty"`
}
