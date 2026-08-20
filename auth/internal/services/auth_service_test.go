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

func (m *MockRefreshTokenRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
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
		Email: "test@example.com",
		Password: "Password123!",
		Role: "role_participant",
		FirstName: "John",
		LastName: "Doe",
	}

	mockUserRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(nil, errors.New("not found"))
	mockTransactor.On("RunInTx", mock.Anything, mock.Anything).Return(nil)
	mockJWT.On("GenerateTokenPair", mock.Anything, "role_participant", false).Return("access", "refresh", nil)
	mockRtRepo.On("RevokeAllForUser", mock.Anything, mock.Anything).Return(nil)
	mockRtRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.RefreshToken")).Return(nil)
	mockRtRepo.On("DeleteExpired", mock.Anything).Return(nil).Maybe()

	res, err := svc.Register(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "access", res.Token)
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

