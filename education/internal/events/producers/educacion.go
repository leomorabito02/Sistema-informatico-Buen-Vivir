// Package producers arma los eventos que publica Educación a partir de
// sus modelos de dominio y los deja en el outbox. Concentra acá qué
// datos salen en cada payload: SOLO ids y metadatos (nunca nombres,
// textos de entregas, observaciones ni nada personal).
package producers

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"

	"buenvivir/education/internal/events"
	"buenvivir/education/internal/models"
	"buenvivir/education/internal/outbox"
)

// Educacion publica los eventos del servicio de Educación. Un
// *Educacion nil no publica nada (mensajería desactivada), así que los
// repositorios y servicios lo pueden usar sin chequear.
type Educacion struct {
	writer         *outbox.Writer
	topicEducacion string
	topicContenido string
}

func NewEducacion(writer *outbox.Writer, topicEducacion, topicContenido string) *Educacion {
	return &Educacion{writer: writer, topicEducacion: topicEducacion, topicContenido: topicContenido}
}

// Enabled dice si hay que publicar. Los repositorios lo usan para no
// hacer lecturas extra (por ejemplo, el chequeo de curso completado)
// cuando la mensajería está apagada.
func (p *Educacion) Enabled() bool { return p != nil && p.writer != nil }

// Notify avisa al publicador después del commit de una transacción.
func (p *Educacion) Notify() {
	if p.Enabled() {
		p.writer.Notify()
	}
}

// emit escribe en la transacción si hay una, o directo si tx es nil.
func (p *Educacion) emit(ctx context.Context, tx *firestore.Transaction, evs ...outbox.Event) error {
	if !p.Enabled() {
		return nil
	}
	if tx != nil {
		return p.writer.AddTx(ctx, tx, evs...)
	}
	return p.writer.Add(ctx, evs...)
}

func (p *Educacion) event(eventType string, payload any, orderingKey string) outbox.Event {
	if p == nil {
		return outbox.Event{} // emit no lo usa: la mensajería está apagada
	}
	return outbox.Event{Topic: p.topicEducacion, Type: eventType, Payload: payload, OrderingKey: orderingKey}
}

// ── Cursos (viven en Contentful: se publican después de confirmado el
// cambio, sin transacción compartida) ─────────────────────────────────

func cursoPayload(c *models.Course) events.CursoPayload {
	p := events.CursoPayload{CursoID: c.ID, DocenteID: c.ResponsibleTeacher, Modalidad: c.Mode}
	if c.TrainingArea != nil {
		p.AreaFormativaID = c.TrainingArea.ID
	}
	return p
}

func (p *Educacion) CursoCreado(ctx context.Context, c *models.Course) error {
	return p.emit(ctx, nil, p.event(events.CursoCreado, cursoPayload(c), ""))
}

func (p *Educacion) CursoActualizado(ctx context.Context, c *models.Course) error {
	return p.emit(ctx, nil, p.event(events.CursoActualizado, cursoPayload(c), ""))
}

// CursoEliminado: en Contentful el borrado es físico; el historial que
// depende del curso (inscripciones, entregas, asistencias) queda en
// Firestore.
func (p *Educacion) CursoEliminado(ctx context.Context, cursoID string) error {
	return p.emit(ctx, nil, p.event(events.CursoEliminado, events.CursoEliminadoPayload{CursoID: cursoID}, ""))
}

// ContenidoDocenteSubido: un docente cargó un curso, que queda en
// borrador hasta que la administración lo publique (HU-23).
func (p *Educacion) ContenidoDocenteSubido(ctx context.Context, c *models.Course) error {
	return p.emit(ctx, nil, p.event(events.ContenidoDocenteSubido, events.ContenidoDocenteSubidoPayload{
		CursoID: c.ID, DocenteID: c.ResponsibleTeacher, Tipo: events.ContenidoTipoCurso,
	}, ""))
}

// ContenidoDocenteModerado: la administración publicó (aprobó) un curso.
// docenteID es opcional para mantener compatibilidad con tests/llamadores previos;
// cuando está presente permite que Notificaciones avise al docente responsable.
func (p *Educacion) ContenidoDocenteModerado(ctx context.Context, cursoID, resultado string, docenteID ...string) error {
	docente := ""
	if len(docenteID) > 0 {
		docente = docenteID[0]
	}
	return p.emit(ctx, nil, p.event(events.ContenidoDocenteModerado, events.ContenidoDocenteModeradoPayload{
		CursoID:   cursoID,
		DocenteID: docente,
		Tipo:      events.ContenidoTipoCurso,
		Resultado: resultado,
	}, ""))
}

// ── Contenido de docentes externos (ContenidoDocente, Firestore: van en
// la misma transacción que el cambio) ──────────────────────────────────

// MaterialDocenteSubido: un docente subió un material que queda
// pendiente de moderación (HU-23).
func (p *Educacion) MaterialDocenteSubido(ctx context.Context, tx *firestore.Transaction, c models.ContenidoDocente) error {
	return p.emit(ctx, tx, p.event(events.ContenidoDocenteSubido, materialSubidoPayload(c), ""))
}

// Solo ids y el estado: ni el título ni la descripción ni el motivo del
// rechazo (sin texto libre en los eventos, ADR 0007).
func materialSubidoPayload(c models.ContenidoDocente) events.ContenidoDocenteSubidoPayload {
	return events.ContenidoDocenteSubidoPayload{ContenidoID: c.ID, DocenteID: c.DocenteID, Tipo: events.ContenidoTipoMaterial}
}

