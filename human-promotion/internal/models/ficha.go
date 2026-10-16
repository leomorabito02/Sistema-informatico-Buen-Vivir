package models

import "time"

// FichaProcesoPersonal es la ficha confidencial de un participante
// (relación 1:1). SituacionActual y Avances viajan en texto plano
// dentro del servicio y se guardan cifrados (ver FichaDocument).
type FichaProcesoPersonal struct {
	ID              string    `json:"id"`
	ParticipanteID  string    `json:"participante_id"`
	SituacionActual string    `json:"situacion_actual"`
	Avances         string    `json:"avances"`
	CreadoPor       string    `json:"creado_por"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// FichaDocument es la forma persistida en Firestore: los campos
// confidenciales ya cifrados con AES-256-GCM.
type FichaDocument struct {
	ParticipanteID  string    `firestore:"participante_id"`
	SituacionActual string    `firestore:"situacion_actual"`
	Avances         string    `firestore:"avances"`
	CreadoPor       string    `firestore:"creado_por"`
	UpdatedAt       time.Time `firestore:"updated_at"`
	// Estado y FechaBaja: baja lógica cuando el participante se da de
	// baja en auth-service (evento usuarios.participante.eliminado). La
	// ficha no se borra: se conserva el historial.
	Estado    string     `firestore:"estado,omitempty"`
	FechaBaja *time.Time `firestore:"fecha_baja,omitempty"`
}

// FichaEstadoBaja es el estado de una ficha dada de baja.
const FichaEstadoBaja = "baja"

// CreateFichaInput es el body de POST /fichas.
type CreateFichaInput struct {
	ParticipanteID  string `json:"participante_id"`
	SituacionActual string `json:"situacion_actual"`
	Avances         string `json:"avances"`
}

// UpdateFichaInput es el body de PUT /fichas/{id}. Ambos campos son
// obligatorios también al editar (HU-31).
type UpdateFichaInput struct {
	SituacionActual string `json:"situacion_actual"`
	Avances         string `json:"avances"`
}
