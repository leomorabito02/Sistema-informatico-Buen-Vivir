package repositories

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"buenvivir/education/internal/models"
)

func TestSkillRepository_SetActivitySkills_CreatesCatalogAndLinks(t *testing.T) {
	client := newTestFirestore(t)
	for _, c := range []string{skillsCollection, activitySkillsCollection, participantSkillsCollection} {
		wipeCollection(t, client, c)
	}
	repo := NewSkillRepository(client)
	ctx := t.Context()
	activity := &models.Activity{ID: "a1", CursoID: "c1"}

	// Una habilidad ya existe en el catálogo con otro nombre/categoría:
	// se respeta la existente, no se pisa.
	_, err := client.Collection(skillsCollection).Doc("trabajo-en-equipo").Set(ctx, models.Skill{Nombre: "Trabajo en equipo", Categoria: models.SkillCategorySocioemoc})
	require.NoError(t, err)

	saved, err := repo.SetActivitySkills(ctx, activity, []models.Skill{
		{ID: "trabajo-en-equipo", Nombre: "Trabajo En Equipo", Categoria: models.SkillCategoryOficio},
		{ID: "armado-de-huerta", Nombre: "Armado de huerta", Categoria: models.SkillCategoryNaturaleza},
	})
	require.NoError(t, err)
	require.Len(t, saved, 2)
	assert.Equal(t, "Trabajo en equipo", saved[0].Nombre, "conserva el nombre del catálogo")
	assert.Equal(t, models.SkillCategorySocioemoc, saved[0].Categoria)

	catalog, err := repo.ListCatalog(ctx)
	require.NoError(t, err)
	assert.Len(t, catalog, 2)

	links, err := repo.ListLinksByActivity(ctx, "a1")
	require.NoError(t, err)
	assert.Len(t, links, 2)
	byCourse, err := repo.ListLinksByCourse(ctx, "c1")
	require.NoError(t, err)
	assert.Len(t, byCourse, 2)

	// Segundo guardado: se quita "trabajo en equipo" y queda solo la
	// huerta. El catálogo no pierde nada.
	_, err = repo.SetActivitySkills(ctx, activity, []models.Skill{
		{ID: "armado-de-huerta", Nombre: "Armado de huerta", Categoria: models.SkillCategoryNaturaleza},
	})
	require.NoError(t, err)
	links, err = repo.ListLinksByActivity(ctx, "a1")
	require.NoError(t, err)
	require.Len(t, links, 1)
	assert.Equal(t, "armado-de-huerta", links[0].HabilidadID)
	catalog, err = repo.ListCatalog(ctx)
	require.NoError(t, err)
	assert.Len(t, catalog, 2)

	// Lista vacía: quita todos los vínculos.
	_, err = repo.SetActivitySkills(ctx, activity, nil)
	require.NoError(t, err)
	links, err = repo.ListLinksByActivity(ctx, "a1")
	require.NoError(t, err)
	assert.Empty(t, links)
}

func TestSkillRepository_GetCatalog(t *testing.T) {
	client := newTestFirestore(t)
	wipeCollection(t, client, skillsCollection)
	repo := NewSkillRepository(client)
	ctx := t.Context()

	_, err := client.Collection(skillsCollection).Doc("uso-de-excel").Set(ctx, models.Skill{Nombre: "Uso de Excel", Categoria: models.SkillCategoryTecnologia})
	require.NoError(t, err)

	got, err := repo.GetCatalog(ctx, []string{"uso-de-excel", "no-existe"})
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "Uso de Excel", got["uso-de-excel"].Nombre)

	empty, err := repo.GetCatalog(ctx, nil)
	require.NoError(t, err)
	assert.Empty(t, empty)
}

