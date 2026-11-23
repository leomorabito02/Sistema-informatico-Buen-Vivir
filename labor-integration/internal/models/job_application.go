package models

import "time"

// JobApplication represents the bridge table/document between JobOffer and Participant.
type JobApplication struct {
	ID string `firestore:"id,omitempty" json:"id"`
	// ParticipantID es el id de participante de Auth (part-...), el mismo de Educación y
	// Promoción Humana. UserID es el id de usuario (usr-...), el del token y el del CV. Las
	// postulaciones viejas no tienen UserID y guardaban el id de usuario en ParticipantID.
	ParticipantID string `firestore:"participante_id" json:"participante_id"`
	UserID        string `firestore:"usuario_id,omitempty" json:"usuario_id,omitempty"`
	// Nombre del postulante: no se guarda, se completa al listar los postulados de una oferta.
	Nombre          string    `firestore:"-" json:"nombre,omitempty"`
	OfferID         string    `firestore:"oferta_id" json:"oferta_id"`
	Status          string    `firestore:"estado" json:"estado"`
	ApplicationDate time.Time `firestore:"fecha_postulacion" json:"fecha_postulacion"`

	CreatedAt time.Time `firestore:"created_at" json:"created_at"`
	UpdatedAt time.Time `firestore:"updated_at" json:"updated_at"`
}
