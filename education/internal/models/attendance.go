package models

import "time"

// Attendance representa un documento de la colección Firestore
// "asistencias". El id del documento es determinístico
// ({actividadId}_{participanteId}, ver AttendanceRepository.RegisterBatch)
// para que volver a registrar asistencia del mismo participante en la
// misma actividad actualice el registro en vez de duplicarlo.
type Attendance struct {
	ID             string    `json:"id" firestore:"-"`
	ParticipanteID string    `json:"participanteId" firestore:"participante_id"`
	ActividadID    string    `json:"actividadId" firestore:"actividad_id"`
	Presente       bool      `json:"presente" firestore:"presente"`
	Fecha          time.Time `json:"fecha" firestore:"fecha"`
}

// AttendanceInput es un ítem del array que recibe
// POST /api/docente/actividades/{actividadId}/asistencia.
type AttendanceInput struct {
	ParticipanteID string `json:"participanteId"`
	Presente       bool   `json:"presente"`
}
