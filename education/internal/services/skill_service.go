package services

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"buenvivir/education/internal/models"
	"buenvivir/education/internal/repositories"
)

const (
	// maxSkillsPerActivity limita cuántas habilidades puede tener una
	// actividad. Una actividad es una tarea concreta: más de 8
	// habilidades por una sola entrega inflaría el perfil del
	// participante y le restaría valor ante una empresa.
	maxSkillsPerActivity = 8
	// maxSkillNameLength evita nombres que en realidad son frases o
	// descripciones ("Aprender a trabajar con otros respetando...").
	maxSkillNameLength = 60
	// defaultCatalogSearchLimit es cuántos resultados devuelve el
	// autocompletado del catálogo.
	defaultCatalogSearchLimit = 20
)

// SkillService maneja el catálogo de habilidades, las de cada actividad y las de cada
// participante (HU-21 / HU-58). El permiso sobre el curso lo valida el controller.
type SkillService interface {
	// SearchCatalog devuelve las habilidades del catálogo cuyo nombre
	// contiene el texto buscado (sin importar mayúsculas ni tildes),
	// ordenadas alfabéticamente. Texto vacío = las primeras del catálogo.
	SearchCatalog(ctx context.Context, query string) ([]models.Skill, error)
	// GetCourseSkills devuelve las habilidades del curso (unión de las de
	// sus actividades) y las de cada actividad.
	GetCourseSkills(ctx context.Context, cursoID string) (*models.CourseSkills, error)
	// SetActivitySkills reemplaza las habilidades de la actividad y se las asigna también
	// a quienes ya la entregaron.
	SetActivitySkills(ctx context.Context, activity *models.Activity, inputs []models.SkillInput) ([]models.Skill, error)
	// ListParticipantSkills devuelve las habilidades que ganó el
	// participante, con nombre y categoría del catálogo.
	ListParticipantSkills(ctx context.Context, participanteID string) ([]models.ParticipantSkillView, error)
	// ListGrantedSkills devuelve todas las habilidades ganadas por todos
	// los participantes, con el curso en el que se ganaron (Dashboard).
	ListGrantedSkills(ctx context.Context) ([]models.GrantedSkill, error)
}

type skillService struct {
	repo           repositories.SkillRepository
	submissionRepo repositories.SubmissionRepository
	now            func() time.Time
}

func NewSkillService(repo repositories.SkillRepository, submissionRepo repositories.SubmissionRepository) SkillService {
	return &skillService{repo: repo, submissionRepo: submissionRepo, now: func() time.Time { return time.Now().UTC() }}
}

func (s *skillService) SearchCatalog(ctx context.Context, query string) ([]models.Skill, error) {
	catalog, err := s.repo.ListCatalog(ctx)
	if err != nil {
		return nil, err
	}
	sortSkills(catalog)

	key := models.SkillKey(query)
	result := make([]models.Skill, 0, defaultCatalogSearchLimit)
	for _, sk := range catalog {
		// sk.ID es la clave normalizada del nombre: comparar contra ella
		// hace que "equipo" encuentre "Trabajo en Equipo" y "diseno"
		// encuentre "Diseño".
		if key == "" || strings.Contains(sk.ID, key) {
			result = append(result, sk)
			if len(result) == defaultCatalogSearchLimit {
				break
			}
		}
	}
	return result, nil
}

func (s *skillService) GetCourseSkills(ctx context.Context, cursoID string) (*models.CourseSkills, error) {
	if strings.TrimSpace(cursoID) == "" {
		return nil, NewValidationError(errCourseIDRequired)
	}
	links, err := s.repo.ListLinksByCourse(ctx, cursoID)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(links))
	seen := make(map[string]bool, len(links))
	for _, l := range links {
		if !seen[l.HabilidadID] {
			seen[l.HabilidadID] = true
			ids = append(ids, l.HabilidadID)
		}
	}
	catalog, err := s.repo.GetCatalog(ctx, ids)
	if err != nil {
		return nil, err
	}

	result := &models.CourseSkills{
		Habilidades:  make([]models.Skill, 0, len(catalog)),
		PorActividad: make(map[string][]models.Skill),
	}
	for _, id := range ids {
		if sk, ok := catalog[id]; ok {
			result.Habilidades = append(result.Habilidades, sk)
		}
	}
	for _, l := range links {
		if sk, ok := catalog[l.HabilidadID]; ok {
			result.PorActividad[l.ActividadID] = append(result.PorActividad[l.ActividadID], sk)
		}
	}
	sortSkills(result.Habilidades)
	for _, list := range result.PorActividad {
		sortSkills(list)
	}
	return result, nil
}

