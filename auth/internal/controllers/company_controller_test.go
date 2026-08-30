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

type MockCompanyService struct {
	mock.Mock
}

func (m *MockCompanyService) CreateCompany(ctx context.Context, req *dto.CreateCompanyRequest) (*dto.CompanyResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.CompanyResponse), args.Error(1)
}

func (m *MockCompanyService) GetCompanyByID(ctx context.Context, id string) (*dto.CompanyResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.CompanyResponse), args.Error(1)
}

func (m *MockCompanyService) GetAllCompanies(ctx context.Context) ([]dto.CompanyResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dto.CompanyResponse), args.Error(1)
}

func (m *MockCompanyService) UpdateCompany(ctx context.Context, id string, req *dto.UpdateCompanyRequest) (*dto.CompanyResponse, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.CompanyResponse), args.Error(1)
}

func (m *MockCompanyService) DeleteCompany(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestCompanyController_GetCompany_Success(t *testing.T) {
	mockSvc := new(MockCompanyService)
	ctrl := controllers.NewCompanyController(mockSvc)

	mockSvc.On("GetCompanyByID", mock.Anything, "comp-1").Return(&dto.CompanyResponse{CompanyID: "comp-1", BusinessName: "Test Co"}, nil)

	req := httptest.NewRequest("GET", "/companies/comp-1", nil)
	
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "comp-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	ctrl.GetCompany(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Test Co")
}

func TestCompanyController_CreateCompany_BadJSON(t *testing.T) {
	mockSvc := new(MockCompanyService)
	ctrl := controllers.NewCompanyController(mockSvc)

	req := httptest.NewRequest("POST", "/companies", bytes.NewBufferString("{bad json"))
	w := httptest.NewRecorder()

	ctrl.CreateCompany(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCompanyController_CreateCompany_ServiceError(t *testing.T) {
	mockSvc := new(MockCompanyService)
	ctrl := controllers.NewCompanyController(mockSvc)

	mockSvc.On("CreateCompany", mock.Anything, mock.AnythingOfType("*dto.CreateCompanyRequest")).Return(nil, errors.New("creation failed"))

	reqBody, _ := json.Marshal(dto.CreateCompanyRequest{BusinessName: "Test Co"})
	req := httptest.NewRequest("POST", "/companies", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	ctrl.CreateCompany(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCompanyController_DeleteCompany_Success(t *testing.T) {
	mockSvc := new(MockCompanyService)
	ctrl := controllers.NewCompanyController(mockSvc)

	mockSvc.On("DeleteCompany", mock.Anything, "comp-1").Return(nil)

	req := httptest.NewRequest("DELETE", "/companies/comp-1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "comp-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	ctrl.DeleteCompany(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCompanyController_CreateCompany_Success(t *testing.T) {
	mockSvc := new(MockCompanyService)
	ctrl := controllers.NewCompanyController(mockSvc)

	mockSvc.On("CreateCompany", mock.Anything, mock.AnythingOfType("*dto.CreateCompanyRequest")).Return(&dto.CompanyResponse{CompanyID: "comp-1"}, nil)

	reqBody, _ := json.Marshal(dto.CreateCompanyRequest{BusinessName: "Test Co"})
	req := httptest.NewRequest("POST", "/companies", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	ctrl.CreateCompany(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestCompanyController_GetCompany_ServiceError(t *testing.T) {
	mockSvc := new(MockCompanyService)
	ctrl := controllers.NewCompanyController(mockSvc)

	mockSvc.On("GetCompanyByID", mock.Anything, "comp-1").Return(nil, errors.New("not found"))

	req := httptest.NewRequest("GET", "/companies/comp-1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "comp-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	ctrl.GetCompany(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCompanyController_GetAllCompanies_Success(t *testing.T) {
	mockSvc := new(MockCompanyService)
	ctrl := controllers.NewCompanyController(mockSvc)

	mockSvc.On("GetAllCompanies", mock.Anything).Return([]dto.CompanyResponse{{CompanyID: "comp-1"}}, nil)

	req := httptest.NewRequest("GET", "/companies", nil)
	w := httptest.NewRecorder()
	ctrl.GetAllCompanies(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCompanyController_GetAllCompanies_ServiceError(t *testing.T) {
	mockSvc := new(MockCompanyService)
	ctrl := controllers.NewCompanyController(mockSvc)

	mockSvc.On("GetAllCompanies", mock.Anything).Return(nil, errors.New("err"))

	req := httptest.NewRequest("GET", "/companies", nil)
	w := httptest.NewRecorder()
	ctrl.GetAllCompanies(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCompanyController_UpdateCompany_Success(t *testing.T) {
	mockSvc := new(MockCompanyService)
	ctrl := controllers.NewCompanyController(mockSvc)

	mockSvc.On("UpdateCompany", mock.Anything, "comp-1", mock.AnythingOfType("*dto.UpdateCompanyRequest")).Return(&dto.CompanyResponse{CompanyID: "comp-1"}, nil)

	reqBody, _ := json.Marshal(dto.UpdateCompanyRequest{BusinessName: "Test Co"})
	req := httptest.NewRequest("PUT", "/companies/comp-1", bytes.NewBuffer(reqBody))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "comp-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	ctrl.UpdateCompany(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCompanyController_UpdateCompany_BadJSON(t *testing.T) {
	mockSvc := new(MockCompanyService)
	ctrl := controllers.NewCompanyController(mockSvc)

	req := httptest.NewRequest("PUT", "/companies/comp-1", bytes.NewBufferString("{invalid"))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "comp-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	ctrl.UpdateCompany(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCompanyController_UpdateCompany_ServiceError(t *testing.T) {
	mockSvc := new(MockCompanyService)
	ctrl := controllers.NewCompanyController(mockSvc)

	mockSvc.On("UpdateCompany", mock.Anything, "comp-1", mock.AnythingOfType("*dto.UpdateCompanyRequest")).Return(nil, errors.New("err"))

	reqBody, _ := json.Marshal(dto.UpdateCompanyRequest{BusinessName: "Test Co"})
	req := httptest.NewRequest("PUT", "/companies/comp-1", bytes.NewBuffer(reqBody))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "comp-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	ctrl.UpdateCompany(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCompanyController_DeleteCompany_ServiceError(t *testing.T) {
	mockSvc := new(MockCompanyService)
	ctrl := controllers.NewCompanyController(mockSvc)

	mockSvc.On("DeleteCompany", mock.Anything, "comp-1").Return(errors.New("err"))

	req := httptest.NewRequest("DELETE", "/companies/comp-1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "comp-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	
	w := httptest.NewRecorder()
	ctrl.DeleteCompany(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
