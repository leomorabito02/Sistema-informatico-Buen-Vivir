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

type MockPermissionService struct {
	mock.Mock
}

func (m *MockPermissionService) CreatePermission(ctx context.Context, req *dto.CreatePermissionRequest) (*dto.PermissionResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PermissionResponse), args.Error(1)
}

func (m *MockPermissionService) GetPermissionByID(ctx context.Context, id string) (*dto.PermissionResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PermissionResponse), args.Error(1)
}

func (m *MockPermissionService) GetAllPermissions(ctx context.Context) ([]dto.PermissionResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dto.PermissionResponse), args.Error(1)
}

func (m *MockPermissionService) UpdatePermission(ctx context.Context, id string, req *dto.UpdatePermissionRequest) (*dto.PermissionResponse, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PermissionResponse), args.Error(1)
}

func (m *MockPermissionService) DeletePermission(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPermissionService) CheckPermission(ctx context.Context, roleID, action, resource string) (bool, error) {
	args := m.Called(ctx, roleID, action, resource)
	return args.Bool(0), args.Error(1)
}

func TestPermissionController_GetPermission_Success(t *testing.T) {
	mockSvc := new(MockPermissionService)
	ctrl := controllers.NewPermissionController(mockSvc)

	mockSvc.On("GetPermissionByID", mock.Anything, "perm-1").Return(&dto.PermissionResponse{PermissionID: "perm-1", Action: "create"}, nil)

	req := httptest.NewRequest("GET", "/permissions/perm-1", nil)
	
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "perm-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	ctrl.GetPermission(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "create")
}

func TestPermissionController_CreatePermission_BadJSON(t *testing.T) {
	mockSvc := new(MockPermissionService)
	ctrl := controllers.NewPermissionController(mockSvc)

	req := httptest.NewRequest("POST", "/permissions", bytes.NewBufferString("{bad json"))
	w := httptest.NewRecorder()

	ctrl.CreatePermission(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPermissionController_CreatePermission_ServiceError(t *testing.T) {
	mockSvc := new(MockPermissionService)
	ctrl := controllers.NewPermissionController(mockSvc)

	mockSvc.On("CreatePermission", mock.Anything, mock.AnythingOfType("*dto.CreatePermissionRequest")).Return(nil, errors.New("creation failed"))

	reqBody, _ := json.Marshal(dto.CreatePermissionRequest{Action: "create"})
	req := httptest.NewRequest("POST", "/permissions", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	ctrl.CreatePermission(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestPermissionController_DeletePermission_Success(t *testing.T) {
	mockSvc := new(MockPermissionService)
	ctrl := controllers.NewPermissionController(mockSvc)

	mockSvc.On("DeletePermission", mock.Anything, "perm-1").Return(nil)

	req := httptest.NewRequest("DELETE", "/permissions/perm-1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "perm-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	ctrl.DeletePermission(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPermissionController_CreatePermission_Success(t *testing.T) {
	mockSvc := new(MockPermissionService)
	ctrl := controllers.NewPermissionController(mockSvc)

	mockSvc.On("CreatePermission", mock.Anything, mock.AnythingOfType("*dto.CreatePermissionRequest")).Return(&dto.PermissionResponse{PermissionID: "perm-1"}, nil)

	reqBody, _ := json.Marshal(dto.CreatePermissionRequest{Action: "read", Resource: "users"})
	req := httptest.NewRequest("POST", "/permissions", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	ctrl.CreatePermission(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestPermissionController_GetPermission_ServiceError(t *testing.T) {
	mockSvc := new(MockPermissionService)
	ctrl := controllers.NewPermissionController(mockSvc)

	mockSvc.On("GetPermissionByID", mock.Anything, "perm-1").Return(nil, errors.New("not found"))

	req := httptest.NewRequest("GET", "/permissions/perm-1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "perm-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	ctrl.GetPermission(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestPermissionController_GetAllPermissions_Success(t *testing.T) {
	mockSvc := new(MockPermissionService)
	ctrl := controllers.NewPermissionController(mockSvc)

	mockSvc.On("GetAllPermissions", mock.Anything).Return([]dto.PermissionResponse{{PermissionID: "perm-1"}}, nil)

	req := httptest.NewRequest("GET", "/permissions", nil)
	w := httptest.NewRecorder()
	ctrl.GetAllPermissions(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPermissionController_GetAllPermissions_ServiceError(t *testing.T) {
	mockSvc := new(MockPermissionService)
	ctrl := controllers.NewPermissionController(mockSvc)

	mockSvc.On("GetAllPermissions", mock.Anything).Return(nil, errors.New("err"))

	req := httptest.NewRequest("GET", "/permissions", nil)
	w := httptest.NewRecorder()
	ctrl.GetAllPermissions(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestPermissionController_UpdatePermission_Success(t *testing.T) {
	mockSvc := new(MockPermissionService)
	ctrl := controllers.NewPermissionController(mockSvc)

	mockSvc.On("UpdatePermission", mock.Anything, "perm-1", mock.AnythingOfType("*dto.UpdatePermissionRequest")).Return(&dto.PermissionResponse{PermissionID: "perm-1"}, nil)

	reqBody, _ := json.Marshal(dto.UpdatePermissionRequest{Action: "write"})
	req := httptest.NewRequest("PUT", "/permissions/perm-1", bytes.NewBuffer(reqBody))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "perm-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	ctrl.UpdatePermission(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPermissionController_UpdatePermission_BadJSON(t *testing.T) {
	mockSvc := new(MockPermissionService)
	ctrl := controllers.NewPermissionController(mockSvc)

	req := httptest.NewRequest("PUT", "/permissions/perm-1", bytes.NewBufferString("{invalid"))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "perm-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	ctrl.UpdatePermission(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPermissionController_UpdatePermission_ServiceError(t *testing.T) {
	mockSvc := new(MockPermissionService)
	ctrl := controllers.NewPermissionController(mockSvc)

	mockSvc.On("UpdatePermission", mock.Anything, "perm-1", mock.AnythingOfType("*dto.UpdatePermissionRequest")).Return(nil, errors.New("err"))

	reqBody, _ := json.Marshal(dto.UpdatePermissionRequest{Action: "write"})
	req := httptest.NewRequest("PUT", "/permissions/perm-1", bytes.NewBuffer(reqBody))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "perm-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	ctrl.UpdatePermission(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestPermissionController_DeletePermission_ServiceError(t *testing.T) {
	mockSvc := new(MockPermissionService)
	ctrl := controllers.NewPermissionController(mockSvc)

	mockSvc.On("DeletePermission", mock.Anything, "perm-1").Return(errors.New("err"))

	req := httptest.NewRequest("DELETE", "/permissions/perm-1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "perm-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	ctrl.DeletePermission(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
