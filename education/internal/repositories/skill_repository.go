package repositories

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"buenvivir/education/internal/models"
)

// Colecciones de habilidades (HU-21 / HU-58). Viven en la misma base
// Firestore que actividades y entregas: son datos de Educación, así que
// no hace falta pedirle nada a otro servicio para leerlas ni escribirlas.
const (
	skillsCollection            = "habilidades"
	activitySkillsCollection    = "actividad_habilidades"
	participantSkillsCollection = "participante_habilidades"
)

// ActivitySkillDocID arma el id del documento de Actividad_Habilidad.
// Clave compuesta del DER (actividad + habilidad) volcada al id: no puede
// haber dos documentos para el mismo par.
func ActivitySkillDocID(actividadID, habilidadID string) string {
	return actividadID + "_" + habilidadID
}

// ParticipantSkillDocID arma el id del documento de
// Participante_Habilidad (participante + habilidad), con el mismo
// criterio que ActivitySkillDocID.
func ParticipantSkillDocID(participanteID, habilidadID string) string {
	return participanteID + "_" + habilidadID
}

// SkillRepository encapsula el acceso a las tres colecciones de
// habilidades: el catálogo, las de cada actividad y las de cada
// participante.
type SkillRepository interface {
	// ListCatalog devuelve todo el catálogo de habilidades.
	ListCatalog(ctx context.Context) ([]models.Skill, error)
	// GetCatalog devuelve, indexadas por id, las habilidades pedidas que
	// existan en el catálogo (las que no existen no aparecen).
	GetCatalog(ctx context.Context, ids []string) (map[string]models.Skill, error)
	ListLinksByActivity(ctx context.Context, actividadID string) ([]models.ActivitySkill, error)
	ListLinksByCourse(ctx context.Context, cursoID string) ([]models.ActivitySkill, error)
	// SetActivitySkills reemplaza las habilidades de una actividad por
	// las recibidas. En una sola transacción: crea en el catálogo las que
	// todavía no existen (las que ya existen quedan como están, con su
	// nombre y categoría originales), agrega los vínculos nuevos y borra
	// los que ya no van. Devuelve las habilidades tal como quedaron en el
	// catálogo.
	SetActivitySkills(ctx context.Context, activity *models.Activity, skills []models.Skill) ([]models.Skill, error)
	// GrantToParticipants asigna las habilidades a cada participante. Si
	// un participante ya tenía una, no se toca (se conserva la primera
	// vez que la ganó). Es idempotente: se puede reintentar sin duplicar.
	GrantToParticipants(ctx context.Context, participanteIDs, habilidadIDs []string, grant models.SkillGrant) error
	ListByParticipant(ctx context.Context, participanteID string) ([]models.ParticipantSkill, error)
	// ListAllParticipantSkills devuelve todas las habilidades ganadas por
	// todos los participantes (para el Dashboard).
	ListAllParticipantSkills(ctx context.Context) ([]models.ParticipantSkill, error)
	// ListLinksBySkill devuelve las actividades que desarrollan la
	// habilidad.
	ListLinksBySkill(ctx context.Context, habilidadID string) ([]models.ActivitySkill, error)
}

type skillRepository struct {
	client *firestore.Client
}

func NewSkillRepository(client *firestore.Client) SkillRepository {
	return &skillRepository{client: client}
}

func (r *skillRepository) ListCatalog(ctx context.Context) ([]models.Skill, error) {
	docs, err := r.client.Collection(skillsCollection).Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("error listando habilidades: %w", err)
	}
	skills := make([]models.Skill, 0, len(docs))
	for _, d := range docs {
		s, err := skillFromDoc(d)
		if err != nil {
			return nil, err
		}
		skills = append(skills, s)
	}
	return skills, nil
}

func (r *skillRepository) GetCatalog(ctx context.Context, ids []string) (map[string]models.Skill, error) {
	result := make(map[string]models.Skill, len(ids))
	if len(ids) == 0 {
		return result, nil
	}
	refs := make([]*firestore.DocumentRef, len(ids))
	for i, id := range ids {
		refs[i] = r.client.Collection(skillsCollection).Doc(id)
	}
	docs, err := r.client.GetAll(ctx, refs)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo habilidades: %w", err)
	}
	for _, d := range docs {
		if !d.Exists() {
			continue
		}
		s, err := skillFromDoc(d)
		if err != nil {
			return nil, err
		}
		result[s.ID] = s
	}
	return result, nil
}

func (r *skillRepository) ListLinksByActivity(ctx context.Context, actividadID string) ([]models.ActivitySkill, error) {
	return r.listLinks(ctx, "actividad_id", actividadID)
}

func (r *skillRepository) ListLinksByCourse(ctx context.Context, cursoID string) ([]models.ActivitySkill, error) {
	return r.listLinks(ctx, "curso_id", cursoID)
}

func (r *skillRepository) ListLinksBySkill(ctx context.Context, habilidadID string) ([]models.ActivitySkill, error) {
	return r.listLinks(ctx, "habilidad_id", habilidadID)
}

