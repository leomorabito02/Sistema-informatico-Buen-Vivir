package services_test

import (
	"context"
	"errors"
	"testing"

	"auth/internal/middlewares"

	"auth/internal/dto"
	"auth/internal/models"
	"auth/internal/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockCompanyRepository struct {
	mock.Mock
}

func (m *MockCompanyRepository) Create(ctx context.Context, company *models.Company) error {
	return m.Called(ctx, company).Error(0)
}

func (m *MockCompanyRepository) mockCompanyReturn(args mock.Arguments) (*models.Company, error) {
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Company), args.Error(1)
}

func (m *MockCompanyRepository) FindByID(ctx context.Context, id string) (*models.Company, error) {
	return m.mockCompanyReturn(m.Called(ctx, id))
}

func (m *MockCompanyRepository) FindByUserID(ctx context.Context, userID string) (*models.Company, error) {
	return m.mockCompanyReturn(m.Called(ctx, userID))
}

func (m *MockCompanyRepository) FindAll(ctx context.Context) ([]models.Company, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Company), args.Error(1)
}

func (m *MockCompanyRepository) Update(ctx context.Context, company *models.Company) error {
	return m.Called(ctx, company).Error(0)
}

func (m *MockCompanyRepository) Delete(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

func TestCompanyService_GetAllCompanies_Success(t *testing.T) {
	mockRepo := new(MockCompanyRepository)
	svc := services.NewCompanyService(mockRepo)

	expected := []models.Company{{CompanyID: "comp-1", BusinessName: "Empresa SA"}}
	mockRepo.On("FindAll", mock.Anything).Return(expected, nil)

	res, err := svc.GetAllCompanies(context.Background())

	assert.NoError(t, err)
	assert.Len(t, res, 1)
	assert.Equal(t, "Empresa SA", res[0].BusinessName)
	mockRepo.AssertExpectations(t)
}

func TestCompanyService_GetCompanyByID_NotFound(t *testing.T) {
	mockRepo := new(MockCompanyRepository)
	svc := services.NewCompanyService(mockRepo)

	mockRepo.On("FindByID", mock.Anything, "invalid").Return(nil, errors.New("not found"))

	res, err := svc.GetCompanyByID(context.Background(), "invalid")

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "not found")
}

