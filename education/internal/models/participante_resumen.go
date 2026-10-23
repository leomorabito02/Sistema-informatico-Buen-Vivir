package models

// ParticipanteResumen es un resultado del buscador de participantes
// (GET /api/participantes?q=): lo justo para elegirlo en un selector.
type ParticipanteResumen struct {
	ParticipanteID string `json:"participanteId"`
	Nombre         string `json:"nombre"`
	Apellido       string `json:"apellido"`
}
