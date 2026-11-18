package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"auth/internal/dto"
	"auth/internal/models"
	"auth/internal/services"
	"auth/internal/utils"

	firebaseAuth "firebase.google.com/go/v4/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Login tests
func TestAuthService_Login_UserNotFound(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	svc := services.NewAuthService(nil, nil, mockUserRepo, nil, nil, nil)

	userIDs := []string{"usr-1"}
	nonce, _ := utils.GetNonceStore().Generate(userIDs, time.Minute)

	req := &dto.LoginRequest{UserID: "usr-1", RoleID: "role_participant", LoginNonce: nonce}
	mockUserRepo.On("FindByID", mock.Anything, "usr-1").Return(nil, errors.New("not found"))

	res, err := svc.Login(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "no fue encontrado")
}

func TestAuthService_Login_RoleMismatch(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	svc := services.NewAuthService(nil, nil, mockUserRepo, nil, nil, nil)

	userIDs := []string{"usr-1"}
	nonce, _ := utils.GetNonceStore().Generate(userIDs, time.Minute)

	req := &dto.LoginRequest{UserID: "usr-1", RoleID: "role_admin", LoginNonce: nonce}
	user := &models.User{UserID: "usr-1", RoleID: "role_participant"}
	mockUserRepo.On("FindByID", mock.Anything, "usr-1").Return(user, nil)

	res, err := svc.Login(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "no corresponde a este usuario")
}

func TestAuthService_Login_StatusBlocked(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	svc := services.NewAuthService(nil, nil, mockUserRepo, nil, nil, nil)

	userIDs := []string{"usr-1"}
	nonce, _ := utils.GetNonceStore().Generate(userIDs, time.Minute)

	req := &dto.LoginRequest{UserID: "usr-1", RoleID: "role_participant", LoginNonce: nonce}
	user := &models.User{UserID: "usr-1", RoleID: "role_participant", Status: "blocked"}
	mockUserRepo.On("FindByID", mock.Anything, "usr-1").Return(user, nil)

	res, err := svc.Login(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	if res != nil {
		assert.Equal(t, "blocked", res.UserStatus)
	}
}

func TestAuthService_Login_Unapproved(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockQuerier := new(MockRoleProfileQuerier)
	svc := services.NewAuthService(nil, mockQuerier, mockUserRepo, nil, nil, nil)

	userIDs := []string{"usr-1"}
	nonce, _ := utils.GetNonceStore().Generate(userIDs, time.Minute)

	req := &dto.LoginRequest{UserID: "usr-1", RoleID: "role_participant", LoginNonce: nonce}
	user := &models.User{UserID: "usr-1", RoleID: "role_participant", Status: "active", IsEmailVerified: true}

	mockUserRepo.On("FindByID", mock.Anything, "usr-1").Return(user, nil)
	mockQuerier.On("GetApprovalStatus", mock.Anything, "participants", "usr-1").Return("rejected", nil)

	res, err := svc.Login(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, services.ErrAccountPendingApproval, err)
}

func TestAuthService_Login_EmailNotVerified(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	svc := services.NewAuthService(nil, nil, mockUserRepo, nil, nil, nil)

	userIDs := []string{"usr-1"}
	nonce, _ := utils.GetNonceStore().Generate(userIDs, time.Minute)

	req := &dto.LoginRequest{UserID: "usr-1", RoleID: "role_participant", LoginNonce: nonce}
	user := &models.User{UserID: "usr-1", RoleID: "role_participant", Status: "active", IsEmailVerified: false}

	mockUserRepo.On("FindByID", mock.Anything, "usr-1").Return(user, nil)

	res, err := svc.Login(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, services.ErrEmailNotVerified, err)
}

func TestAuthService_Login_EmailNotVerified_Admin(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	svc := services.NewAuthService(nil, nil, mockUserRepo, nil, nil, nil)

	userIDs := []string{"usr-admin-1"}
	nonce, _ := utils.GetNonceStore().Generate(userIDs, time.Minute)

	req := &dto.LoginRequest{UserID: "usr-admin-1", RoleID: "role_admin", LoginNonce: nonce}
	user := &models.User{UserID: "usr-admin-1", RoleID: "role_admin", Status: "active", IsEmailVerified: false}

	mockUserRepo.On("FindByID", mock.Anything, "usr-admin-1").Return(user, nil)

	res, err := svc.Login(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, services.ErrEmailNotVerified, err)
}

func TestAuthService_Login_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockQuerier := new(MockRoleProfileQuerier)
	mockJWT := new(MockJWTService)
	mockRtRepo := new(MockRefreshTokenRepository)
	svc := services.NewAuthService(nil, mockQuerier, mockUserRepo, mockRtRepo, mockJWT, nil)

	userIDs := []string{"usr-1"}
	nonce, _ := utils.GetNonceStore().Generate(userIDs, time.Minute)

	req := &dto.LoginRequest{UserID: "usr-1", RoleID: "role_participant", LoginNonce: nonce}
	user := &models.User{UserID: "usr-1", RoleID: "role_participant", Status: "active", IsEmailVerified: true}

	mockUserRepo.On("FindByID", mock.Anything, "usr-1").Return(user, nil)
	mockQuerier.On("GetApprovalStatus", mock.Anything, "participants", "usr-1").Return("approved", nil)
	mockJWT.On("GenerateTokenPair", "usr-1", "role_participant", false).Return("access", "refresh", nil)
	mockRtRepo.On("RevokeAllForUser", mock.Anything, "usr-1").Return(nil)
	mockRtRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.RefreshToken")).Return(nil)
	mockRtRepo.On("DeleteExpired", mock.Anything).Return(nil).Maybe()

	res, err := svc.Login(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "access", res.Token)
	assert.Equal(t, "refresh", res.RefreshToken)
}