func TestSkillRepository_GrantToParticipants_IsIdempotentAndKeepsFirstDate(t *testing.T) {
	client := newTestFirestore(t)
	wipeCollection(t, client, participantSkillsCollection)
	repo := NewSkillRepository(client)
	ctx := t.Context()

	primera := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	segunda := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)

	require.NoError(t, repo.GrantToParticipants(ctx, []string{"p1", "p2"}, []string{"uso-de-excel"}, models.SkillGrant{Origen: models.SkillOrigenActividad, Fecha: primera, ActividadID: "a1", CursoID: "c1"}))
	require.NoError(t, repo.GrantToParticipants(ctx, []string{"p1"}, []string{"uso-de-excel", "armado-de-huerta"}, models.SkillGrant{Origen: models.SkillOrigenActividad, Fecha: segunda, ActividadID: "a2", CursoID: "c2"}))

	p1, err := repo.ListByParticipant(ctx, "p1")
	require.NoError(t, err)
	require.Len(t, p1, 2, "no se duplica la que ya tenía")
	for _, ps := range p1 {
		if ps.HabilidadID == "uso-de-excel" {
			assert.True(t, primera.Equal(ps.Fecha), "conserva la primera vez que la ganó")
			assert.Equal(t, "c1", ps.CursoID, "y el curso donde la ganó")
		} else {
			assert.Equal(t, "a2", ps.ActividadID)
			assert.Equal(t, "c2", ps.CursoID)
		}
		assert.Equal(t, models.SkillOrigenActividad, ps.Origen)
	}

	p2, err := repo.ListByParticipant(ctx, "p2")
	require.NoError(t, err)
	assert.Len(t, p2, 1)

	all, err := repo.ListAllParticipantSkills(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 3, "p1 con dos y p2 con una")
}

func TestActivityRepository_Delete_RemovesSkillLinksButKeepsParticipantSkills(t *testing.T) {
	client := newTestFirestore(t)
	for _, c := range []string{activitiesCollection, skillsCollection, activitySkillsCollection, participantSkillsCollection, submissionsCollection, attendancesCollection} {
		wipeCollection(t, client, c)
	}
	ctx := t.Context()
	activities := NewActivityRepository(client)
	skills := NewSkillRepository(client)

	a, err := activities.Create(ctx, "c1", models.ActivityInput{Titulo: "Huerta", Consigna: "Armá un cantero", Fecha: time.Now().UTC()})
	require.NoError(t, err)
	_, err = skills.SetActivitySkills(ctx, a, []models.Skill{{ID: "armado-de-huerta", Nombre: "Armado de huerta", Categoria: models.SkillCategoryNaturaleza}})
	require.NoError(t, err)
	require.NoError(t, skills.GrantToParticipants(ctx, []string{"p1"}, []string{"armado-de-huerta"}, models.SkillGrant{Origen: models.SkillOrigenActividad, Fecha: time.Now().UTC()}))

	require.NoError(t, activities.Delete(ctx, a.ID))

	links, err := skills.ListLinksByCourse(ctx, "c1")
	require.NoError(t, err)
	assert.Empty(t, links, "el curso ya no muestra habilidades de la actividad borrada")
	owned, err := skills.ListByParticipant(ctx, "p1")
	require.NoError(t, err)
	assert.Len(t, owned, 1, "lo que el participante ya ganó se conserva")
}

