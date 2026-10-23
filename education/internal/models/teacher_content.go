package models

import "time"

// Estados de moderación de un ContenidoDocente (DER: estado_moderacion).
const (
	ModeracionPendiente = "pendiente"
	ModeracionAprobado  = "aprobado"
	ModeracionRechazado = "rechazado"
)

// ContenidoDocente es un material que sube un docente externo (HU-23,
// DER ContenidoDocente). Queda pendiente hasta que la administración lo
// aprueba; aprobado, se publica en la biblioteca pública (HU-25). El
// archivo vive en el bucket privado (el mismo de las entregas), no en
// Contentful: lo pendiente no puede quedar en una CDN pública.
type ContenidoDocente struct {
	ID               string     `firestore:"-" json:"id"`
	DocenteID        string     `firestore:"docente_id" json:"docenteId"`
	Titulo           string     `firestore:"titulo" json:"titulo"`
	Descripcion      string     `firestore:"descripcion" json:"descripcion"`
	ArchivoNombre    string     `firestore:"archivo_nombre" json:"archivoNombre"`
	ArchivoTipo      string     `firestore:"archivo_tipo" json:"archivoTipo"`
	ArchivoTamano    int64      `firestore:"archivo_tamano" json:"archivoTamano"`
	ArchivoRuta      string     `firestore:"archivo_ruta" json:"-"`
	EstadoModeracion string     `firestore:"estado_moderacion" json:"estadoModeracion"`
	MotivoRechazo    string     `firestore:"motivo_rechazo,omitempty" json:"motivoRechazo,omitempty"`
	FechaSubida      time.Time  `firestore:"fecha_subida" json:"fechaSubida"`
	FechaModeracion  *time.Time `firestore:"fecha_moderacion,omitempty" json:"fechaModeracion,omitempty"`
	ModeradoPor      string     `firestore:"moderado_por,omitempty" json:"-"`
}

// ContenidoPublico es lo que ve cualquiera en la biblioteca: sin el
// docente, el estado ni la ruta del archivo.
type ContenidoPublico struct {
	ID            string    `json:"id"`
	Titulo        string    `json:"titulo"`
	Descripcion   string    `json:"descripcion"`
	ArchivoNombre string    `json:"archivoNombre"`
	ArchivoTipo   string    `json:"archivoTipo"`
	ArchivoTamano int64     `json:"archivoTamano"`
	FechaSubida   time.Time `json:"fechaSubida"`
}

// Publico recorta el contenido a lo que se muestra sin login.
func (c ContenidoDocente) Publico() ContenidoPublico {
	return ContenidoPublico{
		ID: c.ID, Titulo: c.Titulo, Descripcion: c.Descripcion,
		ArchivoNombre: c.ArchivoNombre, ArchivoTipo: c.ArchivoTipo,
		ArchivoTamano: c.ArchivoTamano, FechaSubida: c.FechaSubida,
	}
}

// ContenidoInput son los datos de texto al subir o editar un material.
type ContenidoInput struct {
	Titulo      string
	Descripcion string
}

// ArchivoSubido es el archivo de un material, ya leído del multipart.
type ArchivoSubido struct {
	Nombre string
	Tipo   string
	Data   []byte
}

// DescargaBiblioteca registra quién descarga un material de la
// biblioteca pública (HU-25, DER DescargaBiblioteca). Es un contacto
// para el seguimiento del centro, no un Usuario: no pide cuenta. Datos
// personales: solo los ve la administración.
type DescargaBiblioteca struct {
	ID          string    `firestore:"-" json:"id"`
	ContenidoID string    `firestore:"contenido_id" json:"contenidoId"`
	Nombre      string    `firestore:"nombre" json:"nombre"`
	Apellido    string    `firestore:"apellido" json:"apellido"`
	Email       string    `firestore:"email" json:"email"`
	Telefono    string    `firestore:"telefono" json:"telefono"`
	Fecha       time.Time `firestore:"fecha" json:"fecha"`
}

// DescargaInput es el body de POST /api/biblioteca/{id}/descargas.
type DescargaInput struct {
	Nombre   string `json:"nombre"`
	Apellido string `json:"apellido"`
	Email    string `json:"email"`
	Telefono string `json:"telefono"`
}
