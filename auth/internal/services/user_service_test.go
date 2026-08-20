package services_test

import (
	"context"
	"testing"

	"auth/internal/dto"
	"auth/internal/middlewares"
	"auth/internal/models"
	"auth/internal/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository implements repositories.UserRepository for unit testing
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByID(ctx context.Context, id string) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) FindAll(ctx context.Context) ([]models.User, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockUserRepository) FindAllByEmail(ctx context.Context, email string) ([]models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockUserRepository) FindAllByGoogleID(ctx context.Context, googleAuthID string) ([]models.User, error) {
	args := m.Called(ctx, googleAuthID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockUserRepository) FindByEmailAndRole(ctx context.Context, email, roleID string) (*models.User, error) {
	args := m.Called(ctx, email, roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) LinkGoogleAccountByEmail(ctx context.Context, email, googleAuthID string) error {
	args := m.Called(ctx, email, googleAuthID)
	return args.Error(0)
}

func (m *MockUserRepository) UnlinkGoogleAccountByEmail(ctx context.Context, email string) error {
	args := m.Called(ctx, email)
	return args.Error(0)
}

func TestUserService_GetUserByID_AdminSuccess(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	svc := services.NewUserService(mockRepo)

	ctx := context.WithValue(context.Background(), middlewares.RoleIDKey, "role_admin")
	expectedUser := &models.User{
		UserID:    "usr-123",
		FirstName: "Juan",
		LastName:  "Pérez",
		Email:     "juan@email.com",
		RoleID:    "role_participant",
		Status:    "active",
	}

	mockRepo.On("FindByID", ctx, "usr-123").Return(expectedUser, nil)

	// Act
	res, err := svc.GetUserByID(ctx, "usr-123")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "usr-123", res.UserID)
	assert.Equal(t, "Juan", res.FirstName)
	mockRepo.AssertExpectations(t)
}

func TestUserService_GetUserByID_ForbiddenOtherUser(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	svc := services.NewUserService(mockRepo)

	ctx := context.WithValue(context.Background(), middlewares.RoleIDKey, "role_participant")
	ctx = context.WithValue(ctx, middlewares.UserIDKey, "usr-999")
	targetUser := &models.User{
		UserID:    "usr-123",
		FirstName: "Juan",
		Email:     "juan@email.com",
	}

	mockRepo.On("FindByID", ctx, "usr-123").Return(targetUser, nil)

	// Act
	res, err := svc.GetUserByID(ctx, "usr-123")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "forbidden")
	mockRepo.AssertExpectations(t)
}

func TestUserService_GetAllUsers_NonAdminForbidden(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	svc := services.NewUserService(mockRepo)
	ctx := context.WithValue(context.Background(), middlewares.RoleIDKey, "role_participant")

	// Act
	res, err := svc.GetAllUsers(ctx)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "forbidden")
}

func TestUserService_DeleteUser_AdminSuccess(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := services.NewUserService(mockRepo)

	ctx := context.WithValue(context.Background(), middlewares.RoleIDKey, "role_admin")

	expectedUser := &models.User{UserID: "usr-123"}
	mockRepo.On("FindByID", ctx, "usr-123").Return(expectedUser, nil)
	mockRepo.On("Delete", ctx, "usr-123").Return(nil)

	err := svc.DeleteUser(ctx, "usr-123")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUserService_DeleteUser_ForbiddenOtherUser(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := services.NewUserService(mockRepo)

	ctx := context.WithValue(context.Background(), middlewares.RoleIDKey, "role_participant")
	ctx = context.WithValue(ctx, middlewares.UserIDKey, "usr-999")

	targetUser := &models.User{UserID: "usr-123"}
	mockRepo.On("FindByID", ctx, "usr-123").Return(targetUser, nil)

	err := svc.DeleteUser(ctx, "usr-123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")
	mockRepo.AssertNotCalled(t, "Delete")
}

func TestUserService_UpdateUser_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := services.NewUserService(mockRepo)

	ctx := context.WithValue(context.Background(), middlewares.RoleIDKey, "role_admin")
	
	existingUser := &models.User{
		UserID: "usr-123",
		FirstName: "Old",
	}

	mockRepo.On("FindByID", ctx, "usr-123").Return(existingUser, nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*models.User")).Return(nil)

	req := &dto.UpdateUserRequest{
		FirstName: "New",
	}

	res, err := svc.UpdateUser(ctx, "usr-123", req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "New", res.FirstName)
	mockRepo.AssertExpectations(t)
}
