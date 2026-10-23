package models

import "time"

// CursoBaja es un documento de la colección Firestore "cursos_baja"
// (HU-16): la baja lógica de un curso. El curso sigue en Contentful
// (publicado, para que quien lo cursó lo vea en su historial), pero deja
// de aparecer en el catálogo y no acepta inscripciones nuevas. El id
// del documento es el id del curso, así que una baja no se repite.
type CursoBaja struct {
	CursoID           string    `json:"cursoId" firestore:"-"`
	Titulo            string    `json:"titulo" firestore:"titulo"`
	FechaBaja         time.Time `json:"fechaBaja" firestore:"fecha_baja"`
	UsuarioID         string    `json:"usuarioId" firestore:"usuario_id"`
	InscriptosActivos int       `json:"inscriptosActivos" firestore:"inscriptos_activos"`
}
