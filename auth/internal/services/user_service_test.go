package services_test

import (
	"context"
	"errors"
	"github.com/stretchr/testify/require"
	"testing"
	"time"

	"auth/internal/dto"
	"auth/internal/middlewares"
	"auth/internal/models"
	"auth/internal/repositories"
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

func (m *MockUserRepository) CreateWithRoleProfile(ctx context.Context, user *models.User, roleID string, approvedBy string) error {
	return m.Called(ctx, user, roleID, approvedBy).Error(0)
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

func (m *MockUserRepository) FindParticipantsHistory(ctx context.Context) ([]models.ParticipantHistory, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ParticipantHistory), args.Error(1)
}

func (m *MockUserRepository) CountActiveParticipants(ctx context.Context) (int64, error) {
	// Mock implementation for CountActiveParticipants
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) CountVolunteers(ctx context.Context) (int64, error) {
	// Mock implementation for CountVolunteers
	args := m.Called(ctx)
	countResult := args.Get(0).(int64)
	return countResult, args.Error(1)
}

func (m *MockUserRepository) UpdateEmailVerified(ctx context.Context, userID string, verified bool) error {
	args := m.Called(ctx, userID, verified)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateEmailVerifiedByEmail(ctx context.Context, email string, verified bool) error {
	args := m.Called(ctx, email, verified)
	return args.Error(0)
}

func (m *MockUserRepository) UpdatePasswordHashByEmail(ctx context.Context, email string, passwordHash string) error {
	args := m.Called(ctx, email, passwordHash)
	return args.Error(0)
}

func (m *MockUserRepository) UpdatePasswordAndVerifiedByEmail(ctx context.Context, email string, passwordHash string, verified bool) error {
	args := m.Called(ctx, email, passwordHash, verified)
	return args.Error(0)
}

func (m *MockUserRepository) FindCampaignEligibleUsers(ctx context.Context) ([]repositories.CampaignEligibleUser, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repositories.CampaignEligibleUser), args.Error(1)
}

func (m *MockUserRepository) FindActiveUserIDsByRoles(ctx context.Context, roles []string) ([]string, error) {
	args := m.Called(ctx, roles)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
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
		UserID:    "usr-123",
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
		Email:     "test@example.com",
		FirstName: "test",
		LastName:  "test",
		Password:  longPass + "A1!a",
		RoleID:    "role_participant",
		Phone:     "1234567890",
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
		Email:     "test@example.com",
		FirstName: "test",
		LastName:  "test",
		Password:  "Password123!",
		RoleID:    "role_participant",
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

func TestUserService_GetParticipantsHistory(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := services.NewUserService(mockRepo)
	ctx := context.WithValue(context.Background(), middlewares.RoleIDKey, "role_admin")

	alta := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	baja := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)
	mockRepo.On("FindParticipantsHistory", ctx).Return([]models.ParticipantHistory{
		{UserID: "usr-1", Status: "active", ApprovalStatus: "approved", CreatedAt: alta},
		{UserID: "usr-2", Status: "active", ApprovalStatus: "approved", CreatedAt: alta, DeletedAt: &baja},
	}, nil)

	res, err := svc.GetParticipantsHistory(ctx)
	require.NoError(t, err)
	require.Len(t, res, 2)
	assert.Equal(t, "usr-1", res[0].UserID)
	assert.Nil(t, res[0].DeletedAt)
	assert.True(t, baja.Equal(*res[1].DeletedAt))
	assert.Equal(t, "approved", res[1].ApprovalStatus)
}

func TestUserService_GetParticipantsHistory_OnlyAdmin(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := services.NewUserService(mockRepo)
	for _, role := range []string{"role_volunteer", "role_participant", ""} {
		ctx := context.WithValue(context.Background(), middlewares.RoleIDKey, role)
		_, err := svc.GetParticipantsHistory(ctx)
		assert.ErrorContains(t, err, "forbidden", role)
	}
	mockRepo.AssertNotCalled(t, "FindParticipantsHistory", mock.Anything)
}

func TestUserService_GetParticipantsHistory_DBError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := services.NewUserService(mockRepo)
	ctx := context.WithValue(context.Background(), middlewares.RoleIDKey, "role_dev")
	mockRepo.On("FindParticipantsHistory", ctx).Return(nil, errors.New("db error"))
	_, err := svc.GetParticipantsHistory(ctx)
	assert.ErrorContains(t, err, "db error")
}