func materialModeradoPayload(c models.ContenidoDocente) events.ContenidoDocenteModeradoPayload {
	return events.ContenidoDocenteModeradoPayload{
		ContenidoID: c.ID, DocenteID: c.DocenteID, Tipo: events.ContenidoTipoMaterial, Resultado: c.EstadoModeracion,
	}
}

// MaterialDocenteModerado: la administración aprobó o rechazó un
// material. Notificaciones le avisa al docente (HU-23).
func (p *Educacion) MaterialDocenteModerado(ctx context.Context, tx *firestore.Transaction, c models.ContenidoDocente) error {
	return p.emit(ctx, tx, p.event(events.ContenidoDocenteModerado, materialModeradoPayload(c), ""))
}

// ── Inscripciones, entregas, asistencia y habilidades (Firestore: van
// en la misma transacción que el cambio) ──────────────────────────────

func (p *Educacion) InscripcionCreada(ctx context.Context, tx *firestore.Transaction, e models.Enrollment) error {
	return p.emit(ctx, tx, p.event(events.InscripcionCreada, events.InscripcionCreadaPayload{
		InscripcionID:    e.ID,
		CursoID:          e.CursoID,
		ParticipanteID:   e.ParticipanteID,
		DocenteID:        e.TeacherID,
		Titulo:           e.CourseTitle,
		FechaInscripcion: e.FechaInscripcion,
	}, e.ParticipanteID))
}

// ActividadRecordatorio emite el recordatorio programado de una actividad próxima (HU-26).
func (p *Educacion) ActividadRecordatorio(ctx context.Context, tx *firestore.Transaction, payload events.ActividadRecordatorioPayload) error {
	return p.emit(ctx, tx, p.event(events.ActividadRecordatorio, payload, payload.ParticipanteID))
}

// AsistenciaRegistrada sale de cada entrega de actividad.
func (p *Educacion) AsistenciaRegistrada(ctx context.Context, tx *firestore.Transaction, asistenciaID, cursoID string, a models.Attendance) error {
	return p.emit(ctx, tx, p.event(events.AsistenciaRegistrada, events.AsistenciaRegistradaPayload{
		AsistenciaID:   asistenciaID,
		ActividadID:    a.ActividadID,
		CursoID:        cursoID,
		ParticipanteID: a.ParticipanteID,
		Fecha:          a.Fecha,
		Presente:       a.Presente,
	}, a.ParticipanteID))
}

// CursoCompletado: el participante entregó todas las actividades del
// curso. No dispara IA (las habilidades se ganan por actividad, al
// entregar; ver docs/ia-habilidades.md).
func (p *Educacion) CursoCompletado(ctx context.Context, tx *firestore.Transaction, cursoID, participanteID string, at time.Time) error {
	return p.emit(ctx, tx, p.event(events.CursoCompletadoParticipante, events.CursoCompletadoPayload{
		CursoID: cursoID, ParticipanteID: participanteID, CompletadoEn: at,
	}, participanteID))
}

// HabilidadOtorgada: el participante ganó una habilidad (al entregar, o
// retroactivamente al agregarle habilidades a una actividad ya entregada).
// nombre es el nombre de la habilidad para el aviso al participante
// (vacío si no se pudo leer del catálogo).
func (p *Educacion) HabilidadOtorgada(ctx context.Context, tx *firestore.Transaction, s models.ParticipantSkill, nombre string) error {
	return p.emit(ctx, tx, p.event(events.HabilidadOtorgada, events.HabilidadOtorgadaPayload{
		HabilidadID:     s.HabilidadID,
		ParticipanteID:  s.ParticipanteID,
		ActividadID:     s.ActividadID,
		CursoID:         s.CursoID,
		Origen:          s.Origen,
		Fecha:           s.Fecha,
		HabilidadNombre: nombre,
	}, s.ParticipanteID))
}

// AsistenciaGeneral: registrada la primera vez que se carga al
// participante en esa fecha, actualizada si se edita. La observación
// (texto libre) no viaja.
func (p *Educacion) AsistenciaGeneral(ctx context.Context, tx *firestore.Transaction, asistenciaID string, a models.GeneralAttendance, nueva bool) error {
	eventType := events.AsistenciaGeneralActualizada
	if nueva {
		eventType = events.AsistenciaGeneralRegistrada
	}
	return p.emit(ctx, tx, p.event(eventType, events.AsistenciaGeneralPayload{
		AsistenciaID:   asistenciaID,
		ParticipanteID: a.ParticipanteID,
		Fecha:          a.Fecha,
		Presente:       a.Presente,
	}, a.ParticipanteID))
}

// ── Contenido (webhooks de Contentful) ───────────────────────────────

// Contenido deja en el outbox un evento de contenido-eventos. eventID es
// determinístico (ver ContentfulWebhook) para que los reintentos de
// Contentful no generen eventos distintos.
func (p *Educacion) Contenido(ctx context.Context, eventID, eventType string, payload events.ContenidoPayload) error {
	if !p.Enabled() {
		return nil
	}
	return p.writer.Add(ctx, outbox.Event{Topic: p.topicContenido, Type: eventType, Payload: payload, ID: eventID})
}
