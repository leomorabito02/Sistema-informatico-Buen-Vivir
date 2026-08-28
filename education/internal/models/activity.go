package models

import "time"

// Activity representa un documento de la colección Firestore
// "actividades". Los nombres de campo en Firestore van en snake_case
// (curso_id, titulo, fecha, created_at); ID no se persiste como campo,
// es el id del propio documento.
type Activity struct {
	ID        string    `json:"id" firestore:"-"`
	CursoID   string    `json:"cursoId" firestore:"curso_id"`
	Titulo    string    `json:"titulo" firestore:"titulo"`
	Fecha     time.Time `json:"fecha" firestore:"fecha"`
	CreatedAt time.Time `json:"createdAt" firestore:"created_at"`
}

// ActivityInput es el payload para crear una actividad. cursoId no
// forma parte del body: viaja en el path (POST
// /api/docente/cursos/{cursoId}/actividades) y ya fue validado como
// propio del docente autenticado antes de llegar al service.
type ActivityInput struct {
	Titulo string    `json:"titulo"`
	Fecha  time.Time `json:"fecha"`
}
