package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"auth/internal/models"
	"auth/internal/services"
	"auth/internal/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type refreshFixture struct {
	users  *MockUserRepository
	tokens *MockRefreshTokenRepository
	jwt    *MockJWTService
	svc    services.AuthService
}

func newRefreshFixture() *refreshFixture {
	f := &refreshFixture{
		users:  new(MockUserRepository),
		tokens: new(MockRefreshTokenRepository),
		jwt:    new(MockJWTService),
	}
	f.svc = services.NewAuthService(nil, new(MockRoleProfileQuerier), f.users, f.tokens, f.jwt, nil)
	return f
}

func (f *refreshFixture) givenToken(rt *models.RefreshToken) {
	f.tokens.On("FindByTokenHash", mock.Anything, utils.HashToken("old-refresh")).Return(rt, nil)
}

func TestAuthService_RefreshToken_Success(t *testing.T) {
	f := newRefreshFixture()
	f.givenToken(&models.RefreshToken{RefreshTokenID: "rt-1", UserID: "usr-1", ExpiresAt: time.Now().Add(time.Hour)})
	user := &models.User{UserID: "usr-1", RoleID: "role_participant", Status: "active"}
	f.users.On("FindByID", mock.Anything, "usr-1").Return(user, nil)
	f.tokens.On("Revoke", mock.Anything, "rt-1").Return(nil)
	f.jwt.On("GenerateTokenPair", "usr-1", "role_participant", false).Return("new-access", "new-refresh", nil)
	f.tokens.On("Create", mock.Anything, mock.AnythingOfType("*models.RefreshToken")).Return(nil)
	f.tokens.On("RevokeAllForUser", mock.Anything, "usr-1").Return(nil)
	f.tokens.On("DeleteExpired", mock.Anything).Return(nil).Maybe()

	res, err := f.svc.RefreshToken(context.Background(), "old-refresh")

	require.NoError(t, err)
	assert.Equal(t, "new-access", res.Token)
	assert.Equal(t, "new-refresh", res.RefreshToken)
	assert.Equal(t, "active", res.UserStatus)
	f.tokens.AssertCalled(t, "Revoke", mock.Anything, "rt-1") // rotación: el viejo queda revocado
}

func TestAuthService_RefreshToken_NotFound(t *testing.T) {
	f := newRefreshFixture()
	f.tokens.On("FindByTokenHash", mock.Anything, mock.Anything).Return(nil, errors.New("record not found"))

	_, err := f.svc.RefreshToken(context.Background(), "old-refresh")

	assert.EqualError(t, err, "refresh token not found")
}

func TestAuthService_RefreshToken_RevokedOrExpired(t *testing.T) {
	revokedAt := time.Now().Add(-time.Minute)
	cases := map[string]*models.RefreshToken{
		"revocado": {RefreshTokenID: "rt-1", UserID: "usr-1", ExpiresAt: time.Now().Add(time.Hour), RevokedAt: &revokedAt},
		"vencido":  {RefreshTokenID: "rt-1", UserID: "usr-1", ExpiresAt: time.Now().Add(-time.Minute)},
	}
	for name, rt := range cases {
		t.Run(name, func(t *testing.T) {
			f := newRefreshFixture()
			f.givenToken(rt)

			_, err := f.svc.RefreshToken(context.Background(), "old-refresh")

			assert.EqualError(t, err, "refresh token expired or revoked")
			f.jwt.AssertNotCalled(t, "GenerateTokenPair", mock.Anything, mock.Anything, mock.Anything)
		})
	}
}

func TestAuthService_RefreshToken_UserNotFound(t *testing.T) {
	f := newRefreshFixture()
	f.givenToken(&models.RefreshToken{RefreshTokenID: "rt-1", UserID: "usr-1", ExpiresAt: time.Now().Add(time.Hour)})
	f.users.On("FindByID", mock.Anything, "usr-1").Return(nil, errors.New("not found"))

	_, err := f.svc.RefreshToken(context.Background(), "old-refresh")

	assert.EqualError(t, err, "user not found")
}

