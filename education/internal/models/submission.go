package models

import "time"

// Submission representa un documento de la colección Firestore
// "entregas": lo que un participante envió para una actividad.
//
// El id del documento es determinístico ({actividadId}_{participanteId},
// el mismo que el de su Asistencia) y se crea con Create, que falla si
// ya existe: así hay una única entrega por participante y actividad, y
// una vez hecha no se puede reenviar ni anular.
//
// Los binarios de los adjuntos NO viven acá: están en el bucket de
// Cloud Storage de entregas (ver repositories.FileStorage). Firestore
// guarda solo sus metadatos, incluida la ruta del objeto en el bucket.
type Submission struct {
	ID             string           `json:"id" firestore:"-"`
	ParticipanteID string           `json:"participanteId" firestore:"participante_id"`
	ActividadID    string           `json:"actividadId" firestore:"actividad_id"`
	Texto          string           `json:"texto" firestore:"texto"`
	Archivos       []SubmissionFile `json:"archivos" firestore:"archivos"`
	FechaEntrega   time.Time        `json:"fechaEntrega" firestore:"fecha_entrega"`
}

// SubmissionFile son los metadatos de un adjunto de una entrega. Ruta
// es la ruta del objeto dentro del bucket y no se expone por JSON: el
// frontend descarga cada archivo por su índice a través de
// GET /api/entregas/{id}/archivos/{indice}, que vuelve a validar
// permisos antes de leer el bucket.
type SubmissionFile struct {
	Nombre string `json:"nombre" firestore:"nombre"`
	Tipo   string `json:"tipo" firestore:"tipo"`
	Tamano int64  `json:"tamano" firestore:"tamano"`
	Ruta   string `json:"-" firestore:"ruta"`
}

// SubmissionUpload es un adjunto ya leído del multipart, antes de
// subirlo al bucket.
type SubmissionUpload struct {
	Nombre string
	Tipo   string
	Data   []byte
}

const (
	SubmissionEstadoEntregado   = "entregado"
	SubmissionEstadoSinEntregar = "sin_entregar"
)

// SubmissionStatusRow es una fila de
// GET .../actividades/{actividadId}/entregas: un participante con
// inscripción activa en el curso de la actividad y su entrega, si la
// hizo. Nombre/Apellido pueden venir vacíos si auth-service no pudo
// resolver el usuario (el frontend cae al participanteId).
type SubmissionStatusRow struct {
	ParticipanteID string      `json:"participanteId"`
	Nombre         string      `json:"nombre,omitempty"`
	Apellido       string      `json:"apellido,omitempty"`
	Estado         string      `json:"estado"`
	Entrega        *Submission `json:"entrega,omitempty"`
}

// ParticipantActivity es una actividad tal como la ve el participante:
// la actividad más su propia entrega (nil si todavía no entregó).
type ParticipantActivity struct {
	Activity
	Entrega *Submission `json:"entrega,omitempty"`
}

// ParticipantCourseDetail es la respuesta de
// GET /api/participante/cursos/{id}: el curso, sus actividades con el
// estado de entrega del participante autenticado, y su progreso
// (actividades entregadas / total, HU-20).
type ParticipantCourseDetail struct {
	Curso       Course                `json:"curso"`
	Actividades []ParticipantActivity `json:"actividades"`
	Progreso    ParticipantProgress   `json:"progreso"`
}
