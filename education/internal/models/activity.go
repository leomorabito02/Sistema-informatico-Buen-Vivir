package models

import "time"

// Activity es un documento de la colección "actividades". Fecha es el día en que se
// dicta (medianoche UTC) y es la fecha de la asistencia que genera cada entrega.
type Activity struct {
	ID        string    `json:"id" firestore:"-"`
	CursoID   string    `json:"cursoId" firestore:"curso_id"`
	Titulo    string    `json:"titulo" firestore:"titulo"`
	Consigna  string    `json:"consigna" firestore:"consigna"`
	Fecha     time.Time `json:"fecha" firestore:"fecha"`
	CreatedAt time.Time `json:"createdAt" firestore:"created_at"`
}

// ActivityInput es el payload para crear o editar una actividad (el cursoId viaja en el path).
type ActivityInput struct {
	Titulo   string    `json:"titulo"`
	Consigna string    `json:"consigna"`
	Fecha    time.Time `json:"fecha"`
}
