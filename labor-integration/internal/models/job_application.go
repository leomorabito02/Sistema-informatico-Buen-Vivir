package models

import "time"

// JobApplication represents the bridge table/document between JobOffer and Participant.
type JobApplication struct {
	ID string `firestore:"id,omitempty" json:"id"`
	// ParticipantID es el id de participante de Auth (part-...), el mismo
	// que usan Educación y Promoción Humana. UserID es su id de usuario
	// (usr-...), que es el que viaja en el token y el que usa RxResume
	// para el CV. Las postulaciones anteriores al 29/10/2026 no tienen
	// UserID y guardaban el id de usuario en ParticipantID.
	ParticipantID string `firestore:"participante_id" json:"participante_id"`
	UserID        string `firestore:"usuario_id,omitempty" json:"usuario_id,omitempty"`
	// Nombre del postulante: no se guarda, se completa al listar los
	// postulados de una oferta (22/11/2026; antes la pantalla quedaba en
	// "Cargando..." si la persona no tenía CV).
	Nombre          string    `firestore:"-" json:"nombre,omitempty"`
	OfferID         string    `firestore:"oferta_id" json:"oferta_id"`
	Status          string    `firestore:"estado" json:"estado"`
	ApplicationDate time.Time `firestore:"fecha_postulacion" json:"fecha_postulacion"`

	CreatedAt time.Time `firestore:"created_at" json:"created_at"`
	UpdatedAt time.Time `firestore:"updated_at" json:"updated_at"`
}