func TestSubmissionRepository_CreateWithAttendance_GrantsActivitySkills(t *testing.T) {
	client := newTestFirestore(t)
	for _, c := range []string{activitiesCollection, attendancesCollection, submissionsCollection, skillsCollection, activitySkillsCollection, participantSkillsCollection} {
		wipeCollection(t, client, c)
	}
	ctx := t.Context()
	activities := NewActivityRepository(client)
	skills := NewSkillRepository(client)
	subs := NewSubmissionRepository(client)

	a, err := activities.Create(ctx, "c1", models.ActivityInput{Titulo: "Canteros", Consigna: "Armá un cantero", Fecha: time.Now().UTC()})
	require.NoError(t, err)
	_, err = skills.SetActivitySkills(ctx, a, []models.Skill{
		{ID: "armado-de-huerta", Nombre: "Armado de huerta", Categoria: models.SkillCategoryNaturaleza},
		{ID: "trabajo-en-equipo", Nombre: "Trabajo en equipo", Categoria: models.SkillCategorySocioemoc},
	})
	require.NoError(t, err)

	// p1 ya tenía "Trabajo en equipo" de otra actividad, de antes.
	antes := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	require.NoError(t, skills.GrantToParticipants(ctx, []string{"p1"}, []string{"trabajo-en-equipo"}, models.SkillGrant{Origen: models.SkillOrigenActividad, Fecha: antes}))

	entregada := time.Date(2026, 10, 9, 18, 0, 0, 0, time.UTC)
	_, err = subs.CreateWithAttendance(ctx, a, models.Submission{ParticipanteID: "p1", Texto: "listo", FechaEntrega: entregada})
	require.NoError(t, err)

	owned, err := skills.ListByParticipant(ctx, "p1")
	require.NoError(t, err)
	require.Len(t, owned, 2)
	for _, ps := range owned {
		assert.Equal(t, models.SkillOrigenActividad, ps.Origen)
		switch ps.HabilidadID {
		case "armado-de-huerta":
			assert.True(t, entregada.Equal(ps.Fecha), "la nueva toma la fecha de la entrega")
			assert.Equal(t, a.ID, ps.ActividadID, "guarda la actividad entregada")
			assert.Equal(t, "c1", ps.CursoID, "y su curso")
		case "trabajo-en-equipo":
			assert.True(t, antes.Equal(ps.Fecha), "la que ya tenía no se pisa")
		}
	}

	// Una actividad sin habilidades: la entrega funciona igual.
	b, err := activities.Create(ctx, "c1", models.ActivityInput{Titulo: "Sin habilidades", Consigna: "x", Fecha: time.Now().UTC()})
	require.NoError(t, err)
	_, err = subs.CreateWithAttendance(ctx, b, models.Submission{ParticipanteID: "p2", FechaEntrega: entregada})
	require.NoError(t, err)
	none, err := skills.ListByParticipant(ctx, "p2")
	require.NoError(t, err)
	assert.Empty(t, none)

	// Segunda entrega de la misma actividad: sigue fallando como antes y
	// no asigna nada de más.
	_, err = subs.CreateWithAttendance(ctx, a, models.Submission{ParticipanteID: "p1", FechaEntrega: entregada})
	assert.ErrorIs(t, err, ErrAlreadySubmitted)
}

func TestSkillRepository_ListLinksBySkill(t *testing.T) {
	client := newTestFirestore(t)
	for _, c := range []string{activitiesCollection, skillsCollection, activitySkillsCollection} {
		wipeCollection(t, client, c)
	}
	ctx := t.Context()
	activities := NewActivityRepository(client)
	skills := NewSkillRepository(client)
	huerta := models.Skill{ID: "armado-de-huerta", Nombre: "Armado de huerta", Categoria: models.SkillCategoryNaturaleza}

	a, err := activities.Create(ctx, "c1", models.ActivityInput{Titulo: "A", Consigna: "x", Fecha: time.Now().UTC()})
	require.NoError(t, err)
	b, err := activities.Create(ctx, "c2", models.ActivityInput{Titulo: "B", Consigna: "x", Fecha: time.Now().UTC()})
	require.NoError(t, err)
	_, err = skills.SetActivitySkills(ctx, a, []models.Skill{huerta})
	require.NoError(t, err)
	_, err = skills.SetActivitySkills(ctx, b, []models.Skill{huerta})
	require.NoError(t, err)

	links, err := skills.ListLinksBySkill(ctx, "armado-de-huerta")
	require.NoError(t, err)
	cursos := []string{}
	for _, l := range links {
		cursos = append(cursos, l.CursoID)
	}
	assert.ElementsMatch(t, []string{"c1", "c2"}, cursos)
}
