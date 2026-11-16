package services_test

import (
	"context"
	"errors"
	"testing"

	firebaseAuth "firebase.google.com/go/v4/auth"
	"github.com/golang-jwt/jwt/v5"

	"auth/internal/dto"
	"auth/internal/models"
	"auth/internal/services"
	"auth/internal/repositories"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTransactor struct {
	mock.Mock
}

func (m *MockTransactor) RunInTx(ctx context.Context, fn func(tx repositories.Tx) error) error {
	args := m.Called(ctx, fn)
	// We simulate calling fn if it's supposed to succeed
	if fn != nil && args.Error(0) == nil {
		mockTx := new(MockTx)
		mockTx.On("Create", mock.Anything).Return(nil)
		return fn(mockTx)
	}
	return args.Error(0)
}

type MockTx struct {
	mock.Mock
}

func (m *MockTx) Create(value any) error {
	args := m.Called(value)
	return args.Error(0)
}

type MockRoleProfileQuerier struct {
	mock.Mock
}

func (m *MockRoleProfileQuerier) GetApprovalStatus(ctx context.Context, table, userID string) (string, error) {
	args := m.Called(ctx, table, userID)
	return args.String(0), args.Error(1)
}

func (m *MockRoleProfileQuerier) GetIsExternal(ctx context.Context, userID string) (bool, error) {
	args := m.Called(ctx, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockRoleProfileQuerier) GetRoleData(ctx context.Context, table, userID string) (map[string]any, error) {
	args := m.Called(ctx, table, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]any), args.Error(1)
}

func (m *MockRoleProfileQuerier) UpdateRoleData(ctx context.Context, table, userID string, data map[string]any) error {
	args := m.Called(ctx, table, userID, data)
	return args.Error(0)
}

type MockJWTService struct {
	mock.Mock
}

func (m *MockJWTService) GenerateTokenPair(userID string, role string, isExternal bool) (string, string, error) {
	args := m.Called(userID, role, isExternal)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockJWTService) ValidateToken(tokenString string) (jwt.MapClaims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(jwt.MapClaims), args.Error(1)
}

type MockRefreshTokenRepository struct {
	mock.Mock
}

func (m *MockRefreshTokenRepository) Create(ctx context.Context, rt *models.RefreshToken) error {
	args := m.Called(ctx, rt)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*models.RefreshToken, error) {
	args := m.Called(ctx, tokenHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RefreshToken), args.Error(1)
}

func (m *MockRefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) Revoke(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) DeleteExpired(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type MockFirebaseService struct {
	mock.Mock
}

func (m *MockFirebaseService) VerifyIDToken(ctx context.Context, idToken string) (*firebaseAuth.Token, error) {
	args := m.Called(ctx, idToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*firebaseAuth.Token), args.Error(1)
}


func TestAuthService_Register_Success(t *testing.T) {
	mockTransactor := new(MockTransactor)
	mockQuerier := new(MockRoleProfileQuerier)
	mockUserRepo := new(MockUserRepository)
	mockRtRepo := new(MockRefreshTokenRepository)
	mockJWT := new(MockJWTService)
	mockFB := new(MockFirebaseService)

	svc := services.NewAuthService(mockTransactor, mockQuerier, mockUserRepo, mockRtRepo, mockJWT, mockFB)

	req := &dto.RegisterRequest{
		Email: "test@gmail.com",
		Password: "Password123!",
		Role: "role_participant",
		FirstName: "John",
		LastName: "Doe",
	}

	mockUserRepo.On("FindByEmail", mock.Anything, "test@gmail.com").Return(nil, errors.New("not found"))
	mockTransactor.On("RunInTx", mock.Anything, mock.Anything).Return(nil)

	res, err := svc.Register(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Empty(t, res.Token)
	assert.Empty(t, res.RefreshToken)
	assert.NotNil(t, res.User)
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_Register_InvalidInput(t *testing.T) {
	svc := services.NewAuthService(nil, nil, nil, nil, nil, nil)

	req := &dto.RegisterRequest{Email: ""}
	res, err := svc.Register(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "obligatorios")
}

func TestAuthService_CheckCredentials_Missing(t *testing.T) {
	svc := services.NewAuthService(nil, nil, nil, nil, nil, nil)

	req := &dto.CheckCredentialsRequest{Email: ""}
	res, err := svc.CheckCredentials(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "obligatorios")
}

func TestAuthService_CheckCredentials_UserNotFound(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	svc := services.NewAuthService(nil, nil, mockUserRepo, nil, nil, nil)

	req := &dto.CheckCredentialsRequest{Email: "test@test.com", Password: "Pass"}
	mockUserRepo.On("FindAllByEmail", mock.Anything, "test@test.com").Return(nil, errors.New("not found"))

	res, err := svc.CheckCredentials(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "no existe")
}

func TestAuthService_Login_MissingFields(t *testing.T) {
	svc := services.NewAuthService(nil, nil, nil, nil, nil, nil)
	req := &dto.LoginRequest{UserID: "", RoleID: "role_participant"}
	res, err := svc.Login(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "obligatorios")
}

func TestAuthService_GetMyProfile_UserNotFound(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	svc := services.NewAuthService(nil, nil, mockUserRepo, nil, nil, nil)
	
	mockUserRepo.On("FindByID", mock.Anything, "usr-missing").Return(nil, errors.New("not found"))
	
	res, err := svc.GetMyProfile(context.Background(), "role_participant", "usr-missing")
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "usuario no encontrado")
}

func TestAuthService_GetMyProfile_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockQuerier := new(MockRoleProfileQuerier)
	svc := services.NewAuthService(nil, mockQuerier, mockUserRepo, nil, nil, nil)
	
	user := &models.User{UserID: "usr-123", FirstName: "Test", RoleID: "role_participant"}
	mockUserRepo.On("FindByID", mock.Anything, "usr-123").Return(user, nil)
	mockQuerier.On("GetRoleData", mock.Anything, "participants", "usr-123").Return(map[string]any{"extra": "data"}, nil)
	
	res, err := svc.GetMyProfile(context.Background(), "role_participant", "usr-123")
	assert.NoError(t, err)
	assert.NotNil(t, res)
	userData := res["user"].(map[string]interface{})
	assert.Equal(t, "usr-123", userData["user_id"])
}

func TestAuthService_GetMyProfile_RoleAdmin(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	svc := services.NewAuthService(nil, nil, mockUserRepo, nil, nil, nil)
	
	user := &models.User{UserID: "usr-123", RoleID: "role_admin"}
	mockUserRepo.On("FindByID", mock.Anything, "usr-123").Return(user, nil)
	
	res, err := svc.GetMyProfile(context.Background(), "role_admin", "usr-123")
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotContains(t, res, "role_data")
}

func TestAuthService_GetMyProfile_QuerierError(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockQuerier := new(MockRoleProfileQuerier)
	svc := services.NewAuthService(nil, mockQuerier, mockUserRepo, nil, nil, nil)
	
	user := &models.User{UserID: "usr-123", RoleID: "role_participant"}
	mockUserRepo.On("FindByID", mock.Anything, "usr-123").Return(user, nil)
	mockQuerier.On("GetRoleData", mock.Anything, "participants", "usr-123").Return(nil, errors.New("querier error"))
	
	res, err := svc.GetMyProfile(context.Background(), "role_participant", "usr-123")
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotContains(t, res, "role_data")
}

func TestAuthService_UpdateMyPassword_WrongOldPassword(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	svc := services.NewAuthService(nil, nil, mockUserRepo, nil, nil, nil)
	
	user := &models.User{UserID: "usr-123", PasswordHash: "$2a$10$wN9Q.Q4Q.Q4Q.Q4Q.Q4Q.e.Q4Q.Q4Q.Q4Q.Q4Q.Q4Q.Q4Q.Q4Q"} // Dummy bcrypt
	mockUserRepo.On("FindByID", mock.Anything, "usr-123").Return(user, nil)
	
	err := svc.UpdateMyPassword(context.Background(), "usr-123", "wrong", "newpass")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "incorrecta")
}

func TestAuthService_GetApprovalStatus_AllRoles(t *testing.T) {
	mockQuerier := new(MockRoleProfileQuerier)
	svc := services.NewAuthService(nil, mockQuerier, nil, nil, nil, nil)
	
	mockQuerier.On("GetApprovalStatus", mock.Anything, "companies", "usr-1").Return("pending", nil)
	mockQuerier.On("GetApprovalStatus", mock.Anything, "teachers", "usr-2").Return("active", nil)
	
	status1 := svc.GetApprovalStatus(context.Background(), "role_company", "usr-1")
	assert.Equal(t, "pending", status1)
	
	status2 := svc.GetApprovalStatus(context.Background(), "role_teacher", "usr-2")
	assert.Equal(t, "active", status2)
	
	statusAdmin := svc.GetApprovalStatus(context.Background(), "role_admin", "usr-3")
	assert.Equal(t, "approved", statusAdmin)
}

func TestAuthService_Register_UserExists(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	svc := services.NewAuthService(nil, nil, mockUserRepo, nil, nil, nil)

	req := &dto.RegisterRequest{
		Email: "exists@example.com",
		Password: "Password123!",
		Role: "role_participant",
	}

	mockUserRepo.On("FindByEmail", mock.Anything, "exists@example.com").Return(&models.User{}, nil)

	res, err := svc.Register(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "ya existe una cuenta")
}

func TestAuthService_Login_MissingRole(t *testing.T) {
	svc := services.NewAuthService(nil, nil, nil, nil, nil, nil)
	
	req1 := &dto.LoginRequest{UserID: "usr-1", RoleID: ""}
	res1, err1 := svc.Login(context.Background(), req1)
	assert.Error(t, err1)
	assert.Nil(t, res1)
	
	req2 := &dto.LoginRequest{UserID: "", RoleID: "role_admin"}
	res2, err2 := svc.Login(context.Background(), req2)
	assert.Error(t, err2)
	assert.Nil(t, res2)
}

func TestAuthService_CheckGoogleCredentials_InvalidToken(t *testing.T) {
	mockFB := new(MockFirebaseService)
	svc := services.NewAuthService(nil, nil, nil, nil, nil, mockFB)

	mockFB.On("VerifyIDToken", mock.Anything, "invalid").Return(nil, errors.New("invalid token"))

	req := &dto.GoogleLoginRequest{IDToken: "invalid"}
	res, err := svc.CheckGoogleCredentials(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "inválido o expiró")
}

func TestAuthService_CheckGoogleCredentials_MissingEmail(t *testing.T) {
	mockFB := new(MockFirebaseService)
	svc := services.NewAuthService(nil, nil, nil, nil, nil, mockFB)

	fbToken := &firebaseAuth.Token{
		Claims: map[string]interface{}{},
	}
	mockFB.On("VerifyIDToken", mock.Anything, "valid_no_email").Return(fbToken, nil)

	req := &dto.GoogleLoginRequest{IDToken: "valid_no_email"}
	res, err := svc.CheckGoogleCredentials(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "dirección de correo")
}

func TestAuthService_CheckGoogleCredentials_UserNotFound(t *testing.T) {
	mockFB := new(MockFirebaseService)
	mockUserRepo := new(MockUserRepository)
	svc := services.NewAuthService(nil, nil, mockUserRepo, nil, nil, mockFB)

	fbToken := &firebaseAuth.Token{
		UID: "google-123",
		Claims: map[string]interface{}{"email": "test@google.com"},
	}
	mockFB.On("VerifyIDToken", mock.Anything, "valid").Return(fbToken, nil)

	mockUserRepo.On("FindAllByGoogleID", mock.Anything, "google-123").Return([]models.User{}, nil)
	mockUserRepo.On("FindAllByEmail", mock.Anything, "test@google.com").Return([]models.User{}, nil)

	req := &dto.GoogleLoginRequest{IDToken: "valid"}
	res, err := svc.CheckGoogleCredentials(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "no existe ninguna cuenta registrada")
}

func TestAuthService_CheckGoogleCredentials_UserNotLinked(t *testing.T) {
	mockFB := new(MockFirebaseService)
	mockUserRepo := new(MockUserRepository)
	svc := services.NewAuthService(nil, nil, mockUserRepo, nil, nil, mockFB)

	fbToken := &firebaseAuth.Token{
		UID: "google-123",
		Claims: map[string]interface{}{"email": "test@google.com"},
	}
	mockFB.On("VerifyIDToken", mock.Anything, "valid").Return(fbToken, nil)

	mockUserRepo.On("FindAllByGoogleID", mock.Anything, "google-123").Return([]models.User{}, nil)
	mockUserRepo.On("FindAllByEmail", mock.Anything, "test@google.com").Return([]models.User{{UserID: "u1"}}, nil)

	req := &dto.GoogleLoginRequest{IDToken: "valid"}
	res, err := svc.CheckGoogleCredentials(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "no está vinculada a este usuario")
}

func TestAuthService_CheckGoogleCredentials_Success(t *testing.T) {
	mockFB := new(MockFirebaseService)
	mockUserRepo := new(MockUserRepository)
	svc := services.NewAuthService(nil, nil, mockUserRepo, nil, nil, mockFB)

	fbToken := &firebaseAuth.Token{
		UID: "google-123",
		Claims: map[string]interface{}{"email": "test@google.com"},
	}
	mockFB.On("VerifyIDToken", mock.Anything, "valid").Return(fbToken, nil)

	users := []models.User{
		{UserID: "u1", RoleID: "role_admin", Status: "active"},
		{UserID: "u2", RoleID: "role_participant", Status: "active"},
	}
	mockUserRepo.On("FindAllByGoogleID", mock.Anything, "google-123").Return(users, nil)

	req := &dto.GoogleLoginRequest{IDToken: "valid", GoogleID: "google-123"}
	res, err := svc.CheckGoogleCredentials(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res.Accounts, 2)
	assert.NotEmpty(t, res.LoginNonce)
	assert.False(t, res.IsEmailVerified)
	assert.Equal(t, "test@google.com", res.Email)
}

func TestAuthService_Login_InvalidNonce(t *testing.T) {
	svc := services.NewAuthService(nil, nil, nil, nil, nil, nil)

	req := &dto.LoginRequest{UserID: "usr-1", RoleID: "role_participant", LoginNonce: "invalid-nonce"}
	res, err := svc.Login(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "expiró o es inválida")
}
