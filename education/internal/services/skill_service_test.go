package services_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"buenvivir/education/internal/models"
	"buenvivir/education/internal/services"
)

type MockSkillRepository struct {
	mock.Mock
}

func (m *MockSkillRepository) ListCatalog(ctx context.Context) ([]models.Skill, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Skill), args.Error(1)
}

func (m *MockSkillRepository) GetCatalog(ctx context.Context, ids []string) (map[string]models.Skill, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]models.Skill), args.Error(1)
}

func (m *MockSkillRepository) ListLinksByActivity(ctx context.Context, actividadID string) ([]models.ActivitySkill, error) {
	args := m.Called(ctx, actividadID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ActivitySkill), args.Error(1)
}

func (m *MockSkillRepository) ListLinksByCourse(ctx context.Context, cursoID string) ([]models.ActivitySkill, error) {
	args := m.Called(ctx, cursoID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ActivitySkill), args.Error(1)
}

func (m *MockSkillRepository) SetActivitySkills(ctx context.Context, activity *models.Activity, skills []models.Skill) ([]models.Skill, error) {
	args := m.Called(ctx, activity, skills)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Skill), args.Error(1)
}

func (m *MockSkillRepository) GrantToParticipants(ctx context.Context, participanteIDs, habilidadIDs []string, grant models.SkillGrant) error {
	return m.Called(ctx, participanteIDs, habilidadIDs, grant).Error(0)
}

func (m *MockSkillRepository) ListAllParticipantSkills(ctx context.Context) ([]models.ParticipantSkill, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ParticipantSkill), args.Error(1)
}

func (m *MockSkillRepository) ListLinksBySkill(ctx context.Context, habilidadID string) ([]models.ActivitySkill, error) {
	args := m.Called(ctx, habilidadID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ActivitySkill), args.Error(1)
}

func (m *MockSkillRepository) ListByParticipant(ctx context.Context, participanteID string) ([]models.ParticipantSkill, error) {
	args := m.Called(ctx, participanteID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ParticipantSkill), args.Error(1)
}

var (
	skHuerta = models.Skill{ID: "armado-de-huerta", Nombre: "Armado de huerta", Categoria: models.SkillCategoryNaturaleza}
	skEquipo = models.Skill{ID: "trabajo-en-equipo", Nombre: "Trabajo en equipo", Categoria: models.SkillCategorySocioemoc}
	skExcel  = models.Skill{ID: "uso-de-excel", Nombre: "Uso de Excel", Categoria: models.SkillCategoryTecnologia}
)

func newSkillService() (services.SkillService, *MockSkillRepository, *MockSubmissionRepository) {
	repo := new(MockSkillRepository)
	subs := new(MockSubmissionRepository)
	return services.NewSkillService(repo, subs), repo, subs
}

func TestSkillService_SearchCatalog_IgnoresCaseAndAccents(t *testing.T) {
	svc, repo, _ := newSkillService()
	repo.On("ListCatalog", mock.Anything).Return([]models.Skill{skExcel, skHuerta, skEquipo}, nil)

	got, err := svc.SearchCatalog(context.Background(), "EQUIPO")
	require.NoError(t, err)
	assert.Equal(t, []models.Skill{skEquipo}, got)

	all, err := svc.SearchCatalog(context.Background(), "")
	require.NoError(t, err)
	assert.Equal(t, []models.Skill{skHuerta, skEquipo, skExcel}, all, "vacío = todo el catálogo, ordenado")
}

func TestSkillService_SearchCatalog_RepoError(t *testing.T) {
	svc, repo, _ := newSkillService()
	repo.On("ListCatalog", mock.Anything).Return(nil, errors.New("firestore caído"))

	_, err := svc.SearchCatalog(context.Background(), "x")
	assert.Error(t, err)
}