// "accepted" es el valor que se guardaba antes de unificar en "approved":
// hay perfiles reales así y tienen que poder seguir logueándose.
func TestAuthService_Login_LegacyAcceptedStatus(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockQuerier := new(MockRoleProfileQuerier)
	mockJWT := new(MockJWTService)
	mockRtRepo := new(MockRefreshTokenRepository)
	svc := services.NewAuthService(nil, mockQuerier, mockUserRepo, mockRtRepo, mockJWT, nil)

	userIDs := []string{"usr-1"}
	nonce, _ := utils.GetNonceStore().Generate(userIDs, time.Minute)

	req := &dto.LoginRequest{UserID: "usr-1", RoleID: "role_participant", LoginNonce: nonce}
	user := &models.User{UserID: "usr-1", RoleID: "role_participant", Status: "active", IsEmailVerified: true}

	mockUserRepo.On("FindByID", mock.Anything, "usr-1").Return(user, nil)
	mockQuerier.On("GetApprovalStatus", mock.Anything, "participants", "usr-1").Return("accepted", nil)
	mockJWT.On("GenerateTokenPair", "usr-1", "role_participant", false).Return("access", "refresh", nil)
	mockRtRepo.On("RevokeAllForUser", mock.Anything, "usr-1").Return(nil)
	mockRtRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.RefreshToken")).Return(nil)
	mockRtRepo.On("DeleteExpired", mock.Anything).Return(nil).Maybe()

	res, err := svc.Login(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "access", res.Token)
	assert.Equal(t, "refresh", res.RefreshToken)
}

func TestAuthService_Login_JWTError(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockQuerier := new(MockRoleProfileQuerier)
	mockJWT := new(MockJWTService)
	svc := services.NewAuthService(nil, mockQuerier, mockUserRepo, nil, mockJWT, nil)

	userIDs := []string{"usr-1"}
	nonce, _ := utils.GetNonceStore().Generate(userIDs, time.Minute)

	req := &dto.LoginRequest{UserID: "usr-1", RoleID: "role_participant", LoginNonce: nonce}
	user := &models.User{UserID: "usr-1", RoleID: "role_participant", Status: "active", IsEmailVerified: true}

	mockUserRepo.On("FindByID", mock.Anything, "usr-1").Return(user, nil)
	mockQuerier.On("GetApprovalStatus", mock.Anything, "participants", "usr-1").Return("approved", nil)
	mockJWT.On("GenerateTokenPair", "usr-1", "role_participant", false).Return("", "", errors.New("jwt error"))

	res, err := svc.Login(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "error generando token")
}

// Password Update tests
func TestAuthService_UpdateMyPassword_NoPassword(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	svc := services.NewAuthService(nil, nil, mockUserRepo, nil, nil, nil)

	user := &models.User{UserID: "usr-123", PasswordHash: ""}
	mockUserRepo.On("FindByID", mock.Anything, "usr-123").Return(user, nil)

	err := svc.UpdateMyPassword(context.Background(), "usr-123", "wrong", "newpass")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no tiene contraseña configurada")
}

