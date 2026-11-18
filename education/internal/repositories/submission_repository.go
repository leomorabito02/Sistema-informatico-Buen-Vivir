package repositories

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"buenvivir/education/internal/events/producers"
	"buenvivir/education/internal/models"
)

const submissionsCollection = "entregas"

// ErrAlreadySubmitted señala que el participante ya entregó esa
// actividad: la entrega es única e inmutable.
var ErrAlreadySubmitted = errors.New("la actividad ya fue entregada")

// SubmissionDocID es el id determinístico que comparten la Entrega y la
// Asistencia de un participante en una actividad.
func SubmissionDocID(actividadID, participanteID string) string {
	return actividadID + "_" + participanteID
}

// SubmissionRepository encapsula el acceso a la colección Firestore
// "entregas".
type SubmissionRepository interface {
	// CreateWithAttendance crea la entrega y, en la misma transacción,
	// escribe la Asistencia del participante en esa actividad con
	// presente=true y fecha = activity.Fecha, y le asigna al participante
	// las habilidades de la actividad que todavía no tenía. O se guarda
	// todo o nada. Devuelve ErrAlreadySubmitted si ya había una entrega.
	CreateWithAttendance(ctx context.Context, activity *models.Activity, sub models.Submission) (*models.Submission, error)
	Get(ctx context.Context, id string) (*models.Submission, error)
	ListByActivity(ctx context.Context, actividadID string) ([]models.Submission, error)
	// GetForParticipant devuelve, indexadas por actividadId, las
	// entregas del participante en las actividades pedidas (las que no
	// entregó no aparecen en el mapa).
	GetForParticipant(ctx context.Context, participanteID string, actividadIDs []string) (map[string]*models.Submission, error)
}

type submissionRepository struct {
	client *firestore.Client
	events *producers.Educacion
}

// NewSubmissionRepository recibe opcionalmente el productor de eventos
// (asistencia.registrada, habilidad.otorgada y
// curso.completado_por_participante salen de cada entrega).
func NewSubmissionRepository(client *firestore.Client, events ...*producers.Educacion) SubmissionRepository {
	return &submissionRepository{client: client, events: firstProducer(events)}
}

func (r *submissionRepository) CreateWithAttendance(ctx context.Context, activity *models.Activity, sub models.Submission) (*models.Submission, error) {
	docID := SubmissionDocID(activity.ID, sub.ParticipanteID)
	sub.ActividadID = activity.ID
	if sub.Archivos == nil {
		sub.Archivos = []models.SubmissionFile{}
	}
	attendance := models.Attendance{
		ParticipanteID: sub.ParticipanteID,
		ActividadID:    activity.ID,
		Presente:       true,
		Fecha:          activity.Fecha,
	}

	err := r.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		// Habilidades (HU-21): entregar la actividad le da al participante
		// las habilidades que esa actividad desarrolla. Va en la misma
		// transacción que la entrega para que no pueda quedar una entrega
		// sin sus habilidades (o habilidades sin entrega). Firestore pide
		// leer todo antes de la primera escritura.
		newSkills, err := r.missingParticipantSkills(tx, activity.ID, sub.ParticipanteID)
		if err != nil {
			return err
		}
		completes, err := r.completesCourse(tx, activity, sub.ParticipanteID)
		if err != nil {
			return err
		}

		if err := tx.Create(r.client.Collection(submissionsCollection).Doc(docID), sub); err != nil {
			return err
		}
		if err := tx.Set(r.client.Collection(attendancesCollection).Doc(docID), attendance); err != nil {
			return err
		}
		for _, habilidadID := range newSkills {
			ps := models.ParticipantSkill{
				ParticipanteID: sub.ParticipanteID,
				HabilidadID:    habilidadID,
				Origen:         models.SkillOrigenActividad,
				Fecha:          sub.FechaEntrega,
				ActividadID:    activity.ID,
				CursoID:        activity.CursoID,
			}
			ref := r.client.Collection(participantSkillsCollection).Doc(ParticipantSkillDocID(sub.ParticipanteID, habilidadID))
			if err := tx.Create(ref, ps); err != nil {
				return err
			}
			if err := r.events.HabilidadOtorgada(ctx, tx, ps); err != nil {
				return err
			}
		}
		// Eventos (outbox): en la misma transacción que la entrega.
		if err := r.events.AsistenciaRegistrada(ctx, tx, docID, activity.CursoID, attendance); err != nil {
			return err
		}
		if completes {
			return r.events.CursoCompletado(ctx, tx, activity.CursoID, sub.ParticipanteID, sub.FechaEntrega)
		}
		return nil
	})
	if err != nil {
		if status.Code(err) == codes.AlreadyExists {
			return nil, ErrAlreadySubmitted
		}
		return nil, fmt.Errorf("error registrando entrega: %w", err)
	}
	sub.ID = docID
	r.events.Notify()
	return &sub, nil
}

