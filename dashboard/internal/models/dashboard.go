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

// ── Resumen ya calculado (GET /api/v1/dashboard/resumen) ───────────────

// Indicador es un número del dashboard (HU-52). Valor es nil y Disponible
// false cuando el servicio de origen no respondió.
type Indicador struct {
	ID          string `json:"id"`
	Titulo      string `json:"titulo"`
	Descripcion string `json:"descripcion"`
	Modulo      string `json:"modulo"` // "usuarios" | "educacion" | "promocion" | "insercion"
	Fuente      string `json:"fuente"` // "autenticacion" | "educacion" | "promocion" | "insercion"
	Valor       *int   `json:"valor"`
	Disponible  bool   `json:"disponible"`
}

// PuntoSerie es el valor de un mes ("AAAA-MM").
type PuntoSerie struct {
	Mes   string `json:"mes"`
	Valor int    `json:"valor"`
}

type SerieGrafica struct {
	Nombre string       `json:"nombre"`
	Puntos []PuntoSerie `json:"puntos"`
}

// Grafica es la evolución mensual de uno o más indicadores (HU-53).
// Series es nil (null en JSON) cuando el servicio de origen no respondió.
type Grafica struct {
	ID          string         `json:"id"`
	Titulo      string         `json:"titulo"`
	Tipo        string         `json:"tipo"` // "line" | "bar"
	Modulo      string         `json:"modulo"`
	Indicadores []string       `json:"indicadores"`
	Series      []SerieGrafica `json:"series"`
}

// Resumen son los indicadores y gráficas para los filtros pedidos.
type Resumen struct {
	Indicadores []Indicador `json:"indicadores"`
	Graficas    []Grafica   `json:"graficas"`
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

// ── Promoción Humana (indicadores agregados por el equipo) ─────────────

// Promocion: fichas e hitos para "Fichas activas" y "Participantes con
// seguimiento". Solo ids, estados y fechas: Promoción Humana nunca manda
// el contenido confidencial de fichas ni hitos.
type Promocion struct {
	Fichas []FichaPromocion `json:"fichas"`
	Hitos  []HitoPromocion  `json:"hitos"`
}

type FichaPromocion struct {
	ID             string     `json:"id"`
	ParticipanteID string     `json:"participanteId"`
	Estado         string     `json:"estado"` // "activa" | "baja"
	FechaAlta      time.Time  `json:"fechaAlta"`
	FechaBaja      *time.Time `json:"fechaBaja"`
}

type HitoPromocion struct {
	ParticipanteID string `json:"participanteId"`
	Fecha          string `json:"fecha"` // AAAA-MM-DD
}

// Promoción Humana: GET /api/v1/human-promotion/indicadores.
type IndicadoresPH struct {
	Fichas []struct {
		ID             string     `json:"id"`
		ParticipanteID string     `json:"participante_id"`
		Estado         string     `json:"estado"`
		CreadaEn       time.Time  `json:"creada_en"`
		FechaBaja      *time.Time `json:"fecha_baja"`
	} `json:"fichas"`
	Hitos []struct {
		ParticipanteID string `json:"participante_id"`
		Fecha          string `json:"fecha"`
	} `json:"hitos"`
}
