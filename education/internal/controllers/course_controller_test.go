package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"buenvivir/education/internal/controllers"
	"buenvivir/education/internal/models"
	"buenvivir/education/internal/services"
)

type MockCourseService struct {
	mock.Mock
}

func (m *MockCourseService) ListCourses(ctx context.Context) ([]models.Course, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Course), args.Error(1)
}

func (m *MockCourseService) GetCourse(ctx context.Context, id string) (*models.Course, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Course), args.Error(1)
}

func (m *MockCourseService) CreateCourse(ctx context.Context, input models.CourseInput) (*models.Course, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Course), args.Error(1)
}

func (m *MockCourseService) UpdateCourse(ctx context.Context, id string, input models.CourseInput) (*models.Course, error) {
	args := m.Called(ctx, id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Course), args.Error(1)
}

func (m *MockCourseService) DeleteCourse(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCourseService) PublishCourse(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCourseService) UnpublishCourse(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCourseService) ListCoursesByTeacher(ctx context.Context, teacherID string) ([]models.Course, error) {
	args := m.Called(ctx, teacherID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Course), args.Error(1)
}

func (m *MockCourseService) ListDraftCoursesByTeacher(ctx context.Context, teacherID string) ([]models.Course, error) {
	args := m.Called(ctx, teacherID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Course), args.Error(1)
}

func (m *MockCourseService) ListAllDraftCourses(ctx context.Context) ([]models.Course, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Course), args.Error(1)
}

func (m *MockCourseService) CreateCourseAsTeacher(ctx context.Context, teacherID string, input models.CourseInput) (*models.Course, error) {
	args := m.Called(ctx, teacherID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Course), args.Error(1)
}

func (m *MockCourseService) UpdateCourseAsTeacher(ctx context.Context, id, teacherID string, input models.CourseInput) (*models.Course, error) {
	args := m.Called(ctx, id, teacherID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Course), args.Error(1)
}

func (m *MockCourseService) VerifyOwnership(ctx context.Context, cursoID, teacherID string) error {
	args := m.Called(ctx, cursoID, teacherID)
	return args.Error(0)
}

func (m *MockCourseService) VerifyExists(ctx context.Context, cursoID string) error {
	args := m.Called(ctx, cursoID)
	return args.Error(0)
}

