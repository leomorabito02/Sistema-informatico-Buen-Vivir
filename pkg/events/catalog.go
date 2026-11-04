package events

import "time"

// Topics. Los nombres reales se leen de variables de entorno en cada
// servicio (ver la config de cada uno); estos son los valores por
// defecto y los que crea infra/pubsub/setup.sh. Cada topic tiene su
// dead-letter topic: <topic>-dlq.
const (
	TopicUsuarios         = "usuarios-eventos"
	TopicEducacion        = "educacion-eventos"
	TopicPromocionHumana  = "promocion-humana-eventos"
	TopicContenido        = "contenido-eventos"
	TopicInsercionLaboral = "insercion-laboral-eventos"
	TopicCampanas         = "campanas-eventos"

	DLQSuffix = "-dlq"
)

// Productores (campo "producer" del envelope).
const (
	ProducerAutenticacion    = "servicio-autenticacion"
	ProducerEducacion        = "servicio-educacion"
	ProducerPromocionHumana  = "servicio-promocion-humana"
	ProducerContentful       = "receptor-contentful"
	ProducerInsercionLaboral = "servicio-insercion-laboral"
	ProducerCampanas         = "servicio-campanas"
)

// Tipos de evento: "<dominio>.<entidad>.<acción>".
const (
	// Autenticación.
	UsuarioCreado         = "usuarios.usuario.creado"
	EmailSolicitado       = "usuarios.email.solicitado"
	ParticipanteCreado    = "usuarios.participante.creado"
	ParticipanteEliminado = "usuarios.participante.eliminado"
	UsuarioActualizado    = "usuarios.usuario.actualizado"

	// Inserción Laboral.
	OfertaCreada            = "insercion_laboral.oferta.creada"
	OfertaRevisada          = "insercion_laboral.oferta.revisada"
	PostulacionCreada       = "insercion_laboral.postulacion.creada"
	PostulacionSeleccionada = "insercion_laboral.postulacion.seleccionada"
	PostulacionDecidida     = "insercion_laboral.postulacion.decidida"

	// Campañas.
	CampanaCreada = "campanas.campana.creada"

	// Educación.
	CursoCreado                  = "educacion.curso.creado"
	CursoActualizado             = "educacion.curso.actualizado"
	CursoEliminado               = "educacion.curso.eliminado"
	InscripcionCreada            = "educacion.inscripcion.creada"
	AsistenciaRegistrada         = "educacion.asistencia.registrada"
	CursoCompletadoParticipante  = "educacion.curso.completado_por_participante"
	HabilidadOtorgada            = "educacion.habilidad.otorgada"
	AsistenciaGeneralRegistrada  = "educacion.asistencia_general.registrada"
	AsistenciaGeneralActualizada = "educacion.asistencia_general.actualizada"
	ContenidoDocenteSubido       = "educacion.contenido_docente.subido"
	ContenidoDocenteModerado     = "educacion.contenido_docente.moderado"
	ActividadRecordatorio        = "educacion.actividad.recordatorio"

	// Promoción Humana.
	FichaCreada      = "promocion_humana.ficha.creada"
	FichaActualizada = "promocion_humana.ficha.actualizada"
	HitoRegistrado   = "promocion_humana.hito.registrado"
	HitoActualizado  = "promocion_humana.hito.actualizado"

	// Contenido (webhooks de Contentful).
	ContenidoPublicado    = "contenido.publicado"
	ContenidoDespublicado = "contenido.despublicado"
)

// currentVersions es la versión vigente de cada evento. Un cambio
// incompatible del payload implica subir la versión acá (y documentarlo
// en docs/eventos/asyncapi.yaml); los consumidores ignoran las versiones
// que no conocen.
var currentVersions = map[string]int{}

// CurrentVersion devuelve la versión vigente de eventType (1 si no
// figura en currentVersions).
func CurrentVersion(eventType string) int {
	if v, ok := currentVersions[eventType]; ok {
		return v
	}
	return 1
}

// ── Payloads ──────────────────────────────────────────────────────────
//
// Regla (Ley 25.326, RNF-06, RNF-08): SOLO ids y metadatos. Nunca
// nombres, DNI, teléfono, email, texto libre ni contenido de fichas.

// UsuarioCreadoPayload es el payload de usuarios.usuario.creado.
type UsuarioCreadoPayload struct {
	UserID string `json:"user_id"`
	RoleID string `json:"role_id"`
}

