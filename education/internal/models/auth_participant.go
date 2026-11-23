package models

// AuthParticipant es un participante tal como lo devuelve auth-service.
// ParticipantID (no UserID) es el participanteId de las colecciones de education.
type AuthParticipant struct {
	ParticipantID  string `json:"participant_id"`
	UserID         string `json:"user_id"`
	DNI            string `json:"dni,omitempty"`
	ApprovalStatus string `json:"approval_status"`
}

// AuthUser es un usuario tal como lo devuelve auth-service en GET /api/v1/users.
// Se usa para mostrar nombre y apellido de los participantes.
type AuthUser struct {
	UserID    string `json:"user_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}