func TestAuthService_RefreshToken_JWTError(t *testing.T) {
	f := newRefreshFixture()
	f.givenToken(&models.RefreshToken{RefreshTokenID: "rt-1", UserID: "usr-1", ExpiresAt: time.Now().Add(time.Hour)})
	f.users.On("FindByID", mock.Anything, "usr-1").Return(&models.User{UserID: "usr-1", RoleID: "role_admin"}, nil)
	f.tokens.On("Revoke", mock.Anything, "rt-1").Return(nil)
	f.jwt.On("GenerateTokenPair", "usr-1", "role_admin", false).Return("", "", errors.New("jwt error"))

	_, err := f.svc.RefreshToken(context.Background(), "old-refresh")

	assert.EqualError(t, err, "jwt error")
}

func TestAuthService_RefreshToken_SaveError(t *testing.T) {
	f := newRefreshFixture()
	f.givenToken(&models.RefreshToken{RefreshTokenID: "rt-1", UserID: "usr-1", ExpiresAt: time.Now().Add(time.Hour)})
	f.users.On("FindByID", mock.Anything, "usr-1").Return(&models.User{UserID: "usr-1", RoleID: "role_admin"}, nil)
	f.tokens.On("Revoke", mock.Anything, "rt-1").Return(nil)
	f.jwt.On("GenerateTokenPair", "usr-1", "role_admin", false).Return("a", "r", nil)
	f.tokens.On("Create", mock.Anything, mock.Anything).Return(errors.New("db caída"))
	f.tokens.On("RevokeAllForUser", mock.Anything, "usr-1").Return(nil)
	f.tokens.On("DeleteExpired", mock.Anything).Return(nil).Maybe()

	_, err := f.svc.RefreshToken(context.Background(), "old-refresh")

	assert.Error(t, err)
}

func TestAuthService_GetUserProfile(t *testing.T) {
	users, querier := new(MockUserRepository), new(MockRoleProfileQuerier)
	svc := services.NewAuthService(nil, querier, users, nil, nil, nil)

	users.On("FindByID", mock.Anything, "usr-1").Return(&models.User{
		UserID: "usr-1", FirstName: "Ana", Email: "ana@x.com", RoleID: "role_participant", Status: "active",
	}, nil)
	querier.On("GetRoleData", mock.Anything, "participants", "usr-1").Return(map[string]any{"participant_id": "part-1"}, nil)

	profile, err := svc.GetUserProfile(context.Background(), "usr-1")

	require.NoError(t, err)
	user := profile["user"].(map[string]interface{})
	assert.Equal(t, "Ana", user["first_name"])
	assert.Equal(t, "ana@x.com", user["email"])
	assert.Equal(t, map[string]any{"participant_id": "part-1"}, profile["role_data"])
}

func TestAuthService_GetUserProfile_RoleTables(t *testing.T) {
	for role, table := range map[string]string{
		"role_teacher":   "teachers",
		"role_volunteer": "volunteers",
		"role_company":   "companies",
	} {
		users, querier := new(MockUserRepository), new(MockRoleProfileQuerier)
		svc := services.NewAuthService(nil, querier, users, nil, nil, nil)
		users.On("FindByID", mock.Anything, "usr-1").Return(&models.User{UserID: "usr-1", RoleID: role}, nil)
		querier.On("GetRoleData", mock.Anything, table, "usr-1").Return(nil, errors.New("sin perfil"))

		profile, err := svc.GetUserProfile(context.Background(), "usr-1")

		require.NoError(t, err, role)
		assert.NotContains(t, profile, "role_data", "un error del querier no es fatal")
		querier.AssertExpectations(t)
	}
}

func TestAuthService_GetUserProfile_AdminHasNoRoleData(t *testing.T) {
	users, querier := new(MockUserRepository), new(MockRoleProfileQuerier)
	svc := services.NewAuthService(nil, querier, users, nil, nil, nil)
	users.On("FindByID", mock.Anything, "usr-1").Return(&models.User{UserID: "usr-1", RoleID: "role_admin"}, nil)

	profile, err := svc.GetUserProfile(context.Background(), "usr-1")

	require.NoError(t, err)
	assert.NotContains(t, profile, "role_data")
	querier.AssertNotCalled(t, "GetRoleData", mock.Anything, mock.Anything, mock.Anything)
}

func TestAuthService_GetUserProfile_NotFound(t *testing.T) {
	users := new(MockUserRepository)
	svc := services.NewAuthService(nil, nil, users, nil, nil, nil)
	users.On("FindByID", mock.Anything, "zzz").Return(nil, errors.New("record not found"))

	_, err := svc.GetUserProfile(context.Background(), "zzz")

	assert.EqualError(t, err, "usuario no encontrado")
}
