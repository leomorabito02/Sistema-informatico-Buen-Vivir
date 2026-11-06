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
)

type mockPasswordRecoveryRepo struct {
	records        map[string]*models.PasswordRecovery
	createErr      error
	findByHashErr  error
	markAsUsedErr  error
	deleteByUIDErr error
}

func newMockPasswordRecoveryRepo() *mockPasswordRecoveryRepo {
	return &mockPasswordRecoveryRepo{
		records: make(map[string]*models.PasswordRecovery),
	}
}

func (m *mockPasswordRecoveryRepo) Create(ctx context.Context, v *models.PasswordRecovery) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.records[v.TokenHash] = v
	return nil
}

func (m *mockPasswordRecoveryRepo) FindByTokenHash(ctx context.Context, tokenHash string) (*models.PasswordRecovery, error) {
	if m.findByHashErr != nil {
		return nil, m.findByHashErr
	}
	v, ok := m.records[tokenHash]
	if !ok {
		return nil, errors.New("token no encontrado")
	}
	return v, nil
}

func (m *mockPasswordRecoveryRepo) MarkAsUsed(ctx context.Context, id string) error {
	if m.markAsUsedErr != nil {
		return m.markAsUsedErr
	}
	now := time.Now()
	for _, v := range m.records {
		if v.PasswordRecoveryID == id {
			v.UsedAt = &now
			return nil
		}
	}
	return errors.New("not found")
}

func (m *mockPasswordRecoveryRepo) DeleteByUserID(ctx context.Context, userID string) error {
	if m.deleteByUIDErr != nil {
		return m.deleteByUIDErr
	}
	for k, v := range m.records {
		if v.UserID == userID {
			delete(m.records, k)
		}
	}
	return nil
}

func (m *mockPasswordRecoveryRepo) DeleteExpired(ctx context.Context) error {
	return nil
}

type mockRefreshTokenRepoForRecovery struct {
	revokedUserIDs []string
}

func (m *mockRefreshTokenRepoForRecovery) Create(ctx context.Context, rt *models.RefreshToken) error {
	return nil
}
func (m *mockRefreshTokenRepoForRecovery) FindByTokenHash(ctx context.Context, tokenHash string) (*models.RefreshToken, error) {
	return nil, nil
}
func (m *mockRefreshTokenRepoForRecovery) Revoke(ctx context.Context, id string) error {
	return nil
}
func (m *mockRefreshTokenRepoForRecovery) RevokeAllForUser(ctx context.Context, userID string) error {
	m.revokedUserIDs = append(m.revokedUserIDs, userID)
	return nil
}
func (m *mockRefreshTokenRepoForRecovery) DeleteExpired(ctx context.Context) error {
	return nil
}

func TestPasswordRecoveryService_RequestReset_EmptyEmail(t *testing.T) {
	// Arrange
	recoveryRepo := newMockPasswordRecoveryRepo()
	userRepo := newMockUserRepoForVerification()
	rtRepo := &mockRefreshTokenRepoForRecovery{}
	notifSync := &mockNotificationSync{}
	svc := services.NewPasswordRecoveryService(recoveryRepo, userRepo, rtRepo, notifSync)

	// Act
	err := svc.RequestReset(context.Background(), "")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "correo electrónico es obligatorio")
}

func TestPasswordRecoveryService_RequestReset_NonExistentEmail(t *testing.T) {
	// Arrange
	recoveryRepo := newMockPasswordRecoveryRepo()
	userRepo := newMockUserRepoForVerification()
	rtRepo := &mockRefreshTokenRepoForRecovery{}
	notifSync := &mockNotificationSync{}
	svc := services.NewPasswordRecoveryService(recoveryRepo, userRepo, rtRepo, notifSync)

	// Act
	err := svc.RequestReset(context.Background(), "unknown@example.com")

	// Assert: Anti-enumeration should return nil without failing
	assert.NoError(t, err)
	assert.Empty(t, recoveryRepo.records)
}

