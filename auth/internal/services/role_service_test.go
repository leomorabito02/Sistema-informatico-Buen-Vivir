package services_test

import (
	"context"
	"errors"
	"testing"

	"auth/internal/dto"
	"auth/internal/models"
	"auth/internal/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRoleRepository struct {
	mock.Mock
}

func (m *MockRoleRepository) Create(ctx context.Context, role *models.Role) error {
	return m.Called(ctx, role).Error(0)
}

func (m *MockRoleRepository) FindByID(ctx context.Context, id string) (*models.Role, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Role), args.Error(1)
}

func (m *MockRoleRepository) FindAll(ctx context.Context) ([]models.Role, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Role), args.Error(1)
}

func (m *MockRoleRepository) Update(ctx context.Context, role *models.Role) error {
	return m.Called(ctx, role).Error(0)
}

func (m *MockRoleRepository) Delete(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

func TestRoleService_GetAllRoles_Success(t *testing.T) {
	mockRepo := new(MockRoleRepository)
	svc := services.NewRoleService(mockRepo)

	expectedRoles := []models.Role{
		{RoleID: "role_1", Name: "Admin"},
		{RoleID: "role_2", Name: "User"},
	}

	mockRepo.On("FindAll", context.Background()).Return(expectedRoles, nil)

	res, err := svc.GetAllRoles(context.Background())

	assert.NoError(t, err)
	assert.Len(t, res, 2)
	assert.Equal(t, "Admin", res[0].Name)
	mockRepo.AssertExpectations(t)
}

func TestRoleService_GetAllRoles_RepoError(t *testing.T) {
	mockRepo := new(MockRoleRepository)
	svc := services.NewRoleService(mockRepo)

	mockRepo.On("FindAll", context.Background()).Return(nil, errors.New("db error"))

	res, err := svc.GetAllRoles(context.Background())

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "error obteniendo roles")
	mockRepo.AssertExpectations(t)
}

func TestRoleService_DeleteRole_NotFound(t *testing.T) {
	mockRepo := new(MockRoleRepository)
	svc := services.NewRoleService(mockRepo)

	mockRepo.On("FindByID", context.Background(), "role_invalid").Return(nil, errors.New("not found"))

	err := svc.DeleteRole(context.Background(), "role_invalid")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rol no encontrado")
	mockRepo.AssertNotCalled(t, "Delete")
}

func TestRoleService_CreateRole_ValidationError(t *testing.T) {
	svc := services.NewRoleService(nil)
	req := &dto.CreateRoleRequest{
		Name: "", // Assume empty name fails
	}
	res, err := svc.CreateRole(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "datos inválidos")
}

func TestRoleService_CreateRole_DBError(t *testing.T) {
	mockRepo := new(MockRoleRepository)
	svc := services.NewRoleService(mockRepo)
	
	req := &dto.CreateRoleRequest{
		Name: "Admin",
	}

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Role")).Return(errors.New("db error"))

	res, err := svc.CreateRole(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "error creando rol")
}

func TestRoleService_GetRoleByID_DBError(t *testing.T) {
	mockRepo := new(MockRoleRepository)
	svc := services.NewRoleService(mockRepo)

	mockRepo.On("FindByID", mock.Anything, "role_1").Return(nil, errors.New("db error"))

	res, err := svc.GetRoleByID(context.Background(), "role_1")
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "rol no encontrado")
}

func TestRoleService_UpdateRole_ValidationError(t *testing.T) {
	svc := services.NewRoleService(nil)
	
	res, err := svc.UpdateRole(context.Background(), "role_1", &dto.UpdateRoleRequest{})
	assert.Error(t, err)
	assert.Nil(t, res)
	// Even if validation passes empty, let's just make sure we hit the DB error or validation error
}

func TestRoleService_UpdateRole_DBError_Find(t *testing.T) {
	mockRepo := new(MockRoleRepository)
	svc := services.NewRoleService(mockRepo)
	
	mockRepo.On("FindByID", mock.Anything, "role_1").Return(nil, errors.New("db error"))

	res, err := svc.UpdateRole(context.Background(), "role_1", &dto.UpdateRoleRequest{Name: "New"})
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "rol no encontrado")
}

func TestRoleService_UpdateRole_DBError_Update(t *testing.T) {
	mockRepo := new(MockRoleRepository)
	svc := services.NewRoleService(mockRepo)
	
	role := &models.Role{RoleID: "role_1"}
	mockRepo.On("FindByID", mock.Anything, "role_1").Return(role, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.Role")).Return(errors.New("db error"))

	res, err := svc.UpdateRole(context.Background(), "role_1", &dto.UpdateRoleRequest{Name: "New"})
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "error actualizando rol")
}

func TestRoleService_DeleteRole_DBError(t *testing.T) {
	mockRepo := new(MockRoleRepository)
	svc := services.NewRoleService(mockRepo)
	
	role := &models.Role{RoleID: "role_1"}
	mockRepo.On("FindByID", mock.Anything, "role_1").Return(role, nil)
	mockRepo.On("Delete", mock.Anything, "role_1").Return(errors.New("db error"))

	err := svc.DeleteRole(context.Background(), "role_1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error eliminando rol")
}