func TestSkillService_GetCourseSkills_UnionWithoutRepeats(t *testing.T) {
	svc, repo, _ := newSkillService()
	repo.On("ListLinksByCourse", mock.Anything, "c1").Return([]models.ActivitySkill{
		{ActividadID: "a1", HabilidadID: skHuerta.ID, CursoID: "c1"},
		{ActividadID: "a1", HabilidadID: skEquipo.ID, CursoID: "c1"},
		{ActividadID: "a2", HabilidadID: skEquipo.ID, CursoID: "c1"},
		{ActividadID: "a2", HabilidadID: "borrada", CursoID: "c1"},
	}, nil)
	repo.On("GetCatalog", mock.Anything, []string{skHuerta.ID, skEquipo.ID, "borrada"}).
		Return(map[string]models.Skill{skHuerta.ID: skHuerta, skEquipo.ID: skEquipo}, nil)

	got, err := svc.GetCourseSkills(context.Background(), "c1")
	require.NoError(t, err)
	assert.Equal(t, []models.Skill{skHuerta, skEquipo}, got.Habilidades)
	assert.Equal(t, []models.Skill{skHuerta, skEquipo}, got.PorActividad["a1"])
	assert.Equal(t, []models.Skill{skEquipo}, got.PorActividad["a2"], "la que no está en el catálogo se omite")
}

func TestSkillService_GetCourseSkills_Validation(t *testing.T) {
	svc, _, _ := newSkillService()
	_, err := svc.GetCourseSkills(context.Background(), " ")
	assert.True(t, services.IsValidationError(err))
}

func TestSkillService_GetCourseSkills_Errors(t *testing.T) {
	t.Run("ListLinksByCourse error", func(t *testing.T) {
		// Arrange
		svc, repo, _ := newSkillService()
		repo.On("ListLinksByCourse", mock.Anything, "c1").Return(nil, errors.New("firestore down"))

		// Act
		_, err := svc.GetCourseSkills(context.Background(), "c1")

		// Assert
		assert.Error(t, err)
	})

	t.Run("GetCatalog error", func(t *testing.T) {
		// Arrange
		svc, repo, _ := newSkillService()
		repo.On("ListLinksByCourse", mock.Anything, "c1").Return([]models.ActivitySkill{
			{ActividadID: "a1", HabilidadID: skHuerta.ID, CursoID: "c1"},
		}, nil)
		repo.On("GetCatalog", mock.Anything, []string{skHuerta.ID}).Return(nil, errors.New("firestore down"))

		// Act
		_, err := svc.GetCourseSkills(context.Background(), "c1")

		// Assert
		assert.Error(t, err)
	})
}

func TestSkillService_SetActivitySkills_NormalizesAndGrantsToPreviousSubmitters(t *testing.T) {
	svc, repo, subs := newSkillService()
	activity := &models.Activity{ID: "a1", CursoID: "c1"}

	// "trabajo en  Equipo" y "Trabajo en equipo" son la misma: queda la
	// primera, con el nombre prolijo.
	repo.On("SetActivitySkills", mock.Anything, activity, []models.Skill{
		{ID: "trabajo-en-equipo", Nombre: "Trabajo en Equipo", Categoria: models.SkillCategorySocioemoc},
		{ID: "armado-de-huerta", Nombre: "Armado de huerta", Categoria: models.SkillCategoryNaturaleza},
	}).Return([]models.Skill{skHuerta, skEquipo}, nil)
	subs.On("ListByActivity", mock.Anything, "a1").Return([]models.Submission{
		{ParticipanteID: "p1"}, {ParticipanteID: "p2"},
	}, nil)
	repo.On("GrantToParticipants", mock.Anything, []string{"p1", "p2"}, []string{skHuerta.ID, skEquipo.ID}, mock.MatchedBy(func(g models.SkillGrant) bool {
		return g.Origen == models.SkillOrigenActividad && g.ActividadID == "a1" && g.CursoID == "c1" && !g.Fecha.IsZero()
	})).Return(nil)

	saved, err := svc.SetActivitySkills(context.Background(), activity, []models.SkillInput{
		{Nombre: "trabajo en  Equipo", Categoria: models.SkillCategorySocioemoc},
		{Nombre: "Trabajo en equipo", Categoria: models.SkillCategorySocioemoc},
		{Nombre: "armado de huerta", Categoria: models.SkillCategoryNaturaleza},
	})
	require.NoError(t, err)
	assert.Equal(t, []models.Skill{skHuerta, skEquipo}, saved)
	repo.AssertExpectations(t)
}

