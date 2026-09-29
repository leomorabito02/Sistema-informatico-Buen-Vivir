package models

import "time"

// JobApplication represents the bridge table/document between JobOffer and Participant.
type JobApplication struct {
	ID              string    `firestore:"id,omitempty" json:"id"`
	ParticipantID   string    `firestore:"participante_id" json:"participante_id"`
	OfferID         string    `firestore:"oferta_id" json:"oferta_id"`
	Status          string    `firestore:"estado" json:"estado"`
	ApplicationDate time.Time `firestore:"fecha_postulacion" json:"fecha_postulacion"`

	CreatedAt time.Time `firestore:"created_at" json:"created_at"`
	UpdatedAt time.Time `firestore:"updated_at" json:"updated_at"`
}
