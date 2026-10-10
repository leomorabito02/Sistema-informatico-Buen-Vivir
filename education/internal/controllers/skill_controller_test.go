package controllers_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"buenvivir/education/internal/controllers"
	"buenvivir/education/internal/models"
	"buenvivir/education/internal/services"
)

type MockSkillService struct {
	mock.Mock
}

func (m *MockSkillService) SearchCatalog(ctx context.Context, query string) ([]models.Skill, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Skill), args.Error(1)
}

func (m *MockSkillService) GetCourseSkills(ctx context.Context, cursoID string) (*models.CourseSkills, error) {
	args := m.Called(ctx, cursoID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.CourseSkills), args.Error(1)
}

func (m *MockSkillService) SetActivitySkills(ctx context.Context, activity *models.Activity, inputs []models.SkillInput) ([]models.Skill, error) {
	args := m.Called(ctx, activity, inputs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Skill), args.Error(1)
}

func (m *MockSkillService) ListGrantedSkills(ctx context.Context) ([]models.GrantedSkill, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.GrantedSkill), args.Error(1)
}

func (m *MockSkillService) ListParticipantSkills(ctx context.Context, participanteID string) ([]models.ParticipantSkillView, error) {
	args := m.Called(ctx, participanteID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ParticipantSkillView), args.Error(1)
}

var testSkill = models.Skill{ID: "uso-de-excel", Nombre: "Uso de Excel", Categoria: models.SkillCategoryTecnologia}

func newSkillController() (*controllers.SkillController, *MockCourseService, *MockActivityService, *MockSkillService) {
	courseSvc := new(MockCourseService)
	activitySvc := new(MockActivityService)
	skillSvc := new(MockSkillService)
	return controllers.NewSkillController(courseSvc, activitySvc, skillSvc, nil, nil), courseSvc, activitySvc, skillSvc
}

func TestSkillController_SearchCatalog(t *testing.T) {
	ctrl, _, _, skillSvc := newSkillController()
	skillSvc.On("SearchCatalog", mock.Anything, "exc").Return([]models.Skill{testSkill}, nil)

	w := httptest.NewRecorder()
	ctrl.SearchCatalog(w, httptest.NewRequest("GET", "/api/habilidades?q=exc", nil))

	assert.Equal(t, http.StatusOK, w.Code)
	var got []models.Skill
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, []models.Skill{testSkill}, got)
}

func TestSkillController_SearchCatalog_Error(t *testing.T) {
	ctrl, _, _, skillSvc := newSkillController()
	skillSvc.On("SearchCatalog", mock.Anything, "").Return(nil, errors.New("caído"))

	w := httptest.NewRecorder()
	ctrl.SearchCatalog(w, httptest.NewRequest("GET", "/api/habilidades", nil))
	assert.Equal(t, http.StatusBadGateway, w.Code)
}

func TestSkillController_Categories(t *testing.T) {
	ctrl, _, _, _ := newSkillController()
	w := httptest.NewRecorder()
	ctrl.Categories(w, httptest.NewRequest("GET", "/api/habilidades/categorias", nil))

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Cuidado de la naturaleza")
}