// EmailSolicitadoPayload es el payload de usuarios.email.solicitado.
// Tipo: "verificacion", "cuenta_aprobada", "recuperacion_password".
type EmailSolicitadoPayload struct {
	RequestID string `json:"request_id"`
	UserID    string `json:"user_id"`
	Tipo      string `json:"tipo"`
}

// OfertaCreadaPayload es el payload de insercion_laboral.oferta.creada.
type OfertaCreadaPayload struct {
	OfertaID  string `json:"oferta_id"`
	EmpresaID string `json:"empresa_id"`
	Estado    string `json:"estado"`
}

// OfertaRevisadaPayload es el payload de insercion_laboral.oferta.revisada.
type OfertaRevisadaPayload struct {
	OfertaID  string `json:"oferta_id"`
	EmpresaID string `json:"empresa_id"`
	Aprobada  bool   `json:"aprobada"`
}

// PostulacionCreadaPayload es el payload de insercion_laboral.postulacion.creada.
type PostulacionCreadaPayload struct {
	PostulacionID  string `json:"postulacion_id"`
	OfertaID       string `json:"oferta_id"`
	ParticipanteID string `json:"participante_id"`
	EmpresaID      string `json:"empresa_id"`
}

// PostulacionDecididaPayload es el payload de insercion_laboral.postulacion.decidida.
type PostulacionDecididaPayload struct {
	PostulacionID  string `json:"postulacion_id"`
	OfertaID       string `json:"oferta_id"`
	ParticipanteID string `json:"participante_id"`
	EmpresaID      string `json:"empresa_id"`
	Resultado      string `json:"resultado"`
}

// CampanaCreadaPayload es el payload de campanas.campana.creada.
type CampanaCreadaPayload struct {
	CampanaID string `json:"campana_id"`
	CreadoPor string `json:"creado_por"`
}

// ParticipantePayload es el payload de usuarios.participante.creado y
// usuarios.participante.eliminado (propuesto para Autenticación).
// ParticipanteID es el id que usan Educación y Promoción Humana para
// referirse al participante; UserID es el id del usuario asociado.
type ParticipantePayload struct {
	ParticipanteID string `json:"participante_id"`
	UserID         string `json:"user_id"`
}

// UsuarioActualizadoPayload es el payload de usuarios.usuario.actualizado
// (propuesto). Solo el estado de la cuenta, sin datos personales.
type UsuarioActualizadoPayload struct {
	UserID         string `json:"user_id"`
	RoleID         string `json:"role_id"`
	ParticipanteID string `json:"participante_id,omitempty"`
	Status         string `json:"status"`
	ApprovalStatus string `json:"approval_status,omitempty"`
}

// PostulacionSeleccionadaPayload (propuesto para Inserción Laboral).
type PostulacionSeleccionadaPayload struct {
	PostulacionID  string `json:"postulacion_id"`
	OfertaID       string `json:"oferta_id"`
	ParticipanteID string `json:"participante_id"`
	EmpresaID      string `json:"empresa_id"`
}

// CursoPayload es el payload de educacion.curso.creado/actualizado.
type CursoPayload struct {
	CursoID         string `json:"curso_id"`
	AreaFormativaID string `json:"area_formativa_id"`
	DocenteID       string `json:"docente_id"`
	Modalidad       string `json:"modalidad"`
}

// CursoEliminadoPayload es el payload de educacion.curso.eliminado.
type CursoEliminadoPayload struct {
	CursoID string `json:"curso_id"`
}

// InscripcionCreadaPayload: lo consumen Notificaciones (email de
// confirmación, HU-26/HU-60) y Promoción Humana (línea de tiempo).
type InscripcionCreadaPayload struct {
	InscripcionID    string    `json:"inscripcion_id"`
	CursoID          string    `json:"curso_id"`
	ParticipanteID   string    `json:"participante_id"`
	DocenteID        string    `json:"docente_id,omitempty"`
	Titulo           string    `json:"titulo,omitempty"`
	FechaInscripcion time.Time `json:"fecha_inscripcion"`
}

// ActividadRecordatorioPayload: aviso programado de actividad próxima (HU-26).
type ActividadRecordatorioPayload struct {
	ActividadID    string `json:"actividad_id"`
	CursoID        string `json:"curso_id"`
	ParticipanteID string `json:"participante_id"`
	Fecha          string `json:"fecha"`
	Titulo         string `json:"titulo,omitempty"`
}

