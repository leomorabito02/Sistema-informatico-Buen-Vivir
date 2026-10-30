package models

// AuthParticipant es la forma (recortada a lo que education necesita)
// de un participante tal como lo devuelve auth-service en
// GET /api/v1/participants y GET /api/v1/participants/{id}. No es una
// entidad propia de education — Participante vive en la base de auth
// (Postgres), ver DER — se define acá solo para decodificar esa
// respuesta HTTP. `ParticipantID` es el id que se usa como
// participanteId en las colecciones de Firestore de education
// (Inscripcion, Asistencia, AsistenciaGeneral), no `UserID`.
type AuthParticipant struct {
	ParticipantID  string `json:"participant_id"`
	UserID         string `json:"user_id"`
	DNI            string `json:"dni,omitempty"`
	ApprovalStatus string `json:"approval_status"`
}

// AuthUser es la forma (recortada) de un usuario tal como lo devuelve
// auth-service en GET /api/v1/users. Participante no tiene nombre
// propio — vive en Usuario (ver DER) — así que GeneralAttendanceService
// cruza esto contra AuthParticipant.UserID para poder mostrar
// nombre/apellido en la asistencia general en vez de solo un id.
type AuthUser struct {
	UserID    string `json:"user_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}