func TestSkillService_SetActivitySkills_EmptyListRemovesAllWithoutGranting(t *testing.T) {
	svc, repo, subs := newSkillService()
	activity := &models.Activity{ID: "a1", CursoID: "c1"}
	repo.On("SetActivitySkills", mock.Anything, activity, []models.Skill{}).Return([]models.Skill{}, nil)

	saved, err := svc.SetActivitySkills(context.Background(), activity, nil)
	require.NoError(t, err)
	assert.Empty(t, saved)
	subs.AssertNotCalled(t, "ListByActivity", mock.Anything, mock.Anything)
}

func TestSkillService_SetActivitySkills_NoSubmissionsNoGrant(t *testing.T) {
	svc, repo, subs := newSkillService()
	activity := &models.Activity{ID: "a1", CursoID: "c1"}
	repo.On("SetActivitySkills", mock.Anything, activity, mock.Anything).Return([]models.Skill{skExcel}, nil)
	subs.On("ListByActivity", mock.Anything, "a1").Return([]models.Submission{}, nil)

	_, err := svc.SetActivitySkills(context.Background(), activity, []models.SkillInput{{Nombre: "Uso de Excel", Categoria: models.SkillCategoryTecnologia}})
	require.NoError(t, err)
	repo.AssertNotCalled(t, "GrantToParticipants", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestSkillService_SetActivitySkills_GrantFailureIsReported(t *testing.T) {
	svc, repo, subs := newSkillService()
	activity := &models.Activity{ID: "a1", CursoID: "c1"}
	repo.On("SetActivitySkills", mock.Anything, activity, mock.Anything).Return([]models.Skill{skExcel}, nil)
	subs.On("ListByActivity", mock.Anything, "a1").Return([]models.Submission{{ParticipanteID: "p1"}}, nil)
	repo.On("GrantToParticipants", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("firestore caído"))

	_, err := svc.SetActivitySkills(context.Background(), activity, []models.SkillInput{{Nombre: "Uso de Excel", Categoria: models.SkillCategoryTecnologia}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "volvé a guardar")
}

func TestSkillService_SetActivitySkills_RepoError(t *testing.T) {
	svc, repo, _ := newSkillService()
	activity := &models.Activity{ID: "a1"}
	repo.On("SetActivitySkills", mock.Anything, activity, mock.Anything).Return(nil, errors.New("tx falló"))

	_, err := svc.SetActivitySkills(context.Background(), activity, []models.SkillInput{{Nombre: "Uso de Excel", Categoria: models.SkillCategoryTecnologia}})
	assert.Error(t, err)
}

func TestSkillService_SetActivitySkills_ListSubmissionsError(t *testing.T) {
	svc, repo, subs := newSkillService()
	activity := &models.Activity{ID: "a1"}
	repo.On("SetActivitySkills", mock.Anything, activity, mock.Anything).Return([]models.Skill{skExcel}, nil)
	subs.On("ListByActivity", mock.Anything, "a1").Return(nil, errors.New("firestore caído"))

	_, err := svc.SetActivitySkills(context.Background(), activity, []models.SkillInput{{Nombre: "Uso de Excel", Categoria: models.SkillCategoryTecnologia}})
	assert.Error(t, err)
}

func TestSkillService_SetActivitySkills_Validation(t *testing.T) {
	svc, _, _ := newSkillService()
	activity := &models.Activity{ID: "a1"}
	cases := map[string][]models.SkillInput{
		"sin nombre":           {{Nombre: "  ", Categoria: models.SkillCategoryOficio}},
		"categoría inválida":   {{Nombre: "Soldadura", Categoria: "Técnica"}},
		"nombre muy largo":     {{Nombre: "Aprender a trabajar con otras personas respetando los tiempos de cada uno", Categoria: models.SkillCategorySocioemoc}},
		"más de 8 habilidades": manySkills(9),
	}
	for name, in := range cases {
		_, err := svc.SetActivitySkills(context.Background(), activity, in)
		assert.True(t, services.IsValidationError(err), name)
	}
}

func manySkills(n int) []models.SkillInput {
	out := make([]models.SkillInput, n)
	for i := range out {
		out[i] = models.SkillInput{Nombre: "Habilidad " + string(rune('A'+i)), Categoria: models.SkillCategoryOficio}
	}
	return out
}

func TestSkillService_ListParticipantSkills(t *testing.T) {
	svc, repo, _ := newSkillService()
	fecha := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	repo.On("ListByParticipant", mock.Anything, "p1").Return([]models.ParticipantSkill{
		{ParticipanteID: "p1", HabilidadID: skExcel.ID, Origen: "actividad", Fecha: fecha},
		{ParticipanteID: "p1", HabilidadID: skHuerta.ID, Origen: "actividad", Fecha: fecha},
		{ParticipanteID: "p1", HabilidadID: "borrada", Origen: "actividad", Fecha: fecha},
	}, nil)
	repo.On("GetCatalog", mock.Anything, []string{skExcel.ID, skHuerta.ID, "borrada"}).
		Return(map[string]models.Skill{skExcel.ID: skExcel, skHuerta.ID: skHuerta}, nil)

	got, err := svc.ListParticipantSkills(context.Background(), "p1")
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, skHuerta, got[0].Skill)
	assert.Equal(t, skExcel, got[1].Skill)
	assert.Equal(t, "actividad", got[0].Origen)
	assert.True(t, fecha.Equal(got[0].Fecha))
}

func TestSkillService_ListParticipantSkills_Errors(t *testing.T) {
	svc, repo, _ := newSkillService()
	_, err := svc.ListParticipantSkills(context.Background(), "")
	assert.True(t, services.IsValidationError(err))

	repo.On("ListByParticipant", mock.Anything, "p1").Return(nil, errors.New("caído"))
	_, err = svc.ListParticipantSkills(context.Background(), "p1")
	assert.Error(t, err)
}

func TestSkillService_ListGrantedSkills(t *testing.T) {
	svc, repo, subs := newSkillService()
	ene := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	feb := time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC)
	mar := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)

	repo.On("ListAllParticipantSkills", mock.Anything).Return([]models.ParticipantSkill{
		// Nueva: ya trae actividad y curso.
		{ParticipanteID: "p1", HabilidadID: skExcel.ID, Origen: "actividad", Fecha: mar, ActividadID: "a9", CursoID: "c9"},
		// Vieja: se completa con la primera entrega del participante.
		{ParticipanteID: "p1", HabilidadID: skHuerta.ID, Origen: "actividad", Fecha: ene},
		// Vieja sin entregas que la expliquen: queda sin curso.
		{ParticipanteID: "p2", HabilidadID: skHuerta.ID, Origen: "manual", Fecha: feb},
		// Borrada del catálogo: se omite.
		{ParticipanteID: "p2", HabilidadID: "borrada", Origen: "actividad", Fecha: feb},
	}, nil)
	repo.On("GetCatalog", mock.Anything, []string{skExcel.ID, skHuerta.ID, "borrada"}).
		Return(map[string]models.Skill{skExcel.ID: skExcel, skHuerta.ID: skHuerta}, nil)
	repo.On("ListLinksBySkill", mock.Anything, skHuerta.ID).Return([]models.ActivitySkill{
		{ActividadID: "a1", HabilidadID: skHuerta.ID, CursoID: "c1"},
		{ActividadID: "a2", HabilidadID: skHuerta.ID, CursoID: "c2"},
	}, nil).Once()
	subs.On("GetForParticipant", mock.Anything, "p1", []string{"a1", "a2"}).Return(map[string]*models.Submission{
		"a1": {ActividadID: "a1", FechaEntrega: feb},
		"a2": {ActividadID: "a2", FechaEntrega: ene},
	}, nil)
	subs.On("GetForParticipant", mock.Anything, "p2", []string{"a1", "a2"}).Return(map[string]*models.Submission{}, nil)

	got, err := svc.ListGrantedSkills(context.Background())
	require.NoError(t, err)
	require.Len(t, got, 3)
	// Ordenadas por fecha.
	assert.Equal(t, skHuerta.Nombre, got[0].Nombre)
	assert.Equal(t, "a2", got[0].ActividadID, "la entrega más antigua")
	assert.Equal(t, "c2", got[0].CursoID)
	assert.Equal(t, "p2", got[1].ParticipanteID)
	assert.Empty(t, got[1].CursoID)
	assert.Equal(t, "c9", got[2].CursoID)
	assert.Equal(t, skExcel.Categoria, got[2].Categoria)
	repo.AssertExpectations(t)
}

