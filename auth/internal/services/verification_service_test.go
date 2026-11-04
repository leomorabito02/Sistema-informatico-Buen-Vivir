package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"auth/internal/models"
	"auth/internal/repositories"
	"auth/internal/services"

	"github.com/stretchr/testify/assert"
)

type mockEmailVerificationRepo struct {
	verifications map[string]*models.EmailVerification
	createErr     error
	findByHashErr error
	markAsUsedErr error
}

func newMockEmailVerificationRepo() *mockEmailVerificationRepo {
	return &mockEmailVerificationRepo{
		verifications: make(map[string]*models.EmailVerification),
	}
}

func (m *mockEmailVerificationRepo) Create(ctx context.Context, v *models.EmailVerification) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.verifications[v.TokenHash] = v
	return nil
}

func (m *mockEmailVerificationRepo) FindByTokenHash(ctx context.Context, tokenHash string) (*models.EmailVerification, error) {
	if m.findByHashErr != nil {
		return nil, m.findByHashErr
	}
	v, ok := m.verifications[tokenHash]
	if !ok {
		return nil, errors.New("token no encontrado")
	}
	return v, nil
}

func (m *mockEmailVerificationRepo) MarkAsUsed(ctx context.Context, id string) error {
	if m.markAsUsedErr != nil {
		return m.markAsUsedErr
	}
	for _, v := range m.verifications {
		if v.EmailVerificationID == id {
			now := time.Now()
			v.UsedAt = &now
			return nil
		}
	}
	return errors.New("not found")
}

func (m *mockEmailVerificationRepo) Delete(ctx context.Context, id string) error {
	for k, v := range m.verifications {
		if v.EmailVerificationID == id {
			delete(m.verifications, k)
			return nil
		}
	}
	return errors.New("not found")
}

func (m *mockEmailVerificationRepo) DeleteExpired(ctx context.Context) error {
	return nil
}

type mockUserRepoForVerification struct {
	users                 map[string]*models.User
	updateEmailVerifiedErr error
}

func newMockUserRepoForVerification() *mockUserRepoForVerification {
	return &mockUserRepoForVerification{
		users: make(map[string]*models.User),
	}
}

func (m *mockUserRepoForVerification) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, errors.New("usuario no encontrado")
}

func (m *mockUserRepoForVerification) FindByID(ctx context.Context, id string) (*models.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, errors.New("usuario no encontrado")
	}
	return u, nil
}

func (m *mockUserRepoForVerification) UpdateEmailVerified(ctx context.Context, userID string, verified bool) error {
	if m.updateEmailVerifiedErr != nil {
		return m.updateEmailVerifiedErr
	}
	if u, ok := m.users[userID]; ok {
		u.IsEmailVerified = verified
		return nil
	}
	return errors.New("usuario no encontrado")
}

func (m *mockUserRepoForVerification) UpdateEmailVerifiedByEmail(ctx context.Context, email string, verified bool) error {
	if m.updateEmailVerifiedErr != nil {
		return m.updateEmailVerifiedErr
	}
	found := false
	for _, u := range m.users {
		if u.Email == email {
			u.IsEmailVerified = verified
			found = true
		}
	}
	if found {
		return nil
	}
	return errors.New("usuario no encontrado")
}

func (m *mockUserRepoForVerification) UpdatePasswordHashByEmail(ctx context.Context, email string, passwordHash string) error {
	found := false
	for _, u := range m.users {
		if u.Email == email {
			u.PasswordHash = passwordHash
			found = true
		}
	}
	if found {
		return nil
	}
	return errors.New("usuario no encontrado")
}

func (m *mockUserRepoForVerification) UpdatePasswordAndVerifiedByEmail(ctx context.Context, email string, passwordHash string, verified bool) error {
	found := false
	for _, u := range m.users {
		if u.Email == email {
			u.PasswordHash = passwordHash
			u.IsEmailVerified = verified
			found = true
		}
	}
	if found {
		return nil
	}
	return errors.New("usuario no encontrado")
}

func (m *mockUserRepoForVerification) FindAllByEmail(ctx context.Context, email string) ([]models.User, error) {
	var res []models.User
	for _, u := range m.users {
		if u.Email == email {
			res = append(res, *u)
		}
	}
	if len(res) == 0 {
		return nil, errors.New("usuario no encontrado")
	}
	return res, nil
}
func (m *mockUserRepoForVerification) FindAllByGoogleID(ctx context.Context, googleID string) ([]models.User, error) {
	return nil, nil
}
func (m *mockUserRepoForVerification) FindAll(ctx context.Context) ([]models.User, error) {
	return nil, nil
}
func (m *mockUserRepoForVerification) Create(ctx context.Context, user *models.User) error {
	m.users[user.UserID] = user
	return nil
}
func (m *mockUserRepoForVerification) CreateWithRoleProfile(ctx context.Context, user *models.User, roleID string, approvedBy string) error {
	return nil
}
func (m *mockUserRepoForVerification) Update(ctx context.Context, user *models.User) error {
	return nil
}
func (m *mockUserRepoForVerification) Delete(ctx context.Context, id string) error {
	return nil
}
func (m *mockUserRepoForVerification) CountActiveParticipants(ctx context.Context) (int64, error) {
	return 0, nil
}
func (m *mockUserRepoForVerification) CountVolunteers(ctx context.Context) (int64, error) {
	return 0, nil
}
func (m *mockUserRepoForVerification) FindParticipantsHistory(ctx context.Context) ([]models.ParticipantHistory, error) {
	return nil, nil
}
func (m *mockUserRepoForVerification) FindCampaignEligibleUsers(ctx context.Context) ([]repositories.CampaignEligibleUser, error) {
	return nil, nil
}
func (m *mockUserRepoForVerification) FindActiveUserIDsByRoles(ctx context.Context, roles []string) ([]string, error) {
	return nil, nil
}