func TestPasswordRecoveryService_RequestReset_Success(t *testing.T) {
	// Arrange
	recoveryRepo := newMockPasswordRecoveryRepo()
	userRepo := newMockUserRepoForVerification()
	userRepo.users["usr-1"] = &models.User{
		UserID:    "usr-1",
		Email:     "user@example.com",
		FirstName: "Juan",
		LastName:  "Perez",
	}
	rtRepo := &mockRefreshTokenRepoForRecovery{}
	notifSync := &mockNotificationSync{}
	producer := &mockUsuariosProducer{}
	svc := services.NewPasswordRecoveryService(recoveryRepo, userRepo, rtRepo, notifSync, producer)

	// Act
	err := svc.RequestReset(context.Background(), "user@example.com")

	// Assert
	assert.NoError(t, err)
	assert.Len(t, recoveryRepo.records, 0)
	assert.Len(t, producer.emailSolicitadoCalls, 1)
	assert.Equal(t, "usr-1", producer.emailSolicitadoCalls[0].UserID)
	assert.Equal(t, "recuperacion_password", producer.emailSolicitadoCalls[0].Tipo)
	assert.NotEmpty(t, producer.emailSolicitadoCalls[0].RequestID)
}

func TestPasswordRecoveryService_IssueActionLink(t *testing.T) {
	// Arrange
	recoveryRepo := newMockPasswordRecoveryRepo()
	userRepo := newMockUserRepoForVerification()
	rtRepo := &mockRefreshTokenRepoForRecovery{}
	notifSync := &mockNotificationSync{}
	svc := services.NewPasswordRecoveryService(recoveryRepo, userRepo, rtRepo, notifSync)

	// Act
	link, err := svc.IssueActionLink(context.Background(), "pr-req-1", "usr-1")

	// Assert
	assert.NoError(t, err)
	assert.Contains(t, link, "token=")
	assert.Len(t, recoveryRepo.records, 1)

	var savedPR *models.PasswordRecovery
	for _, v := range recoveryRepo.records {
		savedPR = v
		break
	}
	assert.NotNil(t, savedPR)
	assert.Equal(t, "pr-req-1", savedPR.PasswordRecoveryID)
	assert.Equal(t, "usr-1", savedPR.UserID)
	assert.NotEmpty(t, savedPR.TokenHash)
}

func TestPasswordRecoveryService_ResetPassword_EmptyToken(t *testing.T) {
	// Arrange
	recoveryRepo := newMockPasswordRecoveryRepo()
	userRepo := newMockUserRepoForVerification()
	rtRepo := &mockRefreshTokenRepoForRecovery{}
	notifSync := &mockNotificationSync{}
	svc := services.NewPasswordRecoveryService(recoveryRepo, userRepo, rtRepo, notifSync)

	// Act
	err := svc.ResetPassword(context.Background(), "", "NewValidPass123!")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token de recuperación es obligatorio")
}

func TestPasswordRecoveryService_ResetPassword_WeakPassword(t *testing.T) {
	// Arrange
	recoveryRepo := newMockPasswordRecoveryRepo()
	userRepo := newMockUserRepoForVerification()
	rtRepo := &mockRefreshTokenRepoForRecovery{}
	notifSync := &mockNotificationSync{}
	svc := services.NewPasswordRecoveryService(recoveryRepo, userRepo, rtRepo, notifSync)

	// Act
	err := svc.ResetPassword(context.Background(), "token123", "weak")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "al menos 8 caracteres")
}

func TestPasswordRecoveryService_ResetPassword_TokenNotFound(t *testing.T) {
	// Arrange
	recoveryRepo := newMockPasswordRecoveryRepo()
	userRepo := newMockUserRepoForVerification()
	rtRepo := &mockRefreshTokenRepoForRecovery{}
	notifSync := &mockNotificationSync{}
	svc := services.NewPasswordRecoveryService(recoveryRepo, userRepo, rtRepo, notifSync)

	// Act
	err := svc.ResetPassword(context.Background(), "invalid-token", "ValidPassword123!")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "inválido, ya fue utilizado o ha expirado")
}

