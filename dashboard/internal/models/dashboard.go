// Package models tiene lo que el Dashboard devuelve al frontend (en
// castellano, igual que src/app/core/models/dashboard.model.ts) y lo que
// lee de cada servicio de origen.
package models

import "time"

// ── Respuestas del Dashboard ───────────────────────────────────────────

// Participante: un participante aprobado, con sus fechas de alta y baja.
type Participante struct {
	ID        string     `json:"id"`
	Rol       string     `json:"rol"`    // siempre "participante"
	Estado    string     `json:"estado"` // "activo" | "baja"
	FechaAlta time.Time  `json:"fechaAlta"`
	FechaBaja *time.Time `json:"fechaBaja"`
}

type Actividad struct {
	ID    string    `json:"id"`
	Fecha time.Time `json:"fecha"`
}

type Curso struct {
	ID            string      `json:"id"`
	Titulo        string      `json:"titulo"`
	AreaID        string      `json:"areaId"`
	Estado        string      `json:"estado"` // "activo" (publicado) | "borrador"
	FechaCreacion time.Time   `json:"fechaCreacion"`
	Actividades   []Actividad `json:"actividades"`
}

type ParticipanteHabilidad struct {
	ID             string    `json:"id"`
	ParticipanteID string    `json:"participanteId"`
	CursoID        string    `json:"cursoId"`
	AreaID         string    `json:"areaId"`
	Origen         string    `json:"origen"` // "actividad" | "manual"
	Estado         string    `json:"estado"` // siempre "aceptada"
	Fecha          time.Time `json:"fecha"`
}

type Postulacion struct {
	ID             string    `json:"id"`
	ParticipanteID string    `json:"participanteId"`
	Estado         string    `json:"estado"` // "postulado" | "en_revision" | "seleccionado" | "rechazado"
	FechaEstado    time.Time `json:"fechaEstado"`
}

type Area struct {
	ID     string `json:"id"`
	Nombre string `json:"nombre"`
}

// ── Lo que devuelve cada servicio de origen ────────────────────────────

// Autenticación: GET /api/v1/users/participants/history, dentro del
// sobre {success, data, errors}.
type AuthResponse[T any] struct {
	Success bool `json:"success"`
	Data    T    `json:"data"`
}

type ParticipantHistory struct {
	UserID         string     `json:"user_id"`
	Status         string     `json:"status"`
	ApprovalStatus string     `json:"approval_status"`
	ApprovalDate   *time.Time `json:"approval_date"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at"`
}

// Educación: GET /api/cursos y /api/cursos/borradores.
type Course struct {
	ID           string        `json:"id"`
	Title        string        `json:"title"`
	TrainingArea *TrainingArea `json:"trainingArea"`
	CreatedAt    string        `json:"createdAt"`
}

// Educación: GET /api/areas-formativas.
type TrainingArea struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Educación: GET /api/cursos/{id}/actividades.
type Activity struct {
	ID    string    `json:"id"`
	Fecha time.Time `json:"fecha"`
}

// Educación: GET /api/habilidades/otorgadas.
type GrantedSkill struct {
	ParticipanteID string    `json:"participanteId"`
	HabilidadID    string    `json:"habilidadId"`
	Origen         string    `json:"origen"`
	Fecha          time.Time `json:"fecha"`
	CursoID        string    `json:"cursoId"`
}

// Inserción Laboral: GET /api/v1/labor-integration/offers.
type JobOffer struct {
	ID string `json:"id"`
}

// Inserción Laboral: GET /api/v1/labor-integration/offers/{id}/applications.
type JobApplication struct {
	ID              string    `json:"id"`
	ParticipantID   string    `json:"participante_id"`
	Status          string    `json:"estado"`
	ApplicationDate time.Time `json:"fecha_postulacion"`
	UpdatedAt       time.Time `json:"updated_at"`
}