func (s *skillService) SetActivitySkills(ctx context.Context, activity *models.Activity, inputs []models.SkillInput) ([]models.Skill, error) {
	skills, err := normalizeSkillInputs(inputs)
	if err != nil {
		return nil, err
	}

	saved, err := s.repo.SetActivitySkills(ctx, activity, skills)
	if err != nil {
		return nil, err
	}
	// Auditoría (pendiente, HU-35): registrar en LogAuditoria las
	// habilidades asignadas a la actividad y, más abajo, las asignadas a
	// participantes (ver docs/ia-habilidades.md, Pendientes).
	sortSkills(saved)

	if len(saved) == 0 {
		return saved, nil
	}
	if err := s.grantToPreviousSubmitters(ctx, activity, saved); err != nil {
		// Las habilidades de la actividad ya quedaron guardadas; lo que
		// falló es la asignación retroactiva. Volver a guardar completa lo
		// que faltó (GrantToParticipants es idempotente).
		return nil, fmt.Errorf("las habilidades se guardaron, pero no se pudieron asignar a quienes ya entregaron (volvé a guardar para reintentar): %w", err)
	}
	return saved, nil
}

func (s *skillService) grantToPreviousSubmitters(ctx context.Context, activity *models.Activity, skills []models.Skill) error {
	submissions, err := s.submissionRepo.ListByActivity(ctx, activity.ID)
	if err != nil {
		return err
	}
	if len(submissions) == 0 {
		return nil
	}
	participantes := make([]string, 0, len(submissions))
	for _, sub := range submissions {
		participantes = append(participantes, sub.ParticipanteID)
	}
	ids := make([]string, len(skills))
	for i, sk := range skills {
		ids[i] = sk.ID
	}
	return s.repo.GrantToParticipants(ctx, participantes, ids, models.SkillGrant{
		Origen:      models.SkillOrigenActividad,
		Fecha:       s.now(),
		ActividadID: activity.ID,
		CursoID:     activity.CursoID,
	})
}

func (s *skillService) ListGrantedSkills(ctx context.Context) ([]models.GrantedSkill, error) {
	owned, err := s.repo.ListAllParticipantSkills(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(owned))
	seen := make(map[string]bool, len(owned))
	for _, o := range owned {
		if !seen[o.HabilidadID] {
			seen[o.HabilidadID] = true
			ids = append(ids, o.HabilidadID)
		}
	}
	catalog, err := s.repo.GetCatalog(ctx, ids)
	if err != nil {
		return nil, err
	}

	// Las ganadas antes de guardar el curso se completan acá, sin
	// escribir: la actividad es la entrega más antigua del participante
	// entre las actividades que desarrollan esa habilidad.
	linksBySkill := make(map[string][]models.ActivitySkill)
	result := make([]models.GrantedSkill, 0, len(owned))
	for _, o := range owned {
		sk, ok := catalog[o.HabilidadID]
		if !ok {
			continue
		}
		g := models.GrantedSkill{
			ParticipanteID: o.ParticipanteID,
			HabilidadID:    o.HabilidadID,
			Nombre:         sk.Nombre,
			Categoria:      sk.Categoria,
			Origen:         o.Origen,
			Fecha:          o.Fecha,
			ActividadID:    o.ActividadID,
			CursoID:        o.CursoID,
		}
		if g.CursoID == "" {
			links, cached := linksBySkill[o.HabilidadID]
			if !cached {
				links, err = s.repo.ListLinksBySkill(ctx, o.HabilidadID)
				if err != nil {
					return nil, err
				}
				linksBySkill[o.HabilidadID] = links
			}
			if err := s.fillCourseFromSubmissions(ctx, &g, links); err != nil {
				return nil, err
			}
		}
		result = append(result, g)
	}
	sort.SliceStable(result, func(i, j int) bool { return result[i].Fecha.Before(result[j].Fecha) })
	return result, nil
}