func (r *skillRepository) listLinks(ctx context.Context, field, value string) ([]models.ActivitySkill, error) {
	docs, err := r.client.Collection(activitySkillsCollection).Where(field, "==", value).Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("error listando habilidades de actividades: %w", err)
	}
	links := make([]models.ActivitySkill, 0, len(docs))
	for _, d := range docs {
		var l models.ActivitySkill
		if err := d.DataTo(&l); err != nil {
			return nil, fmt.Errorf("error leyendo habilidad de actividad: %w", err)
		}
		links = append(links, l)
	}
	return links, nil
}

func (r *skillRepository) SetActivitySkills(ctx context.Context, activity *models.Activity, skills []models.Skill) ([]models.Skill, error) {
	catalogRefs := make([]*firestore.DocumentRef, len(skills))
	for i, s := range skills {
		catalogRefs[i] = r.client.Collection(skillsCollection).Doc(s.ID)
	}
	var saved []models.Skill

	err := r.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		saved = make([]models.Skill, 0, len(skills))

		// Firestore exige hacer todas las lecturas de la transacción
		// antes de la primera escritura.
		var catalogDocs []*firestore.DocumentSnapshot
		if len(catalogRefs) > 0 {
			var err error
			catalogDocs, err = tx.GetAll(catalogRefs)
			if err != nil {
				return err
			}
		}
		current, err := tx.Documents(r.client.Collection(activitySkillsCollection).Where("actividad_id", "==", activity.ID)).GetAll()
		if err != nil {
			return err
		}

		keep := make(map[string]bool, len(skills))
		for i, s := range skills {
			keep[s.ID] = true
			if catalogDocs[i].Exists() {
				existing, err := skillFromDoc(catalogDocs[i])
				if err != nil {
					return err
				}
				saved = append(saved, existing)
			} else {
				if err := tx.Create(catalogRefs[i], s); err != nil {
					return err
				}
				saved = append(saved, s)
			}
			link := models.ActivitySkill{ActividadID: activity.ID, HabilidadID: s.ID, CursoID: activity.CursoID}
			if err := tx.Set(r.client.Collection(activitySkillsCollection).Doc(ActivitySkillDocID(activity.ID, s.ID)), link); err != nil {
				return err
			}
		}
		for _, d := range current {
			var l models.ActivitySkill
			if err := d.DataTo(&l); err != nil {
				return err
			}
			if !keep[l.HabilidadID] {
				if err := tx.Delete(d.Ref); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error guardando habilidades de la actividad: %w", err)
	}
	return saved, nil
}

func (r *skillRepository) GrantToParticipants(ctx context.Context, participanteIDs, habilidadIDs []string, grant models.SkillGrant) error {
	// Sin transacción a propósito: una actividad con muchas entregas y
	// varias habilidades puede superar el límite de 500 escrituras por
	// transacción de Firestore. Cada documento se crea con Create (falla
	// si ya existe), así que reintentar después de una falla a mitad de
	// camino completa lo que faltó sin duplicar ni pisar nada.
	col := r.client.Collection(participantSkillsCollection)
	for _, p := range participanteIDs {
		for _, h := range habilidadIDs {
			ps := models.ParticipantSkill{
				ParticipanteID: p,
				HabilidadID:    h,
				Origen:         grant.Origen,
				Fecha:          grant.Fecha,
				ActividadID:    grant.ActividadID,
				CursoID:        grant.CursoID,
			}
			if _, err := col.Doc(ParticipantSkillDocID(p, h)).Create(ctx, ps); err != nil {
				if status.Code(err) == codes.AlreadyExists {
					continue
				}
				return fmt.Errorf("error asignando habilidad al participante: %w", err)
			}
		}
	}
	return nil
}

func (r *skillRepository) ListByParticipant(ctx context.Context, participanteID string) ([]models.ParticipantSkill, error) {
	docs, err := r.client.Collection(participantSkillsCollection).Where("participante_id", "==", participanteID).Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("error listando habilidades del participante: %w", err)
	}
	return participantSkillsFromDocs(docs)
}

func (r *skillRepository) ListAllParticipantSkills(ctx context.Context) ([]models.ParticipantSkill, error) {
	docs, err := r.client.Collection(participantSkillsCollection).Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("error listando habilidades ganadas: %w", err)
	}
	return participantSkillsFromDocs(docs)
}

func participantSkillsFromDocs(docs []*firestore.DocumentSnapshot) ([]models.ParticipantSkill, error) {
	result := make([]models.ParticipantSkill, 0, len(docs))
	for _, d := range docs {
		var ps models.ParticipantSkill
		if err := d.DataTo(&ps); err != nil {
			return nil, fmt.Errorf("error leyendo habilidad del participante: %w", err)
		}
		result = append(result, ps)
	}
	return result, nil
}

func skillFromDoc(doc *firestore.DocumentSnapshot) (models.Skill, error) {
	var s models.Skill
	if err := doc.DataTo(&s); err != nil {
		return s, fmt.Errorf("error leyendo habilidad: %w", err)
	}
	s.ID = doc.Ref.ID
	return s, nil
}