// withChiURLParam simula lo que hace el router chi al popular
// chi.URLParam(r, key), necesario porque los controllers lo leen
// directo en vez de r.PathValue.
func withChiURLParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func TestCourseController_List_Success(t *testing.T) {
	mockSvc := new(MockCourseService)
	ctrl := controllers.NewCourseController(mockSvc)

	mockSvc.On("ListCourses", mock.Anything).Return([]models.Course{{ID: "c1"}}, nil)

	req := httptest.NewRequest("GET", "/api/cursos", nil)
	w := httptest.NewRecorder()
	ctrl.List(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCourseController_List_ServiceError(t *testing.T) {
	mockSvc := new(MockCourseService)
	ctrl := controllers.NewCourseController(mockSvc)

	mockSvc.On("ListCourses", mock.Anything).Return(nil, errors.New("contentful caído"))

	req := httptest.NewRequest("GET", "/api/cursos", nil)
	w := httptest.NewRecorder()
	ctrl.List(w, req)

	assert.Equal(t, http.StatusBadGateway, w.Code)
}

func TestCourseController_Get_NotFound(t *testing.T) {
	mockSvc := new(MockCourseService)
	ctrl := controllers.NewCourseController(mockSvc)

	mockSvc.On("GetCourse", mock.Anything, "c1").Return(nil, services.NewValidationError("el id del curso es requerido"))

	req := withChiURLParam(httptest.NewRequest("GET", "/api/cursos/c1", nil), "id", "c1")
	w := httptest.NewRecorder()
	ctrl.Get(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCourseController_Get_Success(t *testing.T) {
	mockSvc := new(MockCourseService)
	ctrl := controllers.NewCourseController(mockSvc)

	mockSvc.On("GetCourse", mock.Anything, "c1").Return(&models.Course{ID: "c1"}, nil)

	req := withChiURLParam(httptest.NewRequest("GET", "/api/cursos/c1", nil), "id", "c1")
	w := httptest.NewRecorder()
	ctrl.Get(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCourseController_Create_BadJSON(t *testing.T) {
	mockSvc := new(MockCourseService)
	ctrl := controllers.NewCourseController(mockSvc)

	req := httptest.NewRequest("POST", "/api/cursos", bytes.NewBufferString("{invalid"))
	w := httptest.NewRecorder()
	ctrl.Create(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCourseController_Create_ValidationError(t *testing.T) {
	mockSvc := new(MockCourseService)
	ctrl := controllers.NewCourseController(mockSvc)

	mockSvc.On("CreateCourse", mock.Anything, mock.AnythingOfType("models.CourseInput")).
		Return(nil, services.NewValidationError("el campo 'title' es requerido"))

	body, _ := json.Marshal(models.CourseInput{})
	req := httptest.NewRequest("POST", "/api/cursos", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	ctrl.Create(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCourseController_Create_Success(t *testing.T) {
	mockSvc := new(MockCourseService)
	ctrl := controllers.NewCourseController(mockSvc)

	mockSvc.On("CreateCourse", mock.Anything, mock.AnythingOfType("models.CourseInput")).
		Return(&models.Course{ID: "c1"}, nil)

	body, _ := json.Marshal(models.CourseInput{Title: "Curso"})
	req := httptest.NewRequest("POST", "/api/cursos", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	ctrl.Create(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestCourseController_Update_BadJSON(t *testing.T) {
	mockSvc := new(MockCourseService)
	ctrl := controllers.NewCourseController(mockSvc)

	req := withChiURLParam(httptest.NewRequest("PUT", "/api/cursos/c1", bytes.NewBufferString("{invalid")), "id", "c1")
	w := httptest.NewRecorder()
	ctrl.Update(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCourseController_Update_Success(t *testing.T) {
	mockSvc := new(MockCourseService)
	ctrl := controllers.NewCourseController(mockSvc)

	mockSvc.On("UpdateCourse", mock.Anything, "c1", mock.AnythingOfType("models.CourseInput")).
		Return(&models.Course{ID: "c1"}, nil)

	body, _ := json.Marshal(models.CourseInput{Title: "Curso"})
	req := withChiURLParam(httptest.NewRequest("PUT", "/api/cursos/c1", bytes.NewBuffer(body)), "id", "c1")
	w := httptest.NewRecorder()
	ctrl.Update(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCourseController_Update_ServiceError(t *testing.T) {
	// Arrange
	mockSvc := new(MockCourseService)
	ctrl := controllers.NewCourseController(mockSvc)

	mockSvc.On("UpdateCourse", mock.Anything, "c1", mock.AnythingOfType("models.CourseInput")).
		Return(nil, errors.New("update failed"))

	body, _ := json.Marshal(models.CourseInput{Title: "Curso"})
	req := withChiURLParam(httptest.NewRequest("PUT", "/api/cursos/c1", bytes.NewBuffer(body)), "id", "c1")
	w := httptest.NewRecorder()

	// Act
	ctrl.Update(w, req)

	// Assert
	assert.Equal(t, http.StatusBadGateway, w.Code)
}

func TestCourseController_Delete_Success(t *testing.T) {
	mockSvc := new(MockCourseService)
	ctrl := controllers.NewCourseController(mockSvc)

	mockSvc.On("DeleteCourse", mock.Anything, "c1").Return(nil)

	req := withChiURLParam(httptest.NewRequest("DELETE", "/api/cursos/c1", nil), "id", "c1")
	w := httptest.NewRecorder()
	ctrl.Delete(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestCourseController_Delete_NotFound(t *testing.T) {
	mockSvc := new(MockCourseService)
	ctrl := controllers.NewCourseController(mockSvc)

	mockSvc.On("DeleteCourse", mock.Anything, "c1").Return(repoErrNotFound())

	req := withChiURLParam(httptest.NewRequest("DELETE", "/api/cursos/c1", nil), "id", "c1")
	w := httptest.NewRecorder()
	ctrl.Delete(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCourseController_Publish_Conflict_Returns409(t *testing.T) {
	mockSvc := new(MockCourseService)
	ctrl := controllers.NewCourseController(mockSvc)

	// Conflicto de versión optimista en Contentful (edición concurrente):
	// writeError lo mapea a 409, ver http_helpers.go.
	mockSvc.On("PublishCourse", mock.Anything, "c1").Return(repoErrConflict())

	req := withChiURLParam(httptest.NewRequest("POST", "/api/cursos/c1/publicar", nil), "id", "c1")
	w := httptest.NewRecorder()
	ctrl.Publish(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestCourseController_Publish_Success(t *testing.T) {
	mockSvc := new(MockCourseService)
	ctrl := controllers.NewCourseController(mockSvc)

	mockSvc.On("PublishCourse", mock.Anything, "c1").Return(nil)

	req := withChiURLParam(httptest.NewRequest("POST", "/api/cursos/c1/publicar", nil), "id", "c1")
	w := httptest.NewRecorder()
	ctrl.Publish(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCourseController_Unpublish_Success(t *testing.T) {
	mockSvc := new(MockCourseService)
	ctrl := controllers.NewCourseController(mockSvc)

	mockSvc.On("UnpublishCourse", mock.Anything, "c1").Return(nil)

	req := withChiURLParam(httptest.NewRequest("POST", "/api/cursos/c1/despublicar", nil), "id", "c1")
	w := httptest.NewRecorder()
	ctrl.Unpublish(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCourseController_ListDrafts_Success(t *testing.T) {
	mockSvc := new(MockCourseService)
	ctrl := controllers.NewCourseController(mockSvc)

	mockSvc.On("ListAllDraftCourses", mock.Anything).Return([]models.Course{{ID: "draft-1"}}, nil)

	req := httptest.NewRequest("GET", "/api/cursos/borradores", nil)
	w := httptest.NewRecorder()
	ctrl.ListDrafts(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCourseController_ListDrafts_ServiceError(t *testing.T) {
	// Arrange
	mockSvc := new(MockCourseService)
	ctrl := controllers.NewCourseController(mockSvc)
	mockSvc.On("ListAllDraftCourses", mock.Anything).Return(nil, errors.New("db error"))

	req := httptest.NewRequest("GET", "/api/cursos/borradores", nil)
	w := httptest.NewRecorder()

	// Act
	ctrl.ListDrafts(w, req)

	// Assert
	assert.Equal(t, http.StatusBadGateway, w.Code)
}

func TestCourseController_Unpublish_ServiceError(t *testing.T) {
	// Arrange
	mockSvc := new(MockCourseService)
	ctrl := controllers.NewCourseController(mockSvc)
	mockSvc.On("UnpublishCourse", mock.Anything, "c1").Return(errors.New("unpublish err"))

	req := withChiURLParam(httptest.NewRequest("POST", "/api/cursos/c1/despublicar", nil), "id", "c1")
	w := httptest.NewRecorder()

	// Act
	ctrl.Unpublish(w, req)

	// Assert
	assert.Equal(t, http.StatusBadGateway, w.Code)
}

