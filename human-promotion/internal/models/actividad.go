package models

// EntradaActividad es un hecho de la actividad del participante en el
// centro que no carga el equipo: llega por eventos de Educación
// (inscripciones, cursos completados y asistencia). Se muestra mezclada
// con los hitos en la línea de tiempo (HU-33, HU-34). Sin números ni
// porcentajes: la asistencia se agrupa por mes, sin contar los días.
type EntradaActividad struct {
	ID          string `json:"id"`
	Tipo        string `json:"tipo"`
	Fecha       string `json:"fecha"`
	Descripcion string `json:"descripcion"`
	// CursoID permite que el frontend muestre el nombre del curso (vive
	// en Educación). Vacío en la asistencia al centro.
	CursoID string `json:"curso_id,omitempty"`
}

// LineaDeTiempoPropia es la respuesta de /mi-linea-de-tiempo (HU-34):
// los hitos visibles de la ficha más la actividad en el centro.
type LineaDeTiempoPropia struct {
	Hitos     []HitoParticipante `json:"hitos"`
	Actividad []EntradaActividad `json:"actividad"`
}
