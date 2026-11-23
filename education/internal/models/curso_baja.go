package models

import "time"

// CursoBaja es la baja lógica de un curso (HU-16), en "cursos_baja". El curso sigue
// en Contentful para el historial pero sale del catálogo. Id = id del curso.
type CursoBaja struct {
	CursoID           string    `json:"cursoId" firestore:"-"`
	Titulo            string    `json:"titulo" firestore:"titulo"`
	FechaBaja         time.Time `json:"fechaBaja" firestore:"fecha_baja"`
	UsuarioID         string    `json:"usuarioId" firestore:"usuario_id"`
	InscriptosActivos int       `json:"inscriptosActivos" firestore:"inscriptos_activos"`
}
