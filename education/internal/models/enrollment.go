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
	// MotivoBaja y FechaBaja se completan cuando la inscripción pasa a
	// "baja" por un evento (por ejemplo, el participante se dio de baja
	// en auth-service). Vacíos en las inscripciones activas.
	MotivoBaja string     `json:"motivoBaja,omitempty" firestore:"motivo_baja,omitempty"`
	FechaBaja  *time.Time `json:"fechaBaja,omitempty" firestore:"fecha_baja,omitempty"`
}

// MotivoBajaParticipanteEliminado es el motivo que se guarda cuando la
// baja viene de usuarios.participante.eliminado (o de una cuenta
// desactivada en usuarios.usuario.actualizado).
const MotivoBajaParticipanteEliminado = "participante_dado_de_baja"

const (
	EnrollmentEstadoActiva     = "activa"
	EnrollmentEstadoFinalizada = "finalizada"
	EnrollmentEstadoBaja       = "baja"
)
