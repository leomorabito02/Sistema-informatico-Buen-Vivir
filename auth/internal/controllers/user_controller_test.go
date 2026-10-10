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

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.UserResponse), args.Error(1)
}

func (m *MockUserService) GetUserByID(ctx context.Context, id string) (*dto.UserResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.UserResponse), args.Error(1)
}

func (m *MockUserService) GetParticipantsHistory(ctx context.Context) ([]dto.ParticipantHistoryResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dto.ParticipantHistoryResponse), args.Error(1)
}

func (m *MockUserService) GetAllUsers(ctx context.Context) ([]dto.UserResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dto.UserResponse), args.Error(1)
}

func (m *MockUserService) UpdateUser(ctx context.Context, id string, req *dto.UpdateUserRequest) (*dto.UserResponse, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.UserResponse), args.Error(1)
}

func (m *MockUserService) DeleteUser(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestUserController_GetUser_Success(t *testing.T) {
	mockSvc := new(MockUserService)
	ctrl := controllers.NewUserController(mockSvc)

	mockSvc.On("GetUserByID", mock.Anything, "usr-1").Return(&dto.UserResponse{UserID: "usr-1", FirstName: "John"}, nil)

	req := httptest.NewRequest("GET", "/users/usr-1", nil)
	
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "usr-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	ctrl.GetUser(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "John")
}

func TestUserController_CreateUser_BadJSON(t *testing.T) {
	mockSvc := new(MockUserService)
	ctrl := controllers.NewUserController(mockSvc)

	req := httptest.NewRequest("POST", "/users", bytes.NewBufferString("{bad json"))
	w := httptest.NewRecorder()

	ctrl.CreateUser(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserController_CreateUser_ServiceError(t *testing.T) {
	mockSvc := new(MockUserService)
	ctrl := controllers.NewUserController(mockSvc)

	mockSvc.On("CreateUser", mock.Anything, mock.AnythingOfType("*dto.CreateUserRequest")).Return(nil, errors.New("creation failed"))

	reqBody, _ := json.Marshal(dto.CreateUserRequest{FirstName: "Test"})
	req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	ctrl.CreateUser(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUserController_DeleteUser_Success(t *testing.T) {
	mockSvc := new(MockUserService)
	ctrl := controllers.NewUserController(mockSvc)

	mockSvc.On("DeleteUser", mock.Anything, "usr-1").Return(nil)

	req := httptest.NewRequest("DELETE", "/users/usr-1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "usr-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	ctrl.DeleteUser(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUserController_CreateUser_Success(t *testing.T) {
	mockSvc := new(MockUserService)
	ctrl := controllers.NewUserController(mockSvc)

	mockSvc.On("CreateUser", mock.Anything, mock.AnythingOfType("*dto.CreateUserRequest")).Return(&dto.UserResponse{UserID: "usr-1"}, nil)

	reqBody, _ := json.Marshal(dto.CreateUserRequest{FirstName: "Test"})
	req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	ctrl.CreateUser(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestUserController_GetUser_ServiceError(t *testing.T) {
	mockSvc := new(MockUserService)
	ctrl := controllers.NewUserController(mockSvc)

	mockSvc.On("GetUserByID", mock.Anything, "usr-1").Return(nil, errors.New("not found"))

	req := httptest.NewRequest("GET", "/users/usr-1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "usr-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	ctrl.GetUser(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUserController_GetAllUsers_Success(t *testing.T) {
	mockSvc := new(MockUserService)
	ctrl := controllers.NewUserController(mockSvc)

	mockSvc.On("GetAllUsers", mock.Anything).Return([]dto.UserResponse{{UserID: "usr-1"}}, nil)

	req := httptest.NewRequest("GET", "/users", nil)
	w := httptest.NewRecorder()
	ctrl.GetAllUsers(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUserController_GetAllUsers_ServiceError(t *testing.T) {
	mockSvc := new(MockUserService)
	ctrl := controllers.NewUserController(mockSvc)

	mockSvc.On("GetAllUsers", mock.Anything).Return(nil, errors.New("err"))

	req := httptest.NewRequest("GET", "/users", nil)
	w := httptest.NewRecorder()
	ctrl.GetAllUsers(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUserController_UpdateUser_Success(t *testing.T) {
	mockSvc := new(MockUserService)
	ctrl := controllers.NewUserController(mockSvc)

	mockSvc.On("UpdateUser", mock.Anything, "usr-1", mock.AnythingOfType("*dto.UpdateUserRequest")).Return(&dto.UserResponse{UserID: "usr-1"}, nil)

	reqBody, _ := json.Marshal(dto.UpdateUserRequest{FirstName: "Test"})
	req := httptest.NewRequest("PUT", "/users/usr-1", bytes.NewBuffer(reqBody))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "usr-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	ctrl.UpdateUser(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUserController_UpdateUser_BadJSON(t *testing.T) {
	mockSvc := new(MockUserService)
	ctrl := controllers.NewUserController(mockSvc)

	req := httptest.NewRequest("PUT", "/users/usr-1", bytes.NewBufferString("{invalid"))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "usr-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	ctrl.UpdateUser(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserController_UpdateUser_ServiceError(t *testing.T) {
	mockSvc := new(MockUserService)
	ctrl := controllers.NewUserController(mockSvc)

	mockSvc.On("UpdateUser", mock.Anything, "usr-1", mock.AnythingOfType("*dto.UpdateUserRequest")).Return(nil, errors.New("err"))

	reqBody, _ := json.Marshal(dto.UpdateUserRequest{FirstName: "Test"})
	req := httptest.NewRequest("PUT", "/users/usr-1", bytes.NewBuffer(reqBody))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "usr-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	ctrl.UpdateUser(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUserController_DeleteUser_ServiceError(t *testing.T) {
	mockSvc := new(MockUserService)
	ctrl := controllers.NewUserController(mockSvc)

	mockSvc.On("DeleteUser", mock.Anything, "usr-1").Return(errors.New("err"))

	req := httptest.NewRequest("DELETE", "/users/usr-1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "usr-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	ctrl.DeleteUser(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUserController_GetParticipantsHistory(t *testing.T) {
	cases := []struct {
		name   string
		ret    []dto.ParticipantHistoryResponse
		err    error
		status int
	}{
		{"ok", []dto.ParticipantHistoryResponse{{UserID: "usr-1", Status: "active"}}, nil, http.StatusOK},
		{"forbidden", nil, errors.New("forbidden: only administrators"), http.StatusForbidden},
		{"db", nil, errors.New("db error"), http.StatusInternalServerError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockSvc := new(MockUserService)
			ctrl := controllers.NewUserController(mockSvc)
			mockSvc.On("GetParticipantsHistory", mock.Anything).Return(tc.ret, tc.err)

			w := httptest.NewRecorder()
			ctrl.GetParticipantsHistory(w, httptest.NewRequest("GET", "/users/participants/history", nil))
			assert.Equal(t, tc.status, w.Code)
			if tc.status == http.StatusOK {
				assert.Contains(t, w.Body.String(), `"user_id":"usr-1"`)
			}
		})
	}
}
