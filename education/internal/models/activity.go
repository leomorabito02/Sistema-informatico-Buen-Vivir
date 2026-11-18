package models

import "time"

// Activity representa un documento de la colección Firestore
// "actividades". Los nombres de campo en Firestore van en snake_case
// (curso_id, titulo, consigna, fecha, created_at); ID no se persiste
// como campo, es el id del propio documento.
//
// Fecha es una fecha calendario (el día en que se dicta la actividad):
// se guarda como medianoche UTC de ese día y es también la fecha que
// toma la Asistencia que genera cada entrega (ver
// SubmissionRepository.CreateWithAttendance).
type Activity struct {
	ID        string    `json:"id" firestore:"-"`
	CursoID   string    `json:"cursoId" firestore:"curso_id"`
	Titulo    string    `json:"titulo" firestore:"titulo"`
	Consigna  string    `json:"consigna" firestore:"consigna"`
	Fecha     time.Time `json:"fecha" firestore:"fecha"`
	CreatedAt time.Time `json:"createdAt" firestore:"created_at"`
}

// ActivityInput es el payload para crear o editar una actividad.
// cursoId no forma parte del body: viaja en el path (POST
// /api/cursos/{id}/actividades o /api/docente/cursos/{cursoId}/actividades)
// y ya fue validado antes de llegar al service.
type ActivityInput struct {
	Titulo   string    `json:"titulo"`
	Consigna string    `json:"consigna"`
	Fecha    time.Time `json:"fecha"`
}