func TestSkillService_ListGrantedSkills_Errors(t *testing.T) {
	svc, repo, _ := newSkillService()
	repo.On("ListAllParticipantSkills", mock.Anything).Return(nil, errors.New("caído")).Once()
	_, err := svc.ListGrantedSkills(context.Background())
	assert.Error(t, err)

	repo.On("ListAllParticipantSkills", mock.Anything).Return([]models.ParticipantSkill{
		{ParticipanteID: "p1", HabilidadID: skHuerta.ID},
	}, nil)
	repo.On("GetCatalog", mock.Anything, mock.Anything).Return(nil, errors.New("caído")).Once()
	_, err = svc.ListGrantedSkills(context.Background())
	assert.Error(t, err)

	repo.On("GetCatalog", mock.Anything, mock.Anything).Return(map[string]models.Skill{skHuerta.ID: skHuerta}, nil)
	repo.On("ListLinksBySkill", mock.Anything, skHuerta.ID).Return(nil, errors.New("caído")).Once()
	_, err = svc.ListGrantedSkills(context.Background())
	assert.Error(t, err)
}

func TestSkillService_ListGrantedSkills_SubmissionsError(t *testing.T) {
	svc, repo, subs := newSkillService()
	repo.On("ListAllParticipantSkills", mock.Anything).Return([]models.ParticipantSkill{
		{ParticipanteID: "p1", HabilidadID: skHuerta.ID},
	}, nil)
	repo.On("GetCatalog", mock.Anything, mock.Anything).Return(map[string]models.Skill{skHuerta.ID: skHuerta}, nil)
	repo.On("ListLinksBySkill", mock.Anything, skHuerta.ID).Return([]models.ActivitySkill{{ActividadID: "a1", CursoID: "c1"}}, nil)
	subs.On("GetForParticipant", mock.Anything, "p1", []string{"a1"}).Return(nil, errors.New("caído"))
	_, err := svc.ListGrantedSkills(context.Background())
	assert.Error(t, err)
}

func TestSkillService_SearchCatalog_LimitBoundary(t *testing.T) {
	// Arrange
	svc, repo, _ := newSkillService()
	var catalog []models.Skill
	for i := 0; i < 25; i++ {
		catalog = append(catalog, models.Skill{
			ID:     fmt.Sprintf("skill-%02d", i),
			Nombre: fmt.Sprintf("Skill %02d", i),
		})
	}
	repo.On("ListCatalog", mock.Anything).Return(catalog, nil)

	// Act
	got, err := svc.SearchCatalog(context.Background(), "skill")

	// Assert: capped at defaultCatalogSearchLimit (20)
	require.NoError(t, err)
	assert.Len(t, got, 20)
}

func TestSkillService_ListParticipantSkills_GetCatalogError(t *testing.T) {
	// Arrange
	svc, repo, _ := newSkillService()
	repo.On("ListByParticipant", mock.Anything, "p1").Return([]models.ParticipantSkill{
		{HabilidadID: "h1"},
	}, nil)
	repo.On("GetCatalog", mock.Anything, []string{"h1"}).Return(nil, errors.New("catalog error"))

	// Act
	got, err := svc.ListParticipantSkills(context.Background(), "p1")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, got)
}
