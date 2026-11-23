package models

import "time"

// Attendance es un documento de "asistencias". Lo genera cada entrega (presente=true,
// con la fecha de la actividad). Id: {actividadId}_{participanteId}, igual que su entrega.
type Attendance struct {
	ID             string    `json:"id" firestore:"-"`
	ParticipanteID string    `json:"participanteId" firestore:"participante_id"`
	ActividadID    string    `json:"actividadId" firestore:"actividad_id"`
	Presente       bool      `json:"presente" firestore:"presente"`
	Fecha          time.Time `json:"fecha" firestore:"fecha"`
}
