package models

// ParticipantProgress es la fila de respuesta de
// GET /api/docente/cursos/{cursoId}/progreso: para un participante
// inscripto, cuántas de las actividades del curso tienen asistencia
// presente=true.
type ParticipantProgress struct {
	ParticipanteID       string  `json:"participanteId"`
	ActividadesTotal     int     `json:"actividadesTotal"`
	ActividadesAsistidas int     `json:"actividadesAsistidas"`
	PorcentajeAsistencia float64 `json:"porcentajeAsistencia"`
}