type mockNotificationSync struct{}

func (m *mockNotificationSync) InitUserPreferencesAsync(userID string) {}

func TestVerificationService_GenerateAndSendVerification(t *testing.T) {
	// Arrange
	evRepo := newMockEmailVerificationRepo()
	userRepo := newMockUserRepoForVerification()
	notifSync := &mockNotificationSync{}
	svc := services.NewVerificationService(evRepo, userRepo, notifSync)

	user := &models.User{
		UserID:    "usr-1",
		FirstName: "Juan",
		LastName:  "Perez",
		Email:     "juan@example.com",
	}

	// Act
	err := svc.GenerateAndSendVerification(context.Background(), user)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 1, len(evRepo.verifications))
}

func TestVerificationService_VerifyEmail_Success(t *testing.T) {
	// Arrange
	evRepo := newMockEmailVerificationRepo()
	userRepo := newMockUserRepoForVerification()
	notifSync := &mockNotificationSync{}
	svc := services.NewVerificationService(evRepo, userRepo, notifSync)

	user := &models.User{
		UserID:          "usr-1",
		Email:           "juan@example.com",
		IsEmailVerified: false,
	}
	_ = userRepo.Create(context.Background(), user)
	_ = svc.GenerateAndSendVerification(context.Background(), user)

	// Verify with invalid token
	errInvalid := svc.VerifyEmail(context.Background(), "invalid-token")
	assert.Error(t, errInvalid)

	_ = svc.VerifyEmail(context.Background(), "") // should fail empty
	assert.False(t, user.IsEmailVerified)

	// Inject manual record
	manualHash := "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae" // sha256 of "foo"
	evRepo.verifications[manualHash] = &models.EmailVerification{
		EmailVerificationID: "ev-manual",
		UserID:              "usr-1",
		TokenHash:           manualHash,
		ExpiresAt:           time.Now().Add(1 * time.Hour),
	}

	// Act
	err := svc.VerifyEmail(context.Background(), "foo")

	// Assert
	assert.NoError(t, err)
	assert.True(t, user.IsEmailVerified)

	// Re-verifying deleted token should fail because temporal link is destroyed
	errReused := svc.VerifyEmail(context.Background(), "foo")
	assert.Error(t, errReused)
	assert.Contains(t, errReused.Error(), "no existe")
	assert.Nil(t, evRepo.verifications[manualHash])
}

func TestVerificationService_ResendVerification(t *testing.T) {
	// Arrange
	evRepo := newMockEmailVerificationRepo()
	userRepo := newMockUserRepoForVerification()
	notifSync := &mockNotificationSync{}
	svc := services.NewVerificationService(evRepo, userRepo, notifSync)

	user := &models.User{
		UserID:          "usr-1",
		Email:           "juan@example.com",
		IsEmailVerified: false,
	}
	_ = userRepo.Create(context.Background(), user)

	// Act
	err := svc.ResendVerification(context.Background(), "juan@example.com")

	// Assert
	assert.NoError(t, err)

	// Resend for already verified user should fail
	user.IsEmailVerified = true
	errVerified := svc.ResendVerification(context.Background(), "juan@example.com")
	assert.Error(t, errVerified)
	assert.Contains(t, errVerified.Error(), "ya ha sido verificado")

	// Resend for nonexistent user should fail
	errNotFound := svc.ResendVerification(context.Background(), "nonexistent@example.com")
	assert.Error(t, errNotFound)
}

func TestVerificationService_VerifyEmail_MultiAccount(t *testing.T) {
	// Arrange
	evRepo := newMockEmailVerificationRepo()
	userRepo := newMockUserRepoForVerification()
	notifSync := &mockNotificationSync{}
	svc := services.NewVerificationService(evRepo, userRepo, notifSync)

	u1 := &models.User{UserID: "usr-1", Email: "multi@example.com", IsEmailVerified: false}
	u2 := &models.User{UserID: "usr-2", Email: "multi@example.com", IsEmailVerified: false}
	_ = userRepo.Create(context.Background(), u1)
	_ = userRepo.Create(context.Background(), u2)

	manualHash := "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae" // sha256 of "foo"
	evRepo.verifications[manualHash] = &models.EmailVerification{
		EmailVerificationID: "ev-multi",
		UserID:              "usr-1",
		TokenHash:           manualHash,
		ExpiresAt:           time.Now().Add(1 * time.Hour),
	}

	// Act
	err := svc.VerifyEmail(context.Background(), "foo")

	// Assert
	assert.NoError(t, err)
	assert.True(t, u1.IsEmailVerified)
	assert.True(t, u2.IsEmailVerified)
}