func TestAuthService_UpdateMyPassword_UserNotFound(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	svc := services.NewAuthService(nil, nil, mockUserRepo, nil, nil, nil)

	mockUserRepo.On("FindByID", mock.Anything, "usr-123").Return(nil, errors.New("not found"))

	err := svc.UpdateMyPassword(context.Background(), "usr-123", "wrong", "newpass")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "usuario no encontrado")
}

func TestAuthService_LinkGoogleAccount_InvalidToken(t *testing.T) {
	mockFB := new(MockFirebaseService)
	svc := services.NewAuthService(nil, nil, nil, nil, nil, mockFB)

	mockFB.On("VerifyIDToken", mock.Anything, "invalid").Return(nil, errors.New("invalid"))

	req := &dto.LinkGoogleRequest{IDToken: "invalid"}
	err := svc.LinkGoogleAccount(context.Background(), "usr-1", req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "inválido o expiró")
}

func TestAuthService_LinkGoogleAccount_AlreadyLinked(t *testing.T) {
	mockFB := new(MockFirebaseService)
	mockUserRepo := new(MockUserRepository)
	svc := services.NewAuthService(nil, nil, mockUserRepo, nil, nil, mockFB)

	fbToken := &firebaseAuth.Token{
		UID: "google-123",
	}
	mockFB.On("VerifyIDToken", mock.Anything, "valid").Return(fbToken, nil)

	currentUser := &models.User{UserID: "usr-1", Email: "test@test.com"}
	mockUserRepo.On("FindByID", mock.Anything, "usr-1").Return(currentUser, nil)

	otherUser := &models.User{UserID: "usr-2", GoogleAuthID: "google-123", Email: "other@test.com"}
	mockUserRepo.On("FindAll", mock.Anything).Return([]models.User{*currentUser, *otherUser}, nil)

	req := &dto.LinkGoogleRequest{IDToken: "valid"}
	err := svc.LinkGoogleAccount(context.Background(), "usr-1", req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ya está vinculada a otro usuario")
}

func TestAuthService_UnlinkGoogleAccount_NoPassword(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	svc := services.NewAuthService(nil, nil, mockUserRepo, nil, nil, nil)

	currentUser := &models.User{UserID: "usr-1", Email: "test@test.com"}
	mockUserRepo.On("FindByID", mock.Anything, "usr-1").Return(currentUser, nil)

	users := []models.User{
		{UserID: "usr-1", Email: "test@test.com", PasswordHash: ""},
	}
	mockUserRepo.On("FindAllByEmail", mock.Anything, "test@test.com").Return(users, nil)

	err := svc.UnlinkGoogleAccount(context.Background(), "usr-1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no tenés una contraseña configurada")
}

func TestAuthService_CheckCredentials_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	svc := services.NewAuthService(nil, nil, mockUserRepo, nil, nil, nil)

	req := &dto.CheckCredentialsRequest{Email: "test@example.com", Password: "RightPassword123!"}
	pwdHash, _ := utils.HashPassword("RightPassword123!")

	users := []models.User{{UserID: "1", Email: "test@example.com", PasswordHash: pwdHash, Status: "Active", IsEmailVerified: true}}
	mockUserRepo.On("FindAllByEmail", mock.Anything, "test@example.com").Return(users, nil)

	res, err := svc.CheckCredentials(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.True(t, res.IsEmailVerified)
	assert.Equal(t, "test@example.com", res.Email)
}

func TestAuthService_UpdateMyPassword_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	svc := services.NewAuthService(nil, nil, mockUserRepo, nil, nil, nil)

	pwdHash, _ := utils.HashPassword("OldPassword123!")
	user := &models.User{UserID: "u1", PasswordHash: pwdHash}
	mockUserRepo.On("FindByID", mock.Anything, "u1").Return(user, nil)
	mockUserRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)

	err := svc.UpdateMyPassword(context.Background(), "u1", "OldPassword123!", "NewPassword123!")
	assert.NoError(t, err)
}

func TestAuthService_LinkGoogleAccount_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockFirebase := new(MockFirebaseService)
	svc := services.NewAuthService(nil, nil, mockUserRepo, nil, nil, mockFirebase)

	req := &dto.LinkGoogleRequest{IDToken: "valid_token", GoogleID: "g1"}
	fbToken := &firebaseAuth.Token{UID: "g1"}

	mockFirebase.On("VerifyIDToken", mock.Anything, "valid_token").Return(fbToken, nil)

	currentUser := &models.User{UserID: "u1", Email: "test@example.com"}
	mockUserRepo.On("FindByID", mock.Anything, "u1").Return(currentUser, nil)
	mockUserRepo.On("FindAll", mock.Anything).Return([]models.User{}, nil)
	mockUserRepo.On("FindAllByEmail", mock.Anything, "test@example.com").Return([]models.User{}, nil)
	mockUserRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)

	err := svc.LinkGoogleAccount(context.Background(), "u1", req)
	assert.NoError(t, err)
}