func TestSkillController_CourseSkills(t *testing.T) {
	ctrl, _, _, skillSvc := newSkillController()
	skillSvc.On("GetCourseSkills", mock.Anything, "c1").Return(&models.CourseSkills{
		Habilidades:  []models.Skill{testSkill},
		PorActividad: map[string][]models.Skill{"a1": {testSkill}},
	}, nil)

	w := httptest.NewRecorder()
	ctrl.CourseSkills(w, withChiURLParams(httptest.NewRequest("GET", "/api/cursos/c1/habilidades", nil), "id", "c1"))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"porActividad"`)

	skillSvc2 := new(MockSkillService)
	skillSvc2.On("GetCourseSkills", mock.Anything, "c1").Return(nil, services.NewValidationError("x"))
	w = httptest.NewRecorder()
	controllers.NewSkillController(nil, nil, skillSvc2, nil, nil).CourseSkills(w, withChiURLParams(httptest.NewRequest("GET", "/", nil), "id", "c1"))
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func setSkillsRequest(body, userID, role string) *http.Request {
	req := httptest.NewRequest("PUT", "/api/actividades/a1/habilidades", strings.NewReader(body))
	return withChiURLParams(withAuth(req, userID, role, false), "actividadId", "a1")
}

func TestSkillController_SetActivitySkills_AdminAnyCourse(t *testing.T) {
	ctrl, courseSvc, activitySvc, skillSvc := newSkillController()
	activity := &models.Activity{ID: "a1", CursoID: "c1"}
	activitySvc.On("GetActivity", mock.Anything, "a1").Return(activity, nil)
	courseSvc.On("VerifyExists", mock.Anything, "c1").Return(nil)
	input := []models.SkillInput{{Nombre: "Uso de Excel", Categoria: models.SkillCategoryTecnologia}}
	skillSvc.On("SetActivitySkills", mock.Anything, activity, input).Return([]models.Skill{testSkill}, nil)

	w := httptest.NewRecorder()
	ctrl.SetActivitySkills(w, setSkillsRequest(`{"habilidades":[{"nombre":"Uso de Excel","categoria":"Tecnología"}]}`, "admin1", "role_admin"))

	assert.Equal(t, http.StatusOK, w.Code)
	courseSvc.AssertNotCalled(t, "VerifyOwnership", mock.Anything, mock.Anything, mock.Anything)
}

func TestSkillController_SetActivitySkills_TeacherNeedsOwnership(t *testing.T) {
	ctrl, courseSvc, activitySvc, skillSvc := newSkillController()
	activity := &models.Activity{ID: "a1", CursoID: "c1"}
	activitySvc.On("GetActivity", mock.Anything, "a1").Return(activity, nil)
	courseSvc.On("VerifyOwnership", mock.Anything, "c1", "doc1").Return(services.NewForbiddenError("no sos el docente responsable de este curso"))

	w := httptest.NewRecorder()
	ctrl.SetActivitySkills(w, setSkillsRequest(`{"habilidades":[]}`, "doc1", "role_teacher"))

	assert.Equal(t, http.StatusForbidden, w.Code)
	skillSvc.AssertNotCalled(t, "SetActivitySkills", mock.Anything, mock.Anything, mock.Anything)
}

func TestSkillController_SetActivitySkills_TeacherOwnCourse(t *testing.T) {
	ctrl, courseSvc, activitySvc, skillSvc := newSkillController()
	activity := &models.Activity{ID: "a1", CursoID: "c1"}
	activitySvc.On("GetActivity", mock.Anything, "a1").Return(activity, nil)
	courseSvc.On("VerifyOwnership", mock.Anything, "c1", "doc1").Return(nil)
	skillSvc.On("SetActivitySkills", mock.Anything, activity, []models.SkillInput{}).Return([]models.Skill{}, nil)

	w := httptest.NewRecorder()
	ctrl.SetActivitySkills(w, setSkillsRequest(`{"habilidades":[]}`, "doc1", "role_teacher"))
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSkillController_SetActivitySkills_Errors(t *testing.T) {
	t.Run("actividad inexistente", func(t *testing.T) {
		ctrl, _, activitySvc, _ := newSkillController()
		activitySvc.On("GetActivity", mock.Anything, "a1").Return(nil, repoErrNotFound())
		w := httptest.NewRecorder()
		ctrl.SetActivitySkills(w, setSkillsRequest(`{}`, "u", "role_admin"))
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
	t.Run("body inválido", func(t *testing.T) {
		ctrl, courseSvc, activitySvc, _ := newSkillController()
		activitySvc.On("GetActivity", mock.Anything, "a1").Return(&models.Activity{ID: "a1", CursoID: "c1"}, nil)
		courseSvc.On("VerifyExists", mock.Anything, "c1").Return(nil)
		w := httptest.NewRecorder()
		ctrl.SetActivitySkills(w, setSkillsRequest(`{no es json`, "u", "role_volunteer"))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	t.Run("validación del service", func(t *testing.T) {
		ctrl, courseSvc, activitySvc, skillSvc := newSkillController()
		activity := &models.Activity{ID: "a1", CursoID: "c1"}
		activitySvc.On("GetActivity", mock.Anything, "a1").Return(activity, nil)
		courseSvc.On("VerifyExists", mock.Anything, "c1").Return(nil)
		skillSvc.On("SetActivitySkills", mock.Anything, activity, mock.Anything).Return(nil, services.NewValidationError("categoría inválida"))
		w := httptest.NewRecorder()
		ctrl.SetActivitySkills(w, setSkillsRequest(`{"habilidades":[{"nombre":"x","categoria":"y"}]}`, "u", "role_admin"))
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "categoría inválida")
	})
}

type MockSkillSuggestionService struct {
	mock.Mock
}

func (m *MockSkillSuggestionService) Suggest(ctx context.Context, req models.SkillSuggestionRequest, course *models.Course) (*models.SkillSuggestionResponse, error) {
	args := m.Called(ctx, req, course)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SkillSuggestionResponse), args.Error(1)
}

func newSuggestController() (*controllers.SkillController, *MockCourseService, *MockActivityService, *MockSkillSuggestionService) {
	courseSvc := new(MockCourseService)
	activitySvc := new(MockActivityService)
	suggestSvc := new(MockSkillSuggestionService)
	return controllers.NewSkillController(courseSvc, activitySvc, new(MockSkillService), suggestSvc, nil), courseSvc, activitySvc, suggestSvc
}

func suggestRequest(body, userID, role string) *http.Request {
	return withAuth(httptest.NewRequest("POST", "/api/actividades/sugerir-habilidades", strings.NewReader(body)), userID, role, false)
}

func TestSkillController_Suggest_Success(t *testing.T) {
	ctrl, courseSvc, activitySvc, suggestSvc := newSuggestController()
	course := &models.Course{ID: "c1", Title: "Huerta"}
	courseSvc.On("VerifyExists", mock.Anything, "c1").Return(nil)
	activitySvc.On("GetActivity", mock.Anything, "a1").Return(&models.Activity{ID: "a1", CursoID: "c1"}, nil)
	courseSvc.On("GetCourse", mock.Anything, "c1").Return(course, nil)
	suggestSvc.On("Suggest", mock.Anything, mock.MatchedBy(func(r models.SkillSuggestionRequest) bool { return r.Titulo == "Canteros" }), course).
		Return(&models.SkillSuggestionResponse{Proveedor: "gemini", Sugerencias: []models.SkillSuggestion{{Nombre: "Compostaje"}}}, nil)

	w := httptest.NewRecorder()
	ctrl.Suggest(w, suggestRequest(`{"cursoId":"c1","actividadId":"a1","titulo":"Canteros","consigna":"x"}`, "admin1", "role_admin"))

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"proveedor":"gemini"`)
}

