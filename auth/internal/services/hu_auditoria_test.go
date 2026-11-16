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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// fakeAttempts es un LoginAttemptRepository en memoria con la misma
// regla que el de GORM.
type fakeAttempts struct{ rows map[string]*models.LoginAttempt }

func newFakeAttempts() *fakeAttempts { return &fakeAttempts{rows: map[string]*models.LoginAttempt{}} }

func (f *fakeAttempts) Get(_ context.Context, key string) (*models.LoginAttempt, error) {
	if la, ok := f.rows[key]; ok {
		c := *la
		return &c, nil
	}
	return nil, nil
}

func (f *fakeAttempts) RegisterFailure(_ context.Context, key string, max int, lock time.Duration) (*models.LoginAttempt, error) {
	la, ok := f.rows[key]
	if !ok {
		la = &models.LoginAttempt{EmailHash: key}
		f.rows[key] = la
	}
	if la.LockedUntil != nil && !la.LockedUntil.After(time.Now()) {
		la.LockedUntil = nil
	}
	la.Failed++
	if la.Failed >= max {
		u := time.Now().Add(lock)
		la.LockedUntil = &u
		la.Failed = 0
	}
	c := *la
	return &c, nil
}

func (f *fakeAttempts) Reset(_ context.Context, key string) error {
	delete(f.rows, key)
	return nil
}

func lockFixture(t *testing.T) (services.AuthService, *fakeAttempts) {
	t.Helper()
	hash, err := utils.HashPassword("Correcta123!")
	require.NoError(t, err)
	users := new(MockUserRepository)
	users.On("FindAllByEmail", mock.Anything, "ana@gmail.com").Return([]models.User{{UserID: "usr-1", Email: "ana@gmail.com", PasswordHash: hash, RoleID: "role_admin", IsEmailVerified: true, Status: "active"}}, nil)
	attempts := newFakeAttempts()
	svc := services.NewAuthService(nil, new(MockRoleProfileQuerier), users, new(MockRefreshTokenRepository), new(MockJWTService), nil, attempts)
	return svc, attempts
}

// HU-02: al 5.º intento fallido la cuenta queda bloqueada, incluso con la
// contraseña correcta, y el error lo reconoce el controlador.
func TestCheckCredentials_LockAfterFiveFailures(t *testing.T) {
	svc, _ := lockFixture(t)
	ctx := context.Background()
	for i := 1; i <= 4; i++ {
		_, err := svc.CheckCredentials(ctx, &dto.CheckCredentialsRequest{Email: "ana@gmail.com", Password: "mal"})
		require.Error(t, err)
		assert.False(t, errors.Is(err, services.ErrLoginLocked), "intento %d no debería bloquear", i)
		assert.Contains(t, err.Error(), "Te quedan")
	}
	_, err := svc.CheckCredentials(ctx, &dto.CheckCredentialsRequest{Email: "ana@gmail.com", Password: "mal"})
	assert.ErrorIs(t, err, services.ErrLoginLocked)

	_, err = svc.CheckCredentials(ctx, &dto.CheckCredentialsRequest{Email: "ana@gmail.com", Password: "Correcta123!"})
	assert.ErrorIs(t, err, services.ErrLoginLocked, "bloqueada aunque la contraseña sea correcta")
}

// HU-02: el bloqueo vence solo y un login correcto reinicia el contador.
func TestCheckCredentials_LockExpiresAndSuccessResets(t *testing.T) {
	svc, attempts := lockFixture(t)
	ctx := context.Background()
	key := utils.HashToken("ana@gmail.com")
	past := time.Now().Add(-time.Second)
	attempts.rows[key] = &models.LoginAttempt{EmailHash: key, LockedUntil: &past}

	resp, err := svc.CheckCredentials(ctx, &dto.CheckCredentialsRequest{Email: "ana@gmail.com", Password: "Correcta123!"})
	require.NoError(t, err)
	assert.NotEmpty(t, resp.LoginNonce)
	_, quedan := attempts.rows[key]
	assert.False(t, quedan, "el login correcto borra los fallos")
}

// HU-01 (decisión 15/11): el registro de participante exige Gmail; el de
// empresa no.
func TestRegister_ParticipantRequiresGmail(t *testing.T) {
	users := new(MockUserRepository)
	users.On("FindByEmail", mock.Anything, mock.Anything).Return(nil, errors.New("not found"))
	tx := new(MockTransactor)
	tx.On("RunInTx", mock.Anything, mock.Anything).Return(nil)
	svc := services.NewAuthService(tx, new(MockRoleProfileQuerier), users, new(MockRefreshTokenRepository), new(MockJWTService), nil)

	_, err := svc.Register(context.Background(), &dto.RegisterRequest{FirstName: "Ana", LastName: "Paz", Email: "ana@hotmail.com", Phone: "351", Password: "Password123!", Role: "participant"})
	assert.ErrorIs(t, err, services.ErrParticipantGmail)

	_, err = svc.Register(context.Background(), &dto.RegisterRequest{FirstName: "Emp", LastName: "SA", Email: "rrhh@empresa.com.ar", Phone: "351", Password: "Password123!", Role: "company"})
	assert.NoError(t, err)
}
