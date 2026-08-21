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

type MockRoleService struct {
	mock.Mock
}

func (m *MockRoleService) CreateRole(ctx context.Context, req *dto.CreateRoleRequest) (*dto.RoleResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.RoleResponse), args.Error(1)
}

func (m *MockRoleService) GetRoleByID(ctx context.Context, id string) (*dto.RoleResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.RoleResponse), args.Error(1)
}

func (m *MockRoleService) GetAllRoles(ctx context.Context) ([]dto.RoleResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dto.RoleResponse), args.Error(1)
}

func (m *MockRoleService) UpdateRole(ctx context.Context, id string, req *dto.UpdateRoleRequest) (*dto.RoleResponse, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.RoleResponse), args.Error(1)
}

func (m *MockRoleService) DeleteRole(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestRoleController_GetRole_Success(t *testing.T) {
	mockSvc := new(MockRoleService)
	ctrl := controllers.NewRoleController(mockSvc)

	mockSvc.On("GetRoleByID", mock.Anything, "role-1").Return(&dto.RoleResponse{RoleID: "role-1", Name: "Admin"}, nil)

	req := httptest.NewRequest("GET", "/roles/role-1", nil)
	
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "role-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	ctrl.GetRole(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Admin")
}

func TestRoleController_CreateRole_BadJSON(t *testing.T) {
	mockSvc := new(MockRoleService)
	ctrl := controllers.NewRoleController(mockSvc)

	req := httptest.NewRequest("POST", "/roles", bytes.NewBufferString("{bad json"))
	w := httptest.NewRecorder()

	ctrl.CreateRole(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRoleController_CreateRole_ServiceError(t *testing.T) {
	mockSvc := new(MockRoleService)
	ctrl := controllers.NewRoleController(mockSvc)

	mockSvc.On("CreateRole", mock.Anything, mock.AnythingOfType("*dto.CreateRoleRequest")).Return(nil, errors.New("creation failed"))

	reqBody, _ := json.Marshal(dto.CreateRoleRequest{Name: "Test"})
	req := httptest.NewRequest("POST", "/roles", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	ctrl.CreateRole(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRoleController_DeleteRole_Success(t *testing.T) {
	mockSvc := new(MockRoleService)
	ctrl := controllers.NewRoleController(mockSvc)

	mockSvc.On("DeleteRole", mock.Anything, "role-1").Return(nil)

	req := httptest.NewRequest("DELETE", "/roles/role-1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "role-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	ctrl.DeleteRole(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRoleController_CreateRole_Success(t *testing.T) {
	mockSvc := new(MockRoleService)
	ctrl := controllers.NewRoleController(mockSvc)

	mockSvc.On("CreateRole", mock.Anything, mock.AnythingOfType("*dto.CreateRoleRequest")).Return(&dto.RoleResponse{RoleID: "role-1"}, nil)

	reqBody, _ := json.Marshal(dto.CreateRoleRequest{Name: "Admin"})
	req := httptest.NewRequest("POST", "/roles", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	ctrl.CreateRole(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestRoleController_GetRole_ServiceError(t *testing.T) {
	mockSvc := new(MockRoleService)
	ctrl := controllers.NewRoleController(mockSvc)

	mockSvc.On("GetRoleByID", mock.Anything, "role-1").Return(nil, errors.New("not found"))

	req := httptest.NewRequest("GET", "/roles/role-1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "role-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	ctrl.GetRole(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRoleController_GetAllRoles_Success(t *testing.T) {
	mockSvc := new(MockRoleService)
	ctrl := controllers.NewRoleController(mockSvc)

	mockSvc.On("GetAllRoles", mock.Anything).Return([]dto.RoleResponse{{RoleID: "role-1"}}, nil)

	req := httptest.NewRequest("GET", "/roles", nil)
	w := httptest.NewRecorder()
	ctrl.GetAllRoles(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRoleController_GetAllRoles_ServiceError(t *testing.T) {
	mockSvc := new(MockRoleService)
	ctrl := controllers.NewRoleController(mockSvc)

	mockSvc.On("GetAllRoles", mock.Anything).Return(nil, errors.New("err"))

	req := httptest.NewRequest("GET", "/roles", nil)
	w := httptest.NewRecorder()
	ctrl.GetAllRoles(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRoleController_UpdateRole_Success(t *testing.T) {
	mockSvc := new(MockRoleService)
	ctrl := controllers.NewRoleController(mockSvc)

	mockSvc.On("UpdateRole", mock.Anything, "role-1", mock.AnythingOfType("*dto.UpdateRoleRequest")).Return(&dto.RoleResponse{RoleID: "role-1"}, nil)

	reqBody, _ := json.Marshal(dto.UpdateRoleRequest{Name: "Admin"})
	req := httptest.NewRequest("PUT", "/roles/role-1", bytes.NewBuffer(reqBody))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "role-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	ctrl.UpdateRole(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRoleController_UpdateRole_BadJSON(t *testing.T) {
	mockSvc := new(MockRoleService)
	ctrl := controllers.NewRoleController(mockSvc)

	req := httptest.NewRequest("PUT", "/roles/role-1", bytes.NewBufferString("{invalid"))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "role-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	ctrl.UpdateRole(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRoleController_UpdateRole_ServiceError(t *testing.T) {
	mockSvc := new(MockRoleService)
	ctrl := controllers.NewRoleController(mockSvc)

	mockSvc.On("UpdateRole", mock.Anything, "role-1", mock.AnythingOfType("*dto.UpdateRoleRequest")).Return(nil, errors.New("err"))

	reqBody, _ := json.Marshal(dto.UpdateRoleRequest{Name: "Admin"})
	req := httptest.NewRequest("PUT", "/roles/role-1", bytes.NewBuffer(reqBody))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "role-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	ctrl.UpdateRole(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRoleController_DeleteRole_ServiceError(t *testing.T) {
	mockSvc := new(MockRoleService)
	ctrl := controllers.NewRoleController(mockSvc)

	mockSvc.On("DeleteRole", mock.Anything, "role-1").Return(errors.New("err"))

	req := httptest.NewRequest("DELETE", "/roles/role-1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "role-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	ctrl.DeleteRole(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
