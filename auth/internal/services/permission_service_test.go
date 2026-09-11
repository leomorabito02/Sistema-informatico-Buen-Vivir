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

type MockPermissionRepository struct {
	mock.Mock
}

func (m *MockPermissionRepository) Create(ctx context.Context, perm *models.Permission) error {
	return m.Called(ctx, perm).Error(0)
}

func (m *MockPermissionRepository) FindByID(ctx context.Context, id string) (*models.Permission, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Permission), args.Error(1)
}

func (m *MockPermissionRepository) FindAll(ctx context.Context) ([]models.Permission, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Permission), args.Error(1)
}

func (m *MockPermissionRepository) Update(ctx context.Context, perm *models.Permission) error {
	return m.Called(ctx, perm).Error(0)
}

func (m *MockPermissionRepository) Delete(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

func (m *MockPermissionRepository) CheckRolePermission(ctx context.Context, roleID, action, resource string, count *int64) error {
	args := m.Called(ctx, roleID, action, resource, count)
	*count = args.Get(0).(int64)
	return args.Error(1)
}

func TestPermissionService_GetAllPermissions_Success(t *testing.T) {
	mockRepo := new(MockPermissionRepository)
	svc := services.NewPermissionService(mockRepo)

	expected := []models.Permission{{PermissionID: "perm-1", Action: "read", Resource: "users"}}
	mockRepo.On("FindAll", mock.Anything).Return(expected, nil)

	res, err := svc.GetAllPermissions(context.Background())

	assert.NoError(t, err)
	assert.Len(t, res, 1)
	assert.Equal(t, "read", res[0].Action)
	mockRepo.AssertExpectations(t)
}

func TestPermissionService_CheckPermission_Success(t *testing.T) {
	mockRepo := new(MockPermissionRepository)
	svc := services.NewPermissionService(mockRepo)

	mockRepo.On("CheckRolePermission", mock.Anything, "role-1", "read", "users", mock.AnythingOfType("*int64")).Run(func(args mock.Arguments) {
		arg := args.Get(4).(*int64)
		*arg = 1
	}).Return(int64(1), nil)

	hasPerm, err := svc.CheckPermission(context.Background(), "role-1", "read", "users")

	assert.NoError(t, err)
	assert.True(t, hasPerm)
}

func TestPermissionService_CheckPermission_NotFound(t *testing.T) {
	mockRepo := new(MockPermissionRepository)
	svc := services.NewPermissionService(mockRepo)

	mockRepo.On("CheckRolePermission", mock.Anything, "role-1", "read", "invalid", mock.AnythingOfType("*int64")).Run(func(args mock.Arguments) {
		arg := args.Get(4).(*int64)
		*arg = 0
	}).Return(int64(0), errors.New("not found"))

	hasPerm, err := svc.CheckPermission(context.Background(), "role-1", "read", "invalid")

	assert.Error(t, err)
	assert.False(t, hasPerm)
}

func TestPermissionService_CreatePermission_ValidationError(t *testing.T) {
	svc := services.NewPermissionService(nil)
	req := &dto.CreatePermissionRequest{
		Action: "", // Assume empty fails
	}
	res, err := svc.CreatePermission(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestPermissionService_CreatePermission_DBError(t *testing.T) {
	mockRepo := new(MockPermissionRepository)
	svc := services.NewPermissionService(mockRepo)
	
	req := &dto.CreatePermissionRequest{
		Action: "read",
	}

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Permission")).Return(errors.New("db error"))

	res, err := svc.CreatePermission(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestPermissionService_GetPermissionByID_DBError(t *testing.T) {
	mockRepo := new(MockPermissionRepository)
	svc := services.NewPermissionService(mockRepo)

	mockRepo.On("FindByID", mock.Anything, "perm-1").Return(nil, errors.New("db error"))

	res, err := svc.GetPermissionByID(context.Background(), "perm-1")
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestPermissionService_GetAllPermissions_DBError(t *testing.T) {
	mockRepo := new(MockPermissionRepository)
	svc := services.NewPermissionService(mockRepo)
	
	mockRepo.On("FindAll", mock.Anything).Return(nil, errors.New("db error"))

	res, err := svc.GetAllPermissions(context.Background())
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestPermissionService_UpdatePermission_ValidationError(t *testing.T) {
	// Create request that will fail validation
	mockRepo := new(MockPermissionRepository)
	svc2 := services.NewPermissionService(mockRepo)
	mockRepo.On("FindByID", mock.Anything, "perm-1").Return(nil, errors.New("db error"))

	res, err := svc2.UpdatePermission(context.Background(), "perm-1", &dto.UpdatePermissionRequest{})
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestPermissionService_UpdatePermission_DBError_Update(t *testing.T) {
	mockRepo := new(MockPermissionRepository)
	svc := services.NewPermissionService(mockRepo)
	
	perm := &models.Permission{PermissionID: "perm-1"}
	mockRepo.On("FindByID", mock.Anything, "perm-1").Return(perm, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.Permission")).Return(errors.New("db error"))

	res, err := svc.UpdatePermission(context.Background(), "perm-1", &dto.UpdatePermissionRequest{Action: "write"})
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestPermissionService_DeletePermission_DBError(t *testing.T) {
	mockRepo := new(MockPermissionRepository)
	svc := services.NewPermissionService(mockRepo)
	
	mockRepo.On("FindByID", mock.Anything, "perm-1").Return(nil, errors.New("db error"))

	err := svc.DeletePermission(context.Background(), "perm-1")
	assert.Error(t, err)
}
