package models

import "time"

// Attendance representa un documento de la colección Firestore
// "asistencias". No se carga a mano: cada documento lo genera la
// entrega de una actividad (ver SubmissionRepository.CreateWithAttendance),
// con presente=true y Fecha = fecha de la actividad (no la de la
// entrega). El id del documento es determinístico
// ({actividadId}_{participanteId}), el mismo que el de su Entrega.
type Attendance struct {
	ID             string    `json:"id" firestore:"-"`
	ParticipanteID string    `json:"participanteId" firestore:"participante_id"`
	ActividadID    string    `json:"actividadId" firestore:"actividad_id"`
	Presente       bool      `json:"presente" firestore:"presente"`
	Fecha          time.Time `json:"fecha" firestore:"fecha"`
}