// completesCourse dice si, con esta entrega, el participante completa
// todas las actividades del curso (todas las demás ya tienen entrega).
// Solo se calcula si hay que publicar eventos: son lecturas extra que
// la entrega en sí no necesita.
func (r *submissionRepository) completesCourse(tx *firestore.Transaction, activity *models.Activity, participanteID string) (bool, error) {
	if !r.events.Enabled() || activity.CursoID == "" {
		return false, nil
	}
	activities, err := tx.Documents(r.client.Collection(activitiesCollection).Where("curso_id", "==", activity.CursoID)).GetAll()
	if err != nil {
		return false, err
	}
	others := make([]*firestore.DocumentRef, 0, len(activities))
	for _, a := range activities {
		if a.Ref.ID != activity.ID {
			others = append(others, r.client.Collection(submissionsCollection).Doc(SubmissionDocID(a.Ref.ID, participanteID)))
		}
	}
	if len(others) == 0 {
		return true, nil
	}
	submitted, err := tx.GetAll(others)
	if err != nil {
		return false, err
	}
	for _, d := range submitted {
		if !d.Exists() {
			return false, nil
		}
	}
	return true, nil
}

// missingParticipantSkills devuelve las habilidades de la actividad que
// el participante todavía no tiene. Las que ya tiene (ganadas en otra
// actividad) no se tocan: se conserva la primera vez que las obtuvo.
func (r *submissionRepository) missingParticipantSkills(tx *firestore.Transaction, actividadID, participanteID string) ([]string, error) {
	links, err := tx.Documents(r.client.Collection(activitySkillsCollection).Where("actividad_id", "==", actividadID)).GetAll()
	if err != nil {
		return nil, err
	}
	if len(links) == 0 {
		return nil, nil
	}
	ids := make([]string, 0, len(links))
	refs := make([]*firestore.DocumentRef, 0, len(links))
	for _, l := range links {
		var link models.ActivitySkill
		if err := l.DataTo(&link); err != nil {
			return nil, err
		}
		ids = append(ids, link.HabilidadID)
		refs = append(refs, r.client.Collection(participantSkillsCollection).Doc(ParticipantSkillDocID(participanteID, link.HabilidadID)))
	}
	owned, err := tx.GetAll(refs)
	if err != nil {
		return nil, err
	}
	missing := make([]string, 0, len(ids))
	for i, doc := range owned {
		if !doc.Exists() {
			missing = append(missing, ids[i])
		}
	}
	return missing, nil
}

func (r *submissionRepository) Get(ctx context.Context, id string) (*models.Submission, error) {
	doc, err := r.client.Collection(submissionsCollection).Doc(id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("error obteniendo entrega: %w", err)
	}
	return submissionFromDoc(doc)
}

func (r *submissionRepository) ListByActivity(ctx context.Context, actividadID string) ([]models.Submission, error) {
	iter := r.client.Collection(submissionsCollection).Where("actividad_id", "==", actividadID).Documents(ctx)
	defer iter.Stop()

	submissions := make([]models.Submission, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error listando entregas: %w", err)
		}
		s, err := submissionFromDoc(doc)
		if err != nil {
			return nil, err
		}
		submissions = append(submissions, *s)
	}
	return submissions, nil
}

func (r *submissionRepository) GetForParticipant(ctx context.Context, participanteID string, actividadIDs []string) (map[string]*models.Submission, error) {
	result := make(map[string]*models.Submission, len(actividadIDs))
	if len(actividadIDs) == 0 {
		return result, nil
	}

	refs := make([]*firestore.DocumentRef, len(actividadIDs))
	for i, id := range actividadIDs {
		refs[i] = r.client.Collection(submissionsCollection).Doc(SubmissionDocID(id, participanteID))
	}
	docs, err := r.client.GetAll(ctx, refs)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo entregas: %w", err)
	}
	for _, doc := range docs {
		if !doc.Exists() {
			continue
		}
		s, err := submissionFromDoc(doc)
		if err != nil {
			return nil, err
		}
		result[s.ActividadID] = s
	}
	return result, nil
}

func submissionFromDoc(doc *firestore.DocumentSnapshot) (*models.Submission, error) {
	var s models.Submission
	if err := doc.DataTo(&s); err != nil {
		return nil, fmt.Errorf("error leyendo entrega: %w", err)
	}
	s.ID = doc.Ref.ID
	if s.Archivos == nil {
		s.Archivos = []models.SubmissionFile{}
	}
	return &s, nil
}
