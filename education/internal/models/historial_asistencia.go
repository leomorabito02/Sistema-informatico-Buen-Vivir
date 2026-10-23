package models

import "time"

// HistorialCurso es el historial de asistencia de un participante en un
// curso (HU-19): una fila por actividad del curso, presente o no.
// Responde GET /api/participants/{id}/asistencia, un elemento por
// inscripción.
type HistorialCurso struct {
	CursoID              string               `json:"cursoId"`
	CursoTitulo          string               `json:"cursoTitulo"`
	EstadoInscripcion    string               `json:"estadoInscripcion"`
	FechaInscripcion     time.Time            `json:"fechaInscripcion"`
	ActividadesTotal     int                  `json:"actividadesTotal"`
	ActividadesAsistidas int                  `json:"actividadesAsistidas"`
	PorcentajeAsistencia float64              `json:"porcentajeAsistencia"`
	Actividades          []HistorialActividad `json:"actividades"`
}

// HistorialActividad es una actividad del curso y si el participante
// asistió (la asistencia la genera la entrega).
type HistorialActividad struct {
	ActividadID string    `json:"actividadId"`
	Titulo      string    `json:"titulo"`
	Fecha       time.Time `json:"fecha"`
	Presente    bool      `json:"presente"`
}