func (s *skillService) fillCourseFromSubmissions(ctx context.Context, g *models.GrantedSkill, links []models.ActivitySkill) error {
	if len(links) == 0 {
		return nil
	}
	actividadIDs := make([]string, len(links))
	cursoPorActividad := make(map[string]string, len(links))
	for i, l := range links {
		actividadIDs[i] = l.ActividadID
		cursoPorActividad[l.ActividadID] = l.CursoID
	}
	subs, err := s.submissionRepo.GetForParticipant(ctx, g.ParticipanteID, actividadIDs)
	if err != nil {
		return err
	}
	var primera *models.Submission
	for _, sub := range subs {
		if primera == nil || sub.FechaEntrega.Before(primera.FechaEntrega) {
			primera = sub
		}
	}
	if primera != nil {
		g.ActividadID = primera.ActividadID
		g.CursoID = cursoPorActividad[primera.ActividadID]
	}
	return nil
}

func (s *skillService) ListParticipantSkills(ctx context.Context, participanteID string) ([]models.ParticipantSkillView, error) {
	if strings.TrimSpace(participanteID) == "" {
		return nil, NewValidationError("el id del participante es requerido")
	}
	owned, err := s.repo.ListByParticipant(ctx, participanteID)
	if err != nil {
		return nil, err
	}
	ids := make([]string, len(owned))
	for i, o := range owned {
		ids[i] = o.HabilidadID
	}
	catalog, err := s.repo.GetCatalog(ctx, ids)
	if err != nil {
		return nil, err
	}

	result := make([]models.ParticipantSkillView, 0, len(owned))
	for _, o := range owned {
		sk, ok := catalog[o.HabilidadID]
		if !ok {
			// Habilidad borrada del catálogo a mano: no hay nombre que
			// mostrar, se omite en vez de devolver una fila vacía.
			continue
		}
		result = append(result, models.ParticipantSkillView{Skill: sk, Origen: o.Origen, Fecha: o.Fecha})
	}
	sort.SliceStable(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result, nil
}

// normalizeSkillInputs valida la lista que manda el frontend y la lleva
// al formato del catálogo: nombre prolijo, id = clave normalizada y sin
// repetidas (si llegan "Huerta" y "huerta", queda una sola).
func normalizeSkillInputs(inputs []models.SkillInput) ([]models.Skill, error) {
	skills := make([]models.Skill, 0, len(inputs))
	seen := make(map[string]bool, len(inputs))
	for _, in := range inputs {
		nombre := models.CleanSkillName(in.Nombre)
		key := models.SkillKey(nombre)
		if key == "" {
			return nil, NewValidationError("cada habilidad necesita un nombre")
		}
		if utf8.RuneCountInString(nombre) > maxSkillNameLength {
			return nil, NewValidationError(fmt.Sprintf("el nombre de la habilidad %q es demasiado largo (máximo %d caracteres)", nombre, maxSkillNameLength))
		}
		if !in.Categoria.IsValid() {
			return nil, NewValidationError(fmt.Sprintf("la categoría de %q no es válida: tiene que ser %s", nombre, categoryList()))
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		skills = append(skills, models.Skill{ID: key, Nombre: nombre, Categoria: in.Categoria})
	}
	if len(skills) > maxSkillsPerActivity {
		return nil, NewValidationError(fmt.Sprintf("una actividad puede tener como máximo %d habilidades", maxSkillsPerActivity))
	}
	return skills, nil
}

func categoryList() string {
	names := make([]string, len(models.SkillCategories))
	for i, c := range models.SkillCategories {
		names[i] = string(c)
	}
	return strings.Join(names, ", ")
}

func sortSkills(skills []models.Skill) {
	sort.SliceStable(skills, func(i, j int) bool { return skills[i].ID < skills[j].ID })
}