func TestAuthService_UnlinkGoogleAccount_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	svc := services.NewAuthService(nil, nil, mockUserRepo, nil, nil, nil)

	pwdHash, _ := utils.HashPassword("HasPassword123!")
	user := &models.User{UserID: "u1", Email: "test@example.com", PasswordHash: pwdHash, GoogleAuthID: "g1"}

	mockUserRepo.On("FindByID", mock.Anything, "u1").Return(user, nil)
	mockUserRepo.On("FindAllByEmail", mock.Anything, "test@example.com").Return([]models.User{*user}, nil)
	mockUserRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)

	err := svc.UnlinkGoogleAccount(context.Background(), "u1")
	assert.NoError(t, err)
}

func TestAuthService_UpdateMyProfile_Volunteer(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockQuerier := new(MockRoleProfileQuerier)
	svc := services.NewAuthService(nil, mockQuerier, mockUserRepo, nil, nil, nil)

	user := &models.User{UserID: "vol-1", Email: "old@test.com", Phone: "111"}
	mockUserRepo.On("FindByID", mock.Anything, "vol-1").Return(user, nil)
	mockUserRepo.On("Update", mock.Anything, mock.MatchedBy(func(u *models.User) bool {
		return u.Email == "new@test.com" && u.Phone == "222"
	})).Return(nil)

	// Volunteer sending email, phone and image_url (which is not a volunteer column)
	data := map[string]interface{}{
		"email":     "new@test.com",
		"phone":     "222",
		"image_url": "",
	}

	err := svc.UpdateMyProfile(context.Background(), "role_volunteer", "vol-1", data)
	assert.NoError(t, err)
	mockUserRepo.AssertExpectations(t)
	// UpdateRoleData should NOT be called because volunteers table has no extra columns
	mockQuerier.AssertNotCalled(t, "UpdateRoleData", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestAuthService_UpdateMyProfile_Company_StripsImageUrl(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockQuerier := new(MockRoleProfileQuerier)
	svc := services.NewAuthService(nil, mockQuerier, mockUserRepo, nil, nil, nil)

	user := &models.User{UserID: "comp-1", Email: "empresa@test.com"}
	mockUserRepo.On("FindByID", mock.Anything, "comp-1").Return(user, nil)
	mockUserRepo.On("Update", mock.Anything, mock.MatchedBy(func(u *models.User) bool {
		return u.Email == "new@acme.com"
	})).Return(nil)

	// Company sending email, business_name, cuit, and image_url from form
	data := map[string]interface{}{
		"email":         "new@acme.com",
		"business_name": "Acme Inc",
		"cuit":          "30-12345678-9",
		"image_url":     "",
	}

	mockQuerier.On("UpdateRoleData", mock.Anything, "companies", "comp-1", map[string]interface{}{
		"business_name": "Acme Inc",
		"cuit":          "30-12345678-9",
	}).Return(nil)

	err := svc.UpdateMyProfile(context.Background(), "role_company", "comp-1", data)
	assert.NoError(t, err)
	mockUserRepo.AssertExpectations(t)
	mockQuerier.AssertExpectations(t)
}

func TestAuthService_UpdateMyProfile_Participant_AllowsImageUrlAndStripsPrivilegedFields(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockQuerier := new(MockRoleProfileQuerier)
	svc := services.NewAuthService(nil, mockQuerier, mockUserRepo, nil, nil, nil)

	user := &models.User{UserID: "part-1", Email: "part@test.com"}
	mockUserRepo.On("FindByID", mock.Anything, "part-1").Return(user, nil)

	data := map[string]interface{}{
		"address":         "Calle Falsa 123",
		"image_url":       "https://example.com/avatar.jpg",
		"approval_status": "approved", // malicious attempt
	}

	mockQuerier.On("UpdateRoleData", mock.Anything, "participants", "part-1", map[string]interface{}{
		"address":   "Calle Falsa 123",
		"image_url": "https://example.com/avatar.jpg",
	}).Return(nil)

	err := svc.UpdateMyProfile(context.Background(), "role_participant", "part-1", data)
	assert.NoError(t, err)
	mockQuerier.AssertExpectations(t)
}