func TestCompanyService_UpdateCompany_Success(t *testing.T) {
	mockRepo := new(MockCompanyRepository)
	svc := services.NewCompanyService(mockRepo)

	existing := &models.Company{CompanyID: "comp-1", BusinessName: "Old"}
	mockRepo.On("FindByID", mock.Anything, "comp-1").Return(existing, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.Company")).Return(nil)

	req := &dto.UpdateCompanyRequest{BusinessName: "New"}
	res, err := svc.UpdateCompany(context.Background(), "comp-1", req)

	assert.NoError(t, err)
	assert.Equal(t, "New", res.BusinessName)
	mockRepo.AssertExpectations(t)
}

func TestCompanyService_UpdateCompany_EmptyFields(t *testing.T) {
	mockRepo := new(MockCompanyRepository)
	svc := services.NewCompanyService(mockRepo)

	existing := &models.Company{
		CompanyID:      "comp-1",
		BusinessName:   "old-name",
		Cuit:           "old-cuit",
		ApprovalStatus: "old-status",
	}
	mockRepo.On("FindByID", mock.Anything, "comp-1").Return(existing, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.Company")).Return(nil)

	// Send an empty update request
	req := &dto.UpdateCompanyRequest{}
	res, err := svc.UpdateCompany(context.Background(), "comp-1", req)

	assert.NoError(t, err)
	// Verify fields were NOT updated to empty string
	assert.Equal(t, "old-name", res.BusinessName)
	assert.Equal(t, "old-cuit", res.Cuit)
	assert.Equal(t, "old-status", res.ApprovalStatus)

	mockRepo.AssertExpectations(t)
}

func TestCompanyService_CreateCompany_ValidationError(t *testing.T) {
	svc := services.NewCompanyService(nil)
	req := &dto.CreateCompanyRequest{
		BusinessName: "", // Assuming empty business name fails validation
	}
	res, err := svc.CreateCompany(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestCompanyService_CreateCompany_DBError(t *testing.T) {
	mockRepo := new(MockCompanyRepository)
	svc := services.NewCompanyService(mockRepo)
	
	req := &dto.CreateCompanyRequest{
		BusinessName: "Valid Name",
		Cuit: "20-12345678-9",
		UserID: "usr-1",
	}

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Company")).Return(errors.New("db error"))

	res, err := svc.CreateCompany(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestCompanyService_GetCompanyByID_Forbidden(t *testing.T) {
	mockRepo := new(MockCompanyRepository)
	svc := services.NewCompanyService(mockRepo)

	ctx := context.WithValue(context.Background(), middlewares.RoleIDKey, "role_company")
	ctx = context.WithValue(ctx, middlewares.UserIDKey, "usr-other")

	company := &models.Company{UserID: "usr-123"}
	mockRepo.On("FindByID", mock.Anything, "comp-1").Return(company, nil)

	res, err := svc.GetCompanyByID(ctx, "comp-1")
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "forbidden")
}

func TestCompanyService_GetAllCompanies_Forbidden(t *testing.T) {
	svc := services.NewCompanyService(nil)
	ctx := context.WithValue(context.Background(), middlewares.RoleIDKey, "role_company")

	res, err := svc.GetAllCompanies(ctx)
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "forbidden")
}

func TestCompanyService_GetAllCompanies_DBError(t *testing.T) {
	mockRepo := new(MockCompanyRepository)
	svc := services.NewCompanyService(mockRepo)
	
	ctx := context.WithValue(context.Background(), middlewares.RoleIDKey, "role_admin")
	mockRepo.On("FindAll", mock.Anything).Return(nil, errors.New("db error"))

	res, err := svc.GetAllCompanies(ctx)
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestCompanyService_UpdateCompany_ValidationError(t *testing.T) {
	// Create request that will fail validation
	mockRepo := new(MockCompanyRepository)
	svc2 := services.NewCompanyService(mockRepo)
	mockRepo.On("FindByID", mock.Anything, "comp-1").Return(nil, errors.New("db error"))

	res, err := svc2.UpdateCompany(context.Background(), "comp-1", &dto.UpdateCompanyRequest{})
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestCompanyService_UpdateCompany_Forbidden(t *testing.T) {
	mockRepo := new(MockCompanyRepository)
	svc := services.NewCompanyService(mockRepo)

	ctx := context.WithValue(context.Background(), middlewares.RoleIDKey, "role_company")
	ctx = context.WithValue(ctx, middlewares.UserIDKey, "usr-other")

	company := &models.Company{UserID: "usr-123"}
	mockRepo.On("FindByID", mock.Anything, "comp-1").Return(company, nil)

	res, err := svc.UpdateCompany(ctx, "comp-1", &dto.UpdateCompanyRequest{})
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "forbidden")
}

func TestCompanyService_UpdateCompany_DBError(t *testing.T) {
	mockRepo := new(MockCompanyRepository)
	svc := services.NewCompanyService(mockRepo)

	ctx := context.WithValue(context.Background(), middlewares.RoleIDKey, "role_admin")
	company := &models.Company{UserID: "usr-123"}
	
	mockRepo.On("FindByID", mock.Anything, "comp-1").Return(company, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.Company")).Return(errors.New("db error"))

	res, err := svc.UpdateCompany(ctx, "comp-1", &dto.UpdateCompanyRequest{ApprovalStatus: "approved", ApprovedBy: "admin"})
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestCompanyService_DeleteCompany_Forbidden(t *testing.T) {
	mockRepo := new(MockCompanyRepository)
	svc := services.NewCompanyService(mockRepo)

	ctx := context.WithValue(context.Background(), middlewares.RoleIDKey, "role_company")
	ctx = context.WithValue(ctx, middlewares.UserIDKey, "usr-other")

	company := &models.Company{UserID: "usr-123"}
	mockRepo.On("FindByID", mock.Anything, "comp-1").Return(company, nil)

	err := svc.DeleteCompany(ctx, "comp-1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
}

func TestCompanyService_DeleteCompany_DBError(t *testing.T) {
	mockRepo := new(MockCompanyRepository)
	svc := services.NewCompanyService(mockRepo)

	mockRepo.On("FindByID", mock.Anything, "comp-1").Return(nil, errors.New("db error"))

	err := svc.DeleteCompany(context.Background(), "comp-1")
	assert.Error(t, err)
}
