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

type MockTeacherService struct {
	mock.Mock
}

func (m *MockTeacherService) CreateTeacher(ctx context.Context, req *dto.CreateTeacherRequest) (*dto.TeacherResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TeacherResponse), args.Error(1)
}

func (m *MockTeacherService) GetTeacherByID(ctx context.Context, id string) (*dto.TeacherResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TeacherResponse), args.Error(1)
}

func (m *MockTeacherService) GetAllTeachers(ctx context.Context) ([]dto.TeacherResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dto.TeacherResponse), args.Error(1)
}

func (m *MockTeacherService) UpdateTeacher(ctx context.Context, id string, req *dto.UpdateTeacherRequest) (*dto.TeacherResponse, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TeacherResponse), args.Error(1)
}

func (m *MockTeacherService) DeleteTeacher(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestTeacherController_GetTeacher_Success(t *testing.T) {
	mockSvc := new(MockTeacherService)
	ctrl := controllers.NewTeacherController(mockSvc)

	mockSvc.On("GetTeacherByID", mock.Anything, "teach-1").Return(&dto.TeacherResponse{TeacherID: "teach-1", Specialization: "Math"}, nil)

	req := httptest.NewRequest("GET", "/teachers/teach-1", nil)
	
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "teach-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	ctrl.GetTeacher(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Math")
}

func TestTeacherController_CreateTeacher_BadJSON(t *testing.T) {
	mockSvc := new(MockTeacherService)
	ctrl := controllers.NewTeacherController(mockSvc)

	req := httptest.NewRequest("POST", "/teachers", bytes.NewBufferString("{bad json"))
	w := httptest.NewRecorder()

	ctrl.CreateTeacher(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTeacherController_CreateTeacher_ServiceError(t *testing.T) {
	mockSvc := new(MockTeacherService)
	ctrl := controllers.NewTeacherController(mockSvc)

	mockSvc.On("CreateTeacher", mock.Anything, mock.AnythingOfType("*dto.CreateTeacherRequest")).Return(nil, errors.New("creation failed"))

	reqBody, _ := json.Marshal(dto.CreateTeacherRequest{Specialization: "Math"})
	req := httptest.NewRequest("POST", "/teachers", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	ctrl.CreateTeacher(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestTeacherController_DeleteTeacher_Success(t *testing.T) {
	mockSvc := new(MockTeacherService)
	ctrl := controllers.NewTeacherController(mockSvc)

	mockSvc.On("DeleteTeacher", mock.Anything, "teach-1").Return(nil)

	req := httptest.NewRequest("DELETE", "/teachers/teach-1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "teach-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	ctrl.DeleteTeacher(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTeacherController_CreateTeacher_Success(t *testing.T) {
	mockSvc := new(MockTeacherService)
	ctrl := controllers.NewTeacherController(mockSvc)

	mockSvc.On("CreateTeacher", mock.Anything, mock.AnythingOfType("*dto.CreateTeacherRequest")).Return(&dto.TeacherResponse{TeacherID: "teach-1"}, nil)

	reqBody, _ := json.Marshal(dto.CreateTeacherRequest{Specialization: "Test"})
	req := httptest.NewRequest("POST", "/teachers", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	ctrl.CreateTeacher(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestTeacherController_GetTeacher_ServiceError(t *testing.T) {
	mockSvc := new(MockTeacherService)
	ctrl := controllers.NewTeacherController(mockSvc)

	mockSvc.On("GetTeacherByID", mock.Anything, "teach-1").Return(nil, errors.New("not found"))

	req := httptest.NewRequest("GET", "/teachers/teach-1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "teach-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	ctrl.GetTeacher(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestTeacherController_GetAllTeachers_Success(t *testing.T) {
	mockSvc := new(MockTeacherService)
	ctrl := controllers.NewTeacherController(mockSvc)

	mockSvc.On("GetAllTeachers", mock.Anything).Return([]dto.TeacherResponse{{TeacherID: "teach-1"}}, nil)

	req := httptest.NewRequest("GET", "/teachers", nil)
	w := httptest.NewRecorder()
	ctrl.GetAllTeachers(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTeacherController_GetAllTeachers_ServiceError(t *testing.T) {
	mockSvc := new(MockTeacherService)
	ctrl := controllers.NewTeacherController(mockSvc)

	mockSvc.On("GetAllTeachers", mock.Anything).Return(nil, errors.New("err"))

	req := httptest.NewRequest("GET", "/teachers", nil)
	w := httptest.NewRecorder()
	ctrl.GetAllTeachers(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestTeacherController_UpdateTeacher_Success(t *testing.T) {
	mockSvc := new(MockTeacherService)
	ctrl := controllers.NewTeacherController(mockSvc)

	mockSvc.On("UpdateTeacher", mock.Anything, "teach-1", mock.AnythingOfType("*dto.UpdateTeacherRequest")).Return(&dto.TeacherResponse{TeacherID: "teach-1"}, nil)

	reqBody, _ := json.Marshal(dto.UpdateTeacherRequest{Specialization: "Test"})
	req := httptest.NewRequest("PUT", "/teachers/teach-1", bytes.NewBuffer(reqBody))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "teach-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	ctrl.UpdateTeacher(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTeacherController_UpdateTeacher_BadJSON(t *testing.T) {
	mockSvc := new(MockTeacherService)
	ctrl := controllers.NewTeacherController(mockSvc)

	req := httptest.NewRequest("PUT", "/teachers/teach-1", bytes.NewBufferString("{invalid"))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "teach-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	ctrl.UpdateTeacher(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTeacherController_UpdateTeacher_ServiceError(t *testing.T) {
	mockSvc := new(MockTeacherService)
	ctrl := controllers.NewTeacherController(mockSvc)

	mockSvc.On("UpdateTeacher", mock.Anything, "teach-1", mock.AnythingOfType("*dto.UpdateTeacherRequest")).Return(nil, errors.New("err"))

	reqBody, _ := json.Marshal(dto.UpdateTeacherRequest{Specialization: "Test"})
	req := httptest.NewRequest("PUT", "/teachers/teach-1", bytes.NewBuffer(reqBody))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "teach-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	ctrl.UpdateTeacher(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestTeacherController_DeleteTeacher_ServiceError(t *testing.T) {
	mockSvc := new(MockTeacherService)
	ctrl := controllers.NewTeacherController(mockSvc)

	mockSvc.On("DeleteTeacher", mock.Anything, "teach-1").Return(errors.New("err"))

	req := httptest.NewRequest("DELETE", "/teachers/teach-1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "teach-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	ctrl.DeleteTeacher(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
