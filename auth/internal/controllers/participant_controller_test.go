package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"auth/internal/controllers"
	"auth/internal/dto"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockParticipantService struct {
	mock.Mock
}

func (m *MockParticipantService) CreateParticipant(ctx context.Context, req *dto.CreateParticipantRequest) (*dto.ParticipantResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ParticipantResponse), args.Error(1)
}

func (m *MockParticipantService) GetParticipantByID(ctx context.Context, id string) (*dto.ParticipantResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ParticipantResponse), args.Error(1)
}

func (m *MockParticipantService) GetAllParticipants(ctx context.Context) ([]dto.ParticipantResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dto.ParticipantResponse), args.Error(1)
}

func (m *MockParticipantService) UpdateParticipant(ctx context.Context, id string, req *dto.UpdateParticipantRequest) (*dto.ParticipantResponse, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ParticipantResponse), args.Error(1)
}

func (m *MockParticipantService) DeleteParticipant(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestParticipantController_GetParticipant_Success(t *testing.T) {
	mockSvc := new(MockParticipantService)
	ctrl := controllers.NewParticipantController(mockSvc)

	mockSvc.On("GetParticipantByID", mock.Anything, "part-1").Return(&dto.ParticipantResponse{ParticipantID: "part-1", Address: "Tech School"}, nil)

	req := httptest.NewRequest("GET", "/participants/part-1", nil)
	
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "part-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	ctrl.GetParticipant(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Tech School")
}

func TestParticipantController_CreateParticipant_BadJSON(t *testing.T) {
	mockSvc := new(MockParticipantService)
	ctrl := controllers.NewParticipantController(mockSvc)

	req := httptest.NewRequest("POST", "/participants", bytes.NewBufferString("{bad json"))
	w := httptest.NewRecorder()

	ctrl.CreateParticipant(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestParticipantController_CreateParticipant_ServiceError(t *testing.T) {
	mockSvc := new(MockParticipantService)
	ctrl := controllers.NewParticipantController(mockSvc)

	mockSvc.On("CreateParticipant", mock.Anything, mock.AnythingOfType("*dto.CreateParticipantRequest")).Return(nil, errors.New("creation failed"))

	reqBody, _ := json.Marshal(dto.CreateParticipantRequest{Address: "Test"})
	req := httptest.NewRequest("POST", "/participants", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	ctrl.CreateParticipant(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestParticipantController_DeleteParticipant_Success(t *testing.T) {
	mockSvc := new(MockParticipantService)
	ctrl := controllers.NewParticipantController(mockSvc)

	mockSvc.On("DeleteParticipant", mock.Anything, "part-1").Return(nil)

	req := httptest.NewRequest("DELETE", "/participants/part-1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "part-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	ctrl.DeleteParticipant(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestParticipantController_CreateParticipant_Success(t *testing.T) {
	mockSvc := new(MockParticipantService)
	ctrl := controllers.NewParticipantController(mockSvc)

	mockSvc.On("CreateParticipant", mock.Anything, mock.AnythingOfType("*dto.CreateParticipantRequest")).Return(&dto.ParticipantResponse{ParticipantID: "part-1"}, nil)

	reqBody, _ := json.Marshal(dto.CreateParticipantRequest{Address: "Test"})
	req := httptest.NewRequest("POST", "/participants", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	ctrl.CreateParticipant(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestParticipantController_GetParticipant_ServiceError(t *testing.T) {
	mockSvc := new(MockParticipantService)
	ctrl := controllers.NewParticipantController(mockSvc)

	mockSvc.On("GetParticipantByID", mock.Anything, "part-1").Return(nil, errors.New("not found"))

	req := httptest.NewRequest("GET", "/participants/part-1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "part-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	ctrl.GetParticipant(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestParticipantController_GetAllParticipants_Success(t *testing.T) {
	mockSvc := new(MockParticipantService)
	ctrl := controllers.NewParticipantController(mockSvc)

	mockSvc.On("GetAllParticipants", mock.Anything).Return([]dto.ParticipantResponse{{ParticipantID: "part-1"}}, nil)

	req := httptest.NewRequest("GET", "/participants", nil)
	w := httptest.NewRecorder()
	ctrl.GetAllParticipants(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestParticipantController_GetAllParticipants_ServiceError(t *testing.T) {
	mockSvc := new(MockParticipantService)
	ctrl := controllers.NewParticipantController(mockSvc)

	mockSvc.On("GetAllParticipants", mock.Anything).Return(nil, errors.New("err"))

	req := httptest.NewRequest("GET", "/participants", nil)
	w := httptest.NewRecorder()
	ctrl.GetAllParticipants(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestParticipantController_UpdateParticipant_Success(t *testing.T) {
	mockSvc := new(MockParticipantService)
	ctrl := controllers.NewParticipantController(mockSvc)

	mockSvc.On("UpdateParticipant", mock.Anything, "part-1", mock.AnythingOfType("*dto.UpdateParticipantRequest")).Return(&dto.ParticipantResponse{ParticipantID: "part-1"}, nil)

	reqBody, _ := json.Marshal(dto.UpdateParticipantRequest{Address: "Test"})
	req := httptest.NewRequest("PUT", "/participants/part-1", bytes.NewBuffer(reqBody))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "part-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	ctrl.UpdateParticipant(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestParticipantController_UpdateParticipant_BadJSON(t *testing.T) {
	mockSvc := new(MockParticipantService)
	ctrl := controllers.NewParticipantController(mockSvc)

	req := httptest.NewRequest("PUT", "/participants/part-1", bytes.NewBufferString("{invalid"))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "part-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	ctrl.UpdateParticipant(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestParticipantController_UpdateParticipant_ServiceError(t *testing.T) {
	mockSvc := new(MockParticipantService)
	ctrl := controllers.NewParticipantController(mockSvc)

	mockSvc.On("UpdateParticipant", mock.Anything, "part-1", mock.AnythingOfType("*dto.UpdateParticipantRequest")).Return(nil, errors.New("err"))

	reqBody, _ := json.Marshal(dto.UpdateParticipantRequest{Address: "Test"})
	req := httptest.NewRequest("PUT", "/participants/part-1", bytes.NewBuffer(reqBody))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "part-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	ctrl.UpdateParticipant(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestParticipantController_DeleteParticipant_ServiceError(t *testing.T) {
	mockSvc := new(MockParticipantService)
	ctrl := controllers.NewParticipantController(mockSvc)

	mockSvc.On("DeleteParticipant", mock.Anything, "part-1").Return(errors.New("err"))

	req := httptest.NewRequest("DELETE", "/participants/part-1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "part-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	ctrl.DeleteParticipant(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
