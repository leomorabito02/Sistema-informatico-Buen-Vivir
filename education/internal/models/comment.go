package models

import "time"

// Tipos de contenido que se pueden comentar (HU-24). Es la referencia
// polimórfica del DER: entidad_tipo dice de qué tabla es entidad_id.
const (
	ComentarioCurso   = "curso"
	ComentarioCampana = "campana"
)

// Comentario es una opinión sobre un curso o una campaña (HU-24, DER
// Comentario). Admite autores sin cuenta: el contenido público se ve sin
// login, así que se guardan nombre y apellido en el propio comentario.
// UsuarioID queda vacío en esos casos y nunca se devuelve en el JSON.
type Comentario struct {
	ID            string    `firestore:"-" json:"id"`
	EntidadTipo   string    `firestore:"entidad_tipo" json:"entidadTipo"`
	EntidadID     string    `firestore:"entidad_id" json:"entidadId"`
	UsuarioID     string    `firestore:"usuario_id,omitempty" json:"-"`
	NombreAutor   string    `firestore:"nombre_autor" json:"nombre"`
	ApellidoAutor string    `firestore:"apellido_autor" json:"apellido"`
	Texto         string    `firestore:"texto" json:"texto"`
	Fecha         time.Time `firestore:"fecha" json:"fecha"`
}

// ComentarioInput es el body de POST /api/comentarios.
type ComentarioInput struct {
	EntidadTipo string `json:"entidadTipo"`
	EntidadID   string `json:"entidadId"`
	Nombre      string `json:"nombre"`
	Apellido    string `json:"apellido"`
	Texto       string `json:"texto"`
}
