package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"buenvivir/education/internal/controllers"
	"buenvivir/education/internal/models"
	"buenvivir/education/internal/services"
)

type MockTrainingAreaService struct {
	mock.Mock
}

func (m *MockTrainingAreaService) ListTrainingAreas(ctx context.Context) ([]models.TrainingArea, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.TrainingArea), args.Error(1)
}

func (m *MockTrainingAreaService) CreateTrainingArea(ctx context.Context, input models.TrainingAreaInput) (*models.TrainingArea, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TrainingArea), args.Error(1)
}

func (m *MockTrainingAreaService) UpdateTrainingArea(ctx context.Context, id string, input models.TrainingAreaInput) (*models.TrainingArea, error) {
	args := m.Called(ctx, id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TrainingArea), args.Error(1)
}

func (m *MockTrainingAreaService) DeleteTrainingArea(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestTrainingAreaController_List_Success(t *testing.T) {
	mockSvc := new(MockTrainingAreaService)
	ctrl := controllers.NewTrainingAreaController(mockSvc)

	mockSvc.On("ListTrainingAreas", mock.Anything).Return([]models.TrainingArea{{ID: "ta1"}}, nil)

	req := httptest.NewRequest("GET", "/api/areas-formativas", nil)
	w := httptest.NewRecorder()
	ctrl.List(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTrainingAreaController_List_ServiceError(t *testing.T) {
	mockSvc := new(MockTrainingAreaService)
	ctrl := controllers.NewTrainingAreaController(mockSvc)

	mockSvc.On("ListTrainingAreas", mock.Anything).Return(nil, errors.New("contentful caído"))

	req := httptest.NewRequest("GET", "/api/areas-formativas", nil)
	w := httptest.NewRecorder()
	ctrl.List(w, req)

	assert.Equal(t, http.StatusBadGateway, w.Code)
}

func TestTrainingAreaController_Create_BadJSON(t *testing.T) {
	mockSvc := new(MockTrainingAreaService)
	ctrl := controllers.NewTrainingAreaController(mockSvc)

	req := httptest.NewRequest("POST", "/api/areas-formativas", bytes.NewBufferString("{invalid"))
	w := httptest.NewRecorder()
	ctrl.Create(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTrainingAreaController_Create_ValidationError(t *testing.T) {
	mockSvc := new(MockTrainingAreaService)
	ctrl := controllers.NewTrainingAreaController(mockSvc)

	mockSvc.On("CreateTrainingArea", mock.Anything, mock.AnythingOfType("models.TrainingAreaInput")).
		Return(nil, services.NewValidationError("el campo 'name' es requerido"))

	body, _ := json.Marshal(models.TrainingAreaInput{})
	req := httptest.NewRequest("POST", "/api/areas-formativas", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	ctrl.Create(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTrainingAreaController_Create_Success(t *testing.T) {
	mockSvc := new(MockTrainingAreaService)
	ctrl := controllers.NewTrainingAreaController(mockSvc)

	mockSvc.On("CreateTrainingArea", mock.Anything, mock.AnythingOfType("models.TrainingAreaInput")).
		Return(&models.TrainingArea{ID: "ta1"}, nil)

	body, _ := json.Marshal(models.TrainingAreaInput{Name: "Oficios"})
	req := httptest.NewRequest("POST", "/api/areas-formativas", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	ctrl.Create(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestTrainingAreaController_Update_Success(t *testing.T) {
	mockSvc := new(MockTrainingAreaService)
	ctrl := controllers.NewTrainingAreaController(mockSvc)

	mockSvc.On("UpdateTrainingArea", mock.Anything, "ta1", mock.AnythingOfType("models.TrainingAreaInput")).
		Return(&models.TrainingArea{ID: "ta1"}, nil)

	body, _ := json.Marshal(models.TrainingAreaInput{Name: "Oficios"})
	req := withChiURLParam(httptest.NewRequest("PUT", "/api/areas-formativas/ta1", bytes.NewBuffer(body)), "id", "ta1")
	w := httptest.NewRecorder()
	ctrl.Update(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTrainingAreaController_Delete_Success(t *testing.T) {
	mockSvc := new(MockTrainingAreaService)
	ctrl := controllers.NewTrainingAreaController(mockSvc)

	mockSvc.On("DeleteTrainingArea", mock.Anything, "ta1").Return(nil)

	req := withChiURLParam(httptest.NewRequest("DELETE", "/api/areas-formativas/ta1", nil), "id", "ta1")
	w := httptest.NewRecorder()
	ctrl.Delete(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestTrainingAreaController_Update_BadJSON(t *testing.T) {
	// Arrange
	mockSvc := new(MockTrainingAreaService)
	ctrl := controllers.NewTrainingAreaController(mockSvc)

	req := withChiURLParam(httptest.NewRequest("PUT", "/api/areas-formativas/ta1", bytes.NewBufferString("{bad-json")), "id", "ta1")
	w := httptest.NewRecorder()

	// Act
	ctrl.Update(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTrainingAreaController_Update_ServiceError(t *testing.T) {
	// Arrange
	mockSvc := new(MockTrainingAreaService)
	ctrl := controllers.NewTrainingAreaController(mockSvc)
	mockSvc.On("UpdateTrainingArea", mock.Anything, "ta1", mock.Anything).Return(nil, errors.New("update err"))

	body, _ := json.Marshal(models.TrainingAreaInput{Name: "Area"})
	req := withChiURLParam(httptest.NewRequest("PUT", "/api/areas-formativas/ta1", bytes.NewBuffer(body)), "id", "ta1")
	w := httptest.NewRecorder()

	// Act
	ctrl.Update(w, req)

	// Assert
	assert.Equal(t, http.StatusBadGateway, w.Code)
}

func TestTrainingAreaController_Delete_ServiceError(t *testing.T) {
	// Arrange
	mockSvc := new(MockTrainingAreaService)
	ctrl := controllers.NewTrainingAreaController(mockSvc)
	mockSvc.On("DeleteTrainingArea", mock.Anything, "ta1").Return(errors.New("delete err"))

	req := withChiURLParam(httptest.NewRequest("DELETE", "/api/areas-formativas/ta1", nil), "id", "ta1")
	w := httptest.NewRecorder()

	// Act
	ctrl.Delete(w, req)

	// Assert
	assert.Equal(t, http.StatusBadGateway, w.Code)
}

