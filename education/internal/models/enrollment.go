package models

import "time"

// Enrollment representa un documento de la colección Firestore
// "inscripciones". participante_id referencia a un usuario de
// auth-service sin FK física (microservicios de bajo acoplamiento).
type Enrollment struct {
	ID               string    `json:"id" firestore:"-"`
	ParticipanteID   string    `json:"participanteId" firestore:"participante_id"`
	CursoID          string    `json:"cursoId" firestore:"curso_id"`
	FechaInscripcion time.Time `json:"fechaInscripcion" firestore:"fecha_inscripcion"`
	Estado           string    `json:"estado" firestore:"estado"`
}

const (
	EnrollmentEstadoActiva     = "activa"
	EnrollmentEstadoFinalizada = "finalizada"
	EnrollmentEstadoBaja       = "baja"
)
