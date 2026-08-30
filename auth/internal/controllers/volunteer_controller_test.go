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

type MockVolunteerService struct {
	mock.Mock
}

func (m *MockVolunteerService) CreateVolunteer(ctx context.Context, req *dto.CreateVolunteerRequest) (*dto.VolunteerResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.VolunteerResponse), args.Error(1)
}

func (m *MockVolunteerService) GetVolunteerByID(ctx context.Context, id string) (*dto.VolunteerResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.VolunteerResponse), args.Error(1)
}

func (m *MockVolunteerService) GetAllVolunteers(ctx context.Context) ([]dto.VolunteerResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dto.VolunteerResponse), args.Error(1)
}

func (m *MockVolunteerService) UpdateVolunteer(ctx context.Context, id string, req *dto.UpdateVolunteerRequest) (*dto.VolunteerResponse, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.VolunteerResponse), args.Error(1)
}

func (m *MockVolunteerService) DeleteVolunteer(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestVolunteerController_GetVolunteer_Success(t *testing.T) {
	mockSvc := new(MockVolunteerService)
	ctrl := controllers.NewVolunteerController(mockSvc)

	mockSvc.On("GetVolunteerByID", mock.Anything, "vol-1").Return(&dto.VolunteerResponse{VolunteerID: "vol-1", ApprovalStatus: "pending"}, nil)

	req := httptest.NewRequest("GET", "/volunteers/vol-1", nil)
	
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "vol-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	ctrl.GetVolunteer(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "pending")
}

func TestVolunteerController_CreateVolunteer_BadJSON(t *testing.T) {
	mockSvc := new(MockVolunteerService)
	ctrl := controllers.NewVolunteerController(mockSvc)

	req := httptest.NewRequest("POST", "/volunteers", bytes.NewBufferString("{bad json"))
	w := httptest.NewRecorder()

	ctrl.CreateVolunteer(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestVolunteerController_CreateVolunteer_ServiceError(t *testing.T) {
	mockSvc := new(MockVolunteerService)
	ctrl := controllers.NewVolunteerController(mockSvc)

	mockSvc.On("CreateVolunteer", mock.Anything, mock.AnythingOfType("*dto.CreateVolunteerRequest")).Return(nil, errors.New("creation failed"))

	reqBody, _ := json.Marshal(dto.CreateVolunteerRequest{})
	req := httptest.NewRequest("POST", "/volunteers", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	ctrl.CreateVolunteer(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestVolunteerController_DeleteVolunteer_Success(t *testing.T) {
	mockSvc := new(MockVolunteerService)
	ctrl := controllers.NewVolunteerController(mockSvc)

	mockSvc.On("DeleteVolunteer", mock.Anything, "vol-1").Return(nil)

	req := httptest.NewRequest("DELETE", "/volunteers/vol-1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "vol-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	ctrl.DeleteVolunteer(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestVolunteerController_CreateVolunteer_Success(t *testing.T) {
	mockSvc := new(MockVolunteerService)
	ctrl := controllers.NewVolunteerController(mockSvc)

	mockSvc.On("CreateVolunteer", mock.Anything, mock.AnythingOfType("*dto.CreateVolunteerRequest")).Return(&dto.VolunteerResponse{VolunteerID: "vol-1"}, nil)

	reqBody, _ := json.Marshal(dto.CreateVolunteerRequest{ApprovalStatus: "pending"})
	req := httptest.NewRequest("POST", "/volunteers", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	ctrl.CreateVolunteer(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestVolunteerController_GetVolunteer_ServiceError(t *testing.T) {
	mockSvc := new(MockVolunteerService)
	ctrl := controllers.NewVolunteerController(mockSvc)

	mockSvc.On("GetVolunteerByID", mock.Anything, "vol-1").Return(nil, errors.New("not found"))

	req := httptest.NewRequest("GET", "/volunteers/vol-1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "vol-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	ctrl.GetVolunteer(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestVolunteerController_GetAllVolunteers_Success(t *testing.T) {
	mockSvc := new(MockVolunteerService)
	ctrl := controllers.NewVolunteerController(mockSvc)

	mockSvc.On("GetAllVolunteers", mock.Anything).Return([]dto.VolunteerResponse{{VolunteerID: "vol-1"}}, nil)

	req := httptest.NewRequest("GET", "/volunteers", nil)
	w := httptest.NewRecorder()
	ctrl.GetAllVolunteers(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestVolunteerController_GetAllVolunteers_ServiceError(t *testing.T) {
	mockSvc := new(MockVolunteerService)
	ctrl := controllers.NewVolunteerController(mockSvc)

	mockSvc.On("GetAllVolunteers", mock.Anything).Return(nil, errors.New("err"))

	req := httptest.NewRequest("GET", "/volunteers", nil)
	w := httptest.NewRecorder()
	ctrl.GetAllVolunteers(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestVolunteerController_UpdateVolunteer_Success(t *testing.T) {
	mockSvc := new(MockVolunteerService)
	ctrl := controllers.NewVolunteerController(mockSvc)

	mockSvc.On("UpdateVolunteer", mock.Anything, "vol-1", mock.AnythingOfType("*dto.UpdateVolunteerRequest")).Return(&dto.VolunteerResponse{VolunteerID: "vol-1"}, nil)

	reqBody, _ := json.Marshal(dto.UpdateVolunteerRequest{ApprovalStatus: "approved"})
	req := httptest.NewRequest("PUT", "/volunteers/vol-1", bytes.NewBuffer(reqBody))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "vol-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	ctrl.UpdateVolunteer(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestVolunteerController_UpdateVolunteer_BadJSON(t *testing.T) {
	mockSvc := new(MockVolunteerService)
	ctrl := controllers.NewVolunteerController(mockSvc)

	req := httptest.NewRequest("PUT", "/volunteers/vol-1", bytes.NewBufferString("{invalid"))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "vol-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	ctrl.UpdateVolunteer(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestVolunteerController_UpdateVolunteer_ServiceError(t *testing.T) {
	mockSvc := new(MockVolunteerService)
	ctrl := controllers.NewVolunteerController(mockSvc)

	mockSvc.On("UpdateVolunteer", mock.Anything, "vol-1", mock.AnythingOfType("*dto.UpdateVolunteerRequest")).Return(nil, errors.New("err"))

	reqBody, _ := json.Marshal(dto.UpdateVolunteerRequest{ApprovalStatus: "approved"})
	req := httptest.NewRequest("PUT", "/volunteers/vol-1", bytes.NewBuffer(reqBody))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "vol-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	ctrl.UpdateVolunteer(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestVolunteerController_DeleteVolunteer_ServiceError(t *testing.T) {
	mockSvc := new(MockVolunteerService)
	ctrl := controllers.NewVolunteerController(mockSvc)

	mockSvc.On("DeleteVolunteer", mock.Anything, "vol-1").Return(errors.New("err"))

	req := httptest.NewRequest("DELETE", "/volunteers/vol-1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "vol-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	ctrl.DeleteVolunteer(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
