package services_test

import (
	"os"
	"os/exec"
	"testing"
	"time"

	"buenvivir/education/internal/services"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NewJWTService lee JWT_SECRET del entorno y llama a os.Exit(1) si falta
// o es un default inseguro conocido, así que ese camino se prueba en un
// subproceso (mismo patrón que usa la librería estándar de Go para
// probar rutas que llaman a os.Exit) en vez de tumbar el proceso de test.
func TestNewJWTService_MissingSecret_Exits(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		_ = os.Unsetenv("JWT_SECRET")
		services.NewJWTService()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestNewJWTService_MissingSecret_Exits")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()

	exitErr, ok := err.(*exec.ExitError)
	require.True(t, ok, "se esperaba que el proceso saliera con os.Exit(1)")
	assert.False(t, exitErr.Success())
}

func TestNewJWTService_InsecureDefault_Exits(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		_ = os.Setenv("JWT_SECRET", "super-secret-access-key")
		services.NewJWTService()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestNewJWTService_InsecureDefault_Exits")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()

	exitErr, ok := err.(*exec.ExitError)
	require.True(t, ok, "se esperaba que el proceso saliera con os.Exit(1)")
	assert.False(t, exitErr.Success())
}

func TestJWTService_ValidateToken_ValidAccessToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "una-clave-larga-y-segura-de-32-bytes")
	svc := services.NewJWTService()

	claims := jwt.MapClaims{
		"sub":  "usr-1",
		"role": "role_teacher",
		"type": "access",
		"exp":  time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("una-clave-larga-y-segura-de-32-bytes"))
	require.NoError(t, err)

	got, err := svc.ValidateToken(tokenString)
	require.NoError(t, err)
	assert.Equal(t, "usr-1", got["sub"])
	assert.Equal(t, "role_teacher", got["role"])
}

func TestJWTService_ValidateToken_WrongSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "una-clave-larga-y-segura-de-32-bytes")
	svc := services.NewJWTService()

	claims := jwt.MapClaims{"sub": "usr-1", "type": "access", "exp": time.Now().Add(time.Hour).Unix()}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("otra-clave-distinta-de-32-bytes"))
	require.NoError(t, err)

	_, err = svc.ValidateToken(tokenString)
	assert.Error(t, err)
}

func TestJWTService_ValidateToken_InvalidType(t *testing.T) {
	t.Setenv("JWT_SECRET", "una-clave-larga-y-segura-de-32-bytes")
	svc := services.NewJWTService()

	// education-service nunca emite refresh tokens, pero debe rechazar
	// uno igual que rechazaría cualquier claims["type"] != "access".
	claims := jwt.MapClaims{"sub": "usr-1", "type": "refresh", "exp": time.Now().Add(time.Hour).Unix()}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("una-clave-larga-y-segura-de-32-bytes"))
	require.NoError(t, err)

	_, err = svc.ValidateToken(tokenString)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid token type")
}

func TestJWTService_ValidateToken_InvalidSigningMethod(t *testing.T) {
	t.Setenv("JWT_SECRET", "una-clave-larga-y-segura-de-32-bytes")
	svc := services.NewJWTService()

	token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{"sub": "usr-1", "type": "access"})
	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	_, err = svc.ValidateToken(tokenString)
	assert.Error(t, err)
}

func TestJWTService_ValidateToken_Expired(t *testing.T) {
	t.Setenv("JWT_SECRET", "una-clave-larga-y-segura-de-32-bytes")
	svc := services.NewJWTService()

	claims := jwt.MapClaims{
		"sub": "usr-1", "type": "access",
		"exp": time.Now().Add(-1 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("una-clave-larga-y-segura-de-32-bytes"))
	require.NoError(t, err)

	_, err = svc.ValidateToken(tokenString)
	assert.Error(t, err)
}