// AsistenciaRegistradaPayload: asistencia a una actividad, generada al
// entregarla (HU-19/HU-20).
type AsistenciaRegistradaPayload struct {
	AsistenciaID   string    `json:"asistencia_id"`
	ActividadID    string    `json:"actividad_id"`
	CursoID        string    `json:"curso_id"`
	ParticipanteID string    `json:"participante_id"`
	Fecha          time.Time `json:"fecha"`
	Presente       bool      `json:"presente"`
}

// CursoCompletadoPayload: el participante entregó el 100% de las
// actividades del curso.
type CursoCompletadoPayload struct {
	CursoID        string    `json:"curso_id"`
	ParticipanteID string    `json:"participante_id"`
	CompletadoEn   time.Time `json:"completado_en"`
}

// HabilidadOtorgadaPayload: el participante ganó una habilidad al
// entregar una actividad que la desarrolla (HU-21/HU-58).
type HabilidadOtorgadaPayload struct {
	HabilidadID    string    `json:"habilidad_id"`
	ParticipanteID string    `json:"participante_id"`
	ActividadID    string    `json:"actividad_id"`
	CursoID        string    `json:"curso_id"`
	Origen         string    `json:"origen"`
	Fecha          time.Time `json:"fecha"`
}

// AsistenciaGeneralPayload: asistencia diaria al centro (HU-19bis). La
// observación es texto libre y NO viaja.
type AsistenciaGeneralPayload struct {
	AsistenciaID   string `json:"asistencia_id"`
	ParticipanteID string `json:"participante_id"`
	Fecha          string `json:"fecha"` // AAAA-MM-DD
	Presente       bool   `json:"presente"`
}

// Tipos de contenido docente.
const (
	ContenidoTipoCurso    = "curso"    // curso en borrador de un docente interno
	ContenidoTipoMaterial = "material" // ContenidoDocente de un docente externo (DER)
)

// ContenidoDocenteSubidoPayload: un docente cargó contenido que espera
// la revisión de la administración (HU-23). Con tipo "curso" viene
// curso_id; con tipo "material", contenido_id (agregado 2026-10-22, campo
// opcional: compatible con la versión 1).
type ContenidoDocenteSubidoPayload struct {
	CursoID     string `json:"curso_id,omitempty"`
	ContenidoID string `json:"contenido_id,omitempty"`
	DocenteID   string `json:"docente_id"`
	Tipo        string `json:"tipo"`
}

// Resultados de la moderación de contenido docente.
const (
	ModeracionAprobado  = "aprobado"
	ModeracionRechazado = "rechazado"
)

// ContenidoDocenteModeradoPayload: la administración revisó un curso o
// un material. contenido_id, docente_id y tipo se agregaron el 2026-10-22
// (opcionales: compatible con la versión 1).
type ContenidoDocenteModeradoPayload struct {
	CursoID     string `json:"curso_id,omitempty"`
	ContenidoID string `json:"contenido_id,omitempty"`
	DocenteID   string `json:"docente_id,omitempty"`
	Tipo        string `json:"tipo,omitempty"`
	Resultado   string `json:"resultado"`
}

// FichaPayload es el payload de promocion_humana.ficha.creada/actualizada.
// Nunca lleva el contenido de la ficha.
type FichaPayload struct {
	FichaID        string `json:"ficha_id"`
	ParticipanteID string `json:"participante_id"`
	CreadoPor      string `json:"creado_por,omitempty"`
}

// HitoPayload es el payload de promocion_humana.hito.registrado/actualizado.
// Nunca lleva la descripción del hito.
type HitoPayload struct {
	HitoID              string `json:"hito_id"`
	FichaID             string `json:"ficha_id"`
	ParticipanteID      string `json:"participante_id"`
	Fecha               string `json:"fecha"` // AAAA-MM-DD
	ResponsableID       string `json:"responsable_id,omitempty"`
	VisibleParticipante bool   `json:"visible_participante"`
}

// ContenidoPayload es el payload de contenido.publicado/despublicado.
type ContenidoPayload struct {
	EntryID     string `json:"entry_id"`
	ContentType string `json:"content_type"`
	Environment string `json:"environment"`
	// Accion es el topic de Contentful que lo originó: "publish",
	// "unpublish" o "delete".
	Accion   string `json:"accion"`
	Revision int    `json:"revision,omitempty"`
}
