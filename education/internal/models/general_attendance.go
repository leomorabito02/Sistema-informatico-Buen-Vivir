package models

import "time"

// GeneralAttendance representa un documento de la colección Firestore
// "asistencia_general" (HU-19bis/RF-64): control diario de concurrencia
// del centro, independiente de la asistencia por actividad/curso (ver
// Attendance). El id del documento es determinístico
// "{fecha}_{participanteId}" (mismo criterio que
// AttendanceRepository.RegisterBatch) para que volver a cargar el mismo
// participante en la misma fecha actualice el registro existente en
// vez de duplicarlo.
//
// Fecha se guarda como string "YYYY-MM-DD" (no time.Time): lo que
// importa acá es el día calendario elegido en la pestaña de asistencia,
// no un instante — un string permite el id determinístico y el filtro
// `Where("fecha", "==", ...)` sin ambigüedad de huso horario.
type GeneralAttendance struct {
	ID             string    `json:"id" firestore:"-"`
	ParticipanteID string    `json:"participanteId" firestore:"participante_id"`
	Fecha          string    `json:"fecha" firestore:"fecha"`
	Presente       bool      `json:"presente" firestore:"presente"`
	Observacion    string    `json:"observacion,omitempty" firestore:"observacion"`
	UpdatedAt      time.Time `json:"updatedAt" firestore:"updated_at"`
}

// GeneralAttendanceInput es un ítem del array que recibe
// POST /api/asistencia-general.
type GeneralAttendanceInput struct {
	ParticipanteID string `json:"participanteId"`
	Presente       bool   `json:"presente"`
	Observacion    string `json:"observacion,omitempty"`
}

// GeneralAttendanceView es la fila que devuelve GET
// /api/asistencia-general: un participante del centro (auth-service,
// Participante + su Usuario asociado) + su asistencia para la fecha
// pedida, si ya se cargó. Presente queda en nil cuando todavía no hay
// registro para ese participante en esa fecha (a diferencia de false,
// que es "se cargó y marcó ausente"). Nombre/Apellido pueden venir
// vacíos si no se pudo resolver el Usuario correspondiente (ver
// GeneralAttendanceService.ListByDate) — igual se muestra la fila, con
// el id nomás, en vez de ocultar al participante.
type GeneralAttendanceView struct {
	ParticipanteID string `json:"participanteId"`
	Nombre         string `json:"nombre,omitempty"`
	Apellido       string `json:"apellido,omitempty"`
	Presente       *bool  `json:"presente"`
	Observacion    string `json:"observacion,omitempty"`
}