func TestSkillController_Suggest_DraftCourseWithoutContext(t *testing.T) {
	ctrl, courseSvc, _, suggestSvc := newSuggestController()
	courseSvc.On("VerifyExists", mock.Anything, "c1").Return(nil)
	courseSvc.On("GetCourse", mock.Anything, "c1").Return(nil, repoErrNotFound())
	suggestSvc.On("Suggest", mock.Anything, mock.Anything, (*models.Course)(nil)).
		Return(&models.SkillSuggestionResponse{Proveedor: "openai", Sugerencias: []models.SkillSuggestion{}}, nil)

	w := httptest.NewRecorder()
	ctrl.Suggest(w, suggestRequest(`{"cursoId":"c1","titulo":"t","consigna":"c"}`, "v1", "role_volunteer"))
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSkillController_Suggest_Errors(t *testing.T) {
	t.Run("body inválido", func(t *testing.T) {
		ctrl, _, _, _ := newSuggestController()
		w := httptest.NewRecorder()
		ctrl.Suggest(w, suggestRequest(`{`, "u", "role_admin"))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	t.Run("docente de otro curso", func(t *testing.T) {
		ctrl, courseSvc, _, suggestSvc := newSuggestController()
		courseSvc.On("VerifyOwnership", mock.Anything, "c1", "doc1").Return(services.NewForbiddenError("no sos el docente responsable de este curso"))
		w := httptest.NewRecorder()
		ctrl.Suggest(w, suggestRequest(`{"cursoId":"c1","titulo":"t","consigna":"c"}`, "doc1", "role_teacher"))
		assert.Equal(t, http.StatusForbidden, w.Code)
		suggestSvc.AssertNotCalled(t, "Suggest", mock.Anything, mock.Anything, mock.Anything)
	})
	t.Run("actividad de otro curso", func(t *testing.T) {
		ctrl, courseSvc, activitySvc, _ := newSuggestController()
		courseSvc.On("VerifyOwnership", mock.Anything, "c1", "doc1").Return(nil)
		activitySvc.On("GetActivity", mock.Anything, "a9").Return(&models.Activity{ID: "a9", CursoID: "otro"}, nil)
		w := httptest.NewRecorder()
		ctrl.Suggest(w, suggestRequest(`{"cursoId":"c1","actividadId":"a9","titulo":"t","consigna":"c"}`, "doc1", "role_teacher"))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	t.Run("actividad inexistente", func(t *testing.T) {
		ctrl, courseSvc, activitySvc, _ := newSuggestController()
		courseSvc.On("VerifyExists", mock.Anything, "c1").Return(nil)
		activitySvc.On("GetActivity", mock.Anything, "a9").Return(nil, repoErrNotFound())
		w := httptest.NewRecorder()
		ctrl.Suggest(w, suggestRequest(`{"cursoId":"c1","actividadId":"a9","titulo":"t","consigna":"c"}`, "u", "role_admin"))
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
	t.Run("ninguna IA respondió", func(t *testing.T) {
		ctrl, courseSvc, _, suggestSvc := newSuggestController()
		courseSvc.On("VerifyExists", mock.Anything, "c1").Return(nil)
		courseSvc.On("GetCourse", mock.Anything, "c1").Return(nil, repoErrNotFound())
		suggestSvc.On("Suggest", mock.Anything, mock.Anything, mock.Anything).Return(nil, services.NewUnavailableError(services.MsgSuggestionsUnavailable))
		w := httptest.NewRecorder()
		ctrl.Suggest(w, suggestRequest(`{"cursoId":"c1","titulo":"t","consigna":"c"}`, "u", "role_admin"))
		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
		assert.Contains(t, w.Body.String(), "Podés cargarlas a mano")
	})
}

type fakeResolver struct {
	id  string
	err error
}

func (f fakeResolver) ResolveParticipantID(context.Context, string) (string, error) {
	return f.id, f.err
}

func TestSkillController_ParticipantSkills(t *testing.T) {
	skillSvc := new(MockSkillService)
	skillSvc.On("ListParticipantSkills", mock.Anything, "p1").Return([]models.ParticipantSkillView{{Skill: testSkill, Origen: "actividad"}}, nil)
	ctrl := controllers.NewSkillController(nil, nil, skillSvc, nil, nil)

	w := httptest.NewRecorder()
	ctrl.ParticipantSkills(w, withChiURLParams(httptest.NewRequest("GET", "/api/participants/p1/habilidades", nil), "id", "p1"))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Uso de Excel")
}

func TestSkillController_MySkills_UsesIDFromToken(t *testing.T) {
	skillSvc := new(MockSkillService)
	skillSvc.On("ListParticipantSkills", mock.Anything, "p-del-token").Return([]models.ParticipantSkillView{}, nil)
	ctrl := controllers.NewSkillController(nil, nil, skillSvc, nil, fakeResolver{id: "p-del-token"})

	w := httptest.NewRecorder()
	ctrl.MySkills(w, withAuth(httptest.NewRequest("GET", "/api/participante/habilidades?id=otro", nil), "u1", "role_participant", false))
	assert.Equal(t, http.StatusOK, w.Code)
	skillSvc.AssertExpectations(t)
}

func TestSkillController_MySkills_Errors(t *testing.T) {
	ctrl := controllers.NewSkillController(nil, nil, new(MockSkillService), nil, fakeResolver{err: services.NewForbiddenError("no sos participante")})
	w := httptest.NewRecorder()
	ctrl.MySkills(w, withAuth(httptest.NewRequest("GET", "/api/participante/habilidades", nil), "u1", "role_participant", false))
	assert.Equal(t, http.StatusForbidden, w.Code)

	skillSvc := new(MockSkillService)
	skillSvc.On("ListParticipantSkills", mock.Anything, "p1").Return(nil, errors.New("caído"))
	ctrl = controllers.NewSkillController(nil, nil, skillSvc, nil, fakeResolver{id: "p1"})
	w = httptest.NewRecorder()
	ctrl.MySkills(w, withAuth(httptest.NewRequest("GET", "/api/participante/habilidades", nil), "u1", "role_participant", false))
	assert.Equal(t, http.StatusBadGateway, w.Code)
}

func TestSkillController_GrantedSkills(t *testing.T) {
	skillSvc := new(MockSkillService)
	skillSvc.On("ListGrantedSkills", mock.Anything).Return([]models.GrantedSkill{
		{ParticipanteID: "p1", HabilidadID: testSkill.ID, Nombre: testSkill.Nombre, CursoID: "c1"},
	}, nil).Once()
	ctrl := controllers.NewSkillController(nil, nil, skillSvc, nil, nil)

	w := httptest.NewRecorder()
	ctrl.GrantedSkills(w, httptest.NewRequest("GET", "/api/habilidades/otorgadas", nil))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"cursoId":"c1"`)

	skillSvc.On("ListGrantedSkills", mock.Anything).Return(nil, errors.New("firestore caído"))
	w = httptest.NewRecorder()
	ctrl.GrantedSkills(w, httptest.NewRequest("GET", "/api/habilidades/otorgadas", nil))
	assert.Equal(t, http.StatusBadGateway, w.Code)
}