func TestPasswordRecoveryService_ResetPassword_ExpiredToken(t *testing.T) {
	// Arrange
	recoveryRepo := newMockPasswordRecoveryRepo()
	rawToken := "expired-token"
	tokenHash := utils.HashToken(rawToken)
	recoveryRepo.records[tokenHash] = &models.PasswordRecovery{
		PasswordRecoveryID: "pr-1",
		UserID:             "usr-1",
		TokenHash:          tokenHash,
		ExpiresAt:          time.Now().Add(-1 * time.Hour),
	}

	userRepo := newMockUserRepoForVerification()
	rtRepo := &mockRefreshTokenRepoForRecovery{}
	notifSync := &mockNotificationSync{}
	svc := services.NewPasswordRecoveryService(recoveryRepo, userRepo, rtRepo, notifSync)

	// Act
	err := svc.ResetPassword(context.Background(), rawToken, "ValidPassword123!")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ha expirado")
}

func TestPasswordRecoveryService_ResetPassword_UsedToken(t *testing.T) {
	// Arrange
	recoveryRepo := newMockPasswordRecoveryRepo()
	rawToken := "used-token"
	tokenHash := utils.HashToken(rawToken)
	usedTime := time.Now().Add(-10 * time.Minute)
	recoveryRepo.records[tokenHash] = &models.PasswordRecovery{
		PasswordRecoveryID: "pr-1",
		UserID:             "usr-1",
		TokenHash:          tokenHash,
		ExpiresAt:          time.Now().Add(20 * time.Minute),
		UsedAt:             &usedTime,
	}

	userRepo := newMockUserRepoForVerification()
	rtRepo := &mockRefreshTokenRepoForRecovery{}
	notifSync := &mockNotificationSync{}
	svc := services.NewPasswordRecoveryService(recoveryRepo, userRepo, rtRepo, notifSync)

	// Act
	err := svc.ResetPassword(context.Background(), rawToken, "ValidPassword123!")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ya fue utilizado")
}

func TestPasswordRecoveryService_ResetPassword_Success(t *testing.T) {
	// Arrange
	recoveryRepo := newMockPasswordRecoveryRepo()
	rawToken := "valid-token-12345"
	tokenHash := utils.HashToken(rawToken)
	recoveryRepo.records[tokenHash] = &models.PasswordRecovery{
		PasswordRecoveryID: "pr-100",
		UserID:             "usr-1",
		TokenHash:          tokenHash,
		ExpiresAt:          time.Now().Add(20 * time.Minute),
	}

	userRepo := newMockUserRepoForVerification()
	userRepo.users["usr-1"] = &models.User{
		UserID:          "usr-1",
		Email:           "multi@example.com",
		RoleID:          "role_participant",
		IsEmailVerified: false,
	}
	userRepo.users["usr-2"] = &models.User{
		UserID:          "usr-2",
		Email:           "multi@example.com",
		RoleID:          "role_volunteer",
		IsEmailVerified: false,
	}

	rtRepo := &mockRefreshTokenRepoForRecovery{}
	notifSync := &mockNotificationSync{}
	svc := services.NewPasswordRecoveryService(recoveryRepo, userRepo, rtRepo, notifSync)

	// Act
	err := svc.ResetPassword(context.Background(), rawToken, "NewSecurePassword123!")

	// Assert
	assert.NoError(t, err)
	// Check token marked as used
	assert.NotNil(t, recoveryRepo.records[tokenHash].UsedAt)
	// Check user accounts updated
	for _, u := range userRepo.users {
		assert.NotEmpty(t, u.PasswordHash)
		assert.True(t, utils.CheckPasswordHash("NewSecurePassword123!", u.PasswordHash))
		assert.True(t, u.IsEmailVerified)
	}
	// Check all refresh tokens revoked for both roles
	assert.Contains(t, rtRepo.revokedUserIDs, "usr-1")
	assert.Contains(t, rtRepo.revokedUserIDs, "usr-2")
}