type mockVerificationService struct {
	calledWithUser *models.User
}

func (m *mockVerificationService) GenerateAndSendVerification(ctx context.Context, user *models.User) error {
	m.calledWithUser = user
	return nil
}

func (m *mockVerificationService) VerifyEmail(ctx context.Context, token string) error {
	return nil
}

func (m *mockVerificationService) ResendVerification(ctx context.Context, email string) error {
	return nil
}

func (m *mockVerificationService) IssueActionLink(ctx context.Context, requestID, userID string) (string, error) {
	return "", nil
}

func TestUserService_CreateUser_Success_SendsVerification(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockVerif := &mockVerificationService{}
	svc := services.NewUserService(mockRepo, mockVerif)

	ctx := context.WithValue(context.Background(), middlewares.RoleIDKey, "role_admin")
	ctx = context.WithValue(ctx, middlewares.UserIDKey, "admin-1")

	req := &dto.CreateUserRequest{
		Email:     "newuser@gmail.com",
		FirstName: "Nuevo",
		LastName:  "Usuario",
		Password:  "Password123!",
		RoleID:    "role_participant",
		Phone:     "123456789",
		DNI:       "30111222",
		Address:   "San Martín 100",
	}

	mockRepo.On("CreateWithRoleProfile", mock.Anything, mock.MatchedBy(func(u *models.User) bool {
		return u.Email == "newuser@gmail.com" && !u.IsEmailVerified
	}), "role_participant", "admin-1").Return(nil)

	res, err := svc.CreateUser(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.False(t, res.IsEmailVerified)
	assert.NotNil(t, mockVerif.calledWithUser)
	assert.Equal(t, "newuser@gmail.com", mockVerif.calledWithUser.Email)
	assert.False(t, mockVerif.calledWithUser.IsEmailVerified)
}

func TestUserService_GetAllUsers(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := services.NewUserService(mockRepo)

	// Forbidden role
	ctxGuest := context.WithValue(context.Background(), middlewares.RoleIDKey, "role_company")
	_, err := svc.GetAllUsers(ctxGuest)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")

	// Volunteer role filters only participants
	ctxVol := context.WithValue(context.Background(), middlewares.RoleIDKey, "role_volunteer")
	allUsers := []models.User{
		{UserID: "u1", RoleID: "role_participant"},
		{UserID: "u2", RoleID: "role_teacher"},
	}
	mockRepo.On("FindAll", mock.Anything).Return(allUsers, nil)
	volResp, err := svc.GetAllUsers(ctxVol)
	assert.NoError(t, err)
	assert.Len(t, volResp, 1)
	assert.Equal(t, "u1", volResp[0].UserID)

	// Admin role returns all
	ctxAdmin := context.WithValue(context.Background(), middlewares.RoleIDKey, "role_admin")
	adminResp, err := svc.GetAllUsers(ctxAdmin)
	assert.NoError(t, err)
	assert.Len(t, adminResp, 2)
}

func TestUserService_ApproveUser(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockQuerier := new(MockRoleProfileQuerier)
	producer := &mockUsuariosProducer{}
	svc := services.NewUserService(mockRepo, mockQuerier, producer)

	// FindByID error
	mockRepo.On("FindByID", mock.Anything, "u-err").Return(nil, errors.New("db error"))
	_, err := svc.ApproveUser(context.Background(), "u-err")
	assert.Error(t, err)

	// Forbidden to approve
	userTeacher := &models.User{UserID: "u-teach", RoleID: "role_teacher"}
	mockRepo.On("FindByID", mock.Anything, "u-teach").Return(userTeacher, nil)
	ctxVol := context.WithValue(context.Background(), middlewares.RoleIDKey, "role_volunteer")
	_, err = svc.ApproveUser(ctxVol, "u-teach")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")

	// Role not compatible for approval (e.g. role_admin)
	userAdmin := &models.User{UserID: "u-adm", RoleID: "role_admin"}
	mockRepo.On("FindByID", mock.Anything, "u-adm").Return(userAdmin, nil)
	ctxAdmin := context.WithValue(context.Background(), middlewares.RoleIDKey, "role_admin")
	_, err = svc.ApproveUser(ctxAdmin, "u-adm")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rol no compatible")

	// Querier error
	userPart := &models.User{UserID: "u-part", RoleID: "role_participant"}
	mockRepo.On("FindByID", mock.Anything, "u-part").Return(userPart, nil)
	mockQuerier.On("UpdateRoleData", mock.Anything, "participants", "u-part", mock.Anything).Return(errors.New("update err"))
	_, err = svc.ApproveUser(ctxAdmin, "u-part")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error al actualizar")

	// Success with approverID and producer
	ctxAdminWithID := context.WithValue(ctxAdmin, middlewares.UserIDKey, "admin-123")
	userComp := &models.User{UserID: "u-comp", RoleID: "role_company"}
	mockRepo.On("FindByID", mock.Anything, "u-comp").Return(userComp, nil)
	mockQuerier.On("UpdateRoleData", mock.Anything, "companies", "u-comp", mock.MatchedBy(func(data map[string]any) bool {
		return data["approved_by"] == "admin-123" && data["approval_status"] == "approved"
	})).Return(nil)

	res, err := svc.ApproveUser(ctxAdminWithID, "u-comp")
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "approved", res.RoleApprovalStatus)
	assert.NotEmpty(t, producer.emailSolicitadoCalls)
}

