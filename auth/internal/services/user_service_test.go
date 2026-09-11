package services_test

import (
	"context"
	"errors"
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

func (m *MockUserRepository) mockUserReturn(args mock.Arguments) (*models.User, error) {
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) mockUsersReturn(args mock.Arguments) ([]models.User, error) {
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	return m.Called(ctx, user).Error(0)
}

func (m *MockUserRepository) FindByID(ctx context.Context, id string) (*models.User, error) {
	return m.mockUserReturn(m.Called(ctx, id))
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	return m.mockUserReturn(m.Called(ctx, email))
}

func (m *MockUserRepository) FindAll(ctx context.Context) ([]models.User, error) {
	return m.mockUsersReturn(m.Called(ctx))
}

func (m *MockUserRepository) FindAllByEmail(ctx context.Context, email string) ([]models.User, error) {
	return m.mockUsersReturn(m.Called(ctx, email))
}

func (m *MockUserRepository) FindAllByGoogleID(ctx context.Context, googleAuthID string) ([]models.User, error) {
	return m.mockUsersReturn(m.Called(ctx, googleAuthID))
}

func (m *MockUserRepository) FindByEmailAndRole(ctx context.Context, email, roleID string) (*models.User, error) {
	return m.mockUserReturn(m.Called(ctx, email, roleID))
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	return m.Called(ctx, user).Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

func (m *MockUserRepository) LinkGoogleAccountByEmail(ctx context.Context, email, googleAuthID string) error {
	return m.Called(ctx, email, googleAuthID).Error(0)
}

func (m *MockUserRepository) UnlinkGoogleAccountByEmail(ctx context.Context, email string) error {
	return m.Called(ctx, email).Error(0)
}

func (m *MockUserRepository) CountActiveParticipants(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) CountVolunteers(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
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

func TestUserService_UpdateUser_EmptyFields(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := services.NewUserService(mockRepo)

	ctx := context.WithValue(context.Background(), middlewares.RoleIDKey, "role_admin")
	
	existingUser := &models.User{
		UserID:    "usr-123",
		FirstName: "old-first",
		LastName:  "old-last",
		Phone:     "old-phone",
		Status:    "old-status",
		RoleID:    "old-role",
	}

	mockRepo.On("FindByID", ctx, "usr-123").Return(existingUser, nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*models.User")).Return(nil)

	req := &dto.UpdateUserRequest{}
	res, err := svc.UpdateUser(ctx, "usr-123", req)

	assert.NoError(t, err)
	assert.Equal(t, "old-first", res.FirstName)
	assert.Equal(t, "old-last", res.LastName)
	assert.Equal(t, "old-phone", res.Phone)
	assert.Equal(t, "old-status", res.Status)
	assert.Equal(t, "old-role", res.RoleID)

	mockRepo.AssertExpectations(t)
}

func TestUserService_CreateUser_ValidationError(t *testing.T) {
	svc := services.NewUserService(nil)
	req := &dto.CreateUserRequest{
		Email: "invalid-email", // Will fail validation
	}

	res, err := svc.CreateUser(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestUserService_CreateUser_BcryptError(t *testing.T) {
	// We can trigger bcrypt error by providing a password that is too long
	// bcrypt has a 72 byte limit
	svc := services.NewUserService(nil)
	longPass := ""
	for i := 0; i < 100; i++ {
		longPass += "a"
	}
	req := &dto.CreateUserRequest{
		Email: "test@example.com",
		FirstName: "test",
		LastName: "test",
		Password: longPass + "A1!a", 
		RoleID: "role_participant",
		Phone: "1234567890",
	}

	res, err := svc.CreateUser(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "error encrypting password")
}

func TestUserService_CreateUser_DBError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := services.NewUserService(mockRepo)
	
	req := &dto.CreateUserRequest{
		Email: "test@example.com",
		FirstName: "test",
		LastName: "test",
		Password: "Password123!", 
		RoleID: "role_participant",
	}

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).Return(errors.New("db error"))

	res, err := svc.CreateUser(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestUserService_GetUserByID_DBError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := services.NewUserService(mockRepo)

	mockRepo.On("FindByID", mock.Anything, "usr-1").Return(nil, errors.New("db error"))

	res, err := svc.GetUserByID(context.Background(), "usr-1")
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestUserService_GetAllUsers_DBError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := services.NewUserService(mockRepo)

	ctx := context.WithValue(context.Background(), middlewares.RoleIDKey, "role_admin")
	mockRepo.On("FindAll", ctx).Return(nil, errors.New("db error"))

	res, err := svc.GetAllUsers(ctx)
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestUserService_UpdateUser_ValidationError(t *testing.T) {
	// If we don't have strict update rules, let's at least test DB Find error
	mockRepo := new(MockUserRepository)
	svc2 := services.NewUserService(mockRepo)
	
	mockRepo.On("FindByID", mock.Anything, "usr-1").Return(nil, errors.New("db error"))

	res, err := svc2.UpdateUser(context.Background(), "usr-1", &dto.UpdateUserRequest{})
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestUserService_UpdateUser_ForbiddenOtherUser(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := services.NewUserService(mockRepo)

	ctx := context.WithValue(context.Background(), middlewares.RoleIDKey, "role_participant")
	ctx = context.WithValue(ctx, middlewares.UserIDKey, "usr-999")

	targetUser := &models.User{UserID: "usr-123"}
	mockRepo.On("FindByID", ctx, "usr-123").Return(targetUser, nil)

	res, err := svc.UpdateUser(ctx, "usr-123", &dto.UpdateUserRequest{})
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "forbidden")
}

func TestUserService_UpdateUser_DBUpdateError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := services.NewUserService(mockRepo)

	ctx := context.WithValue(context.Background(), middlewares.RoleIDKey, "role_admin")
	
	existingUser := &models.User{UserID: "usr-123"}
	mockRepo.On("FindByID", ctx, "usr-123").Return(existingUser, nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*models.User")).Return(errors.New("db error"))

	res, err := svc.UpdateUser(ctx, "usr-123", &dto.UpdateUserRequest{})
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestUserService_DeleteUser_DBError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := services.NewUserService(mockRepo)

	mockRepo.On("FindByID", mock.Anything, "usr-1").Return(nil, errors.New("db error"))

	err := svc.DeleteUser(context.Background(), "usr-1")
	assert.Error(t, err)
}