func TestUserService_GetCampaignEligibleUsers(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := services.NewUserService(mockRepo)

	mockRepo.On("FindCampaignEligibleUsers", mock.Anything).Return([]repositories.CampaignEligibleUser{
		{UserID: "u1", Email: "e1@test.com"},
	}, nil)

	res, err := svc.GetCampaignEligibleUsers(context.Background())
	assert.NoError(t, err)
	assert.Len(t, res, 1)
}

func TestUserService_IsEmailVerified(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := services.NewUserService(mockRepo)

	// Empty identifier
	_, err := svc.IsEmailVerified(context.Background(), "   ")
	assert.Error(t, err)

	// By Email
	mockRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(&models.User{IsEmailVerified: true}, nil)
	verified, err := svc.IsEmailVerified(context.Background(), "test@example.com")
	assert.NoError(t, err)
	assert.True(t, verified)

	// By UserID
	mockRepo.On("FindByID", mock.Anything, "u-id-123").Return(&models.User{IsEmailVerified: false}, nil)
	verified, err = svc.IsEmailVerified(context.Background(), "u-id-123")
	assert.NoError(t, err)
	assert.False(t, verified)
}

func TestUserService_GetActiveUserIDsByRoles(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := services.NewUserService(mockRepo)

	// Empty / invalid roles
	res, err := svc.GetActiveUserIDsByRoles(context.Background(), []string{"unknown_role"})
	assert.NoError(t, err)
	assert.Empty(t, res)

	// Valid roles
	mockRepo.On("FindActiveUserIDsByRoles", mock.Anything, []string{"role_admin", "role_volunteer"}).Return([]string{"u1", "u2"}, nil)
	res, err = svc.GetActiveUserIDsByRoles(context.Background(), []string{"role_admin", "role_volunteer"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"u1", "u2"}, res)
}

// HU-07 y HU-01: el alta de participante desde Usuarios exige Gmail, DNI
// y dirección; y guarda DNI y dirección en el perfil.
func TestUserService_CreateUser_ParticipantRequiredFields(t *testing.T) {
	ctx := context.WithValue(context.WithValue(context.Background(), middlewares.RoleIDKey, "role_admin"), middlewares.UserIDKey, "admin-1")
	casos := map[string]*dto.CreateUserRequest{
		"no gmail":      {Email: "a@hotmail.com", DNI: "30111222", Address: "Calle 1"},
		"sin DNI":       {Email: "a@gmail.com", Address: "Calle 1"},
		"sin direccion": {Email: "a@gmail.com", DNI: "30111222"},
	}
	for name, req := range casos {
		req.FirstName, req.LastName, req.Password, req.Phone, req.RoleID = "Ana", "Paz", "Password123!", "351", "role_participant"
		mockRepo := new(MockUserRepository)
		svc := services.NewUserService(mockRepo)
		_, err := svc.CreateUser(ctx, req)
		assert.Error(t, err, name)
		mockRepo.AssertNotCalled(t, "CreateWithRoleProfile", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	}
}
