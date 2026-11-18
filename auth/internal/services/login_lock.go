package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"auth/internal/utils"
)

// LoginLockedError se devuelve cuando la cuenta está bloqueada por
// intentos fallidos (HU-02). El controlador responde 423 con este
// mensaje y el código ACCOUNT_LOCKED.
type LoginLockedError struct {
	Until time.Time
}

func (e *LoginLockedError) Error() string {
	min := int(math.Ceil(time.Until(e.Until).Minutes()))
	if min < 1 {
		min = 1
	}
	return fmt.Sprintf("Tu cuenta quedó bloqueada por demasiados intentos fallidos. Probá de nuevo en %d minuto(s) o recuperá tu contraseña.", min)
}

// ErrLoginLocked permite usar errors.Is con un *LoginLockedError.
var ErrLoginLocked = errors.New("ACCOUNT_LOCKED")

func (e *LoginLockedError) Is(target error) bool { return target == ErrLoginLocked }

// loginMaxAttempts y loginLockDuration salen de LOGIN_MAX_ATTEMPTS y
// LOGIN_LOCK_MINUTES (5 intentos y 15 minutos por defecto).
func loginMaxAttempts() int {
	if n, err := strconv.Atoi(os.Getenv("LOGIN_MAX_ATTEMPTS")); err == nil && n > 0 {
		return n
	}
	return 5
}

func loginLockDuration() time.Duration {
	if n, err := strconv.Atoi(os.Getenv("LOGIN_LOCK_MINUTES")); err == nil && n > 0 {
		return time.Duration(n) * time.Minute
	}
	return 15 * time.Minute
}

func loginAttemptKey(email string) string {
	return utils.HashToken(strings.ToLower(strings.TrimSpace(email)))
}

// checkLoginLock falla si la clave está bloqueada. Si la base no
// responde, no bloquea (el rate limit por IP sigue activo).
func (s *authService) checkLoginLock(ctx context.Context, key string) error {
	if s.loginAttempts == nil {
		return nil
	}
	la, err := s.loginAttempts.Get(ctx, key)
	if err != nil {
		slog.Error("login attempts: lectura fallida", slog.String("error", err.Error()))
		return nil
	}
	if la != nil && la.LockedUntil != nil && la.LockedUntil.After(time.Now()) {
		slog.Warn("security: login on locked account", slog.String("event", "account_locked"), slog.String("email_hash", key))
		return &LoginLockedError{Until: *la.LockedUntil}
	}
	return nil
}

// registerLoginFailure suma el fallo y devuelve el error a mostrar:
// cuántos intentos quedan o el bloqueo si se llegó al máximo.
func (s *authService) registerLoginFailure(ctx context.Context, key string) error {
	base := "la contraseña ingresada es incorrecta"
	if s.loginAttempts == nil {
		return errors.New(base)
	}
	max := loginMaxAttempts()
	la, err := s.loginAttempts.RegisterFailure(ctx, key, max, loginLockDuration())
	if err != nil {
		slog.Error("login attempts: escritura fallida", slog.String("error", err.Error()))
		return errors.New(base)
	}
	if la.LockedUntil != nil && la.LockedUntil.After(time.Now()) {
		slog.Warn("security: account locked after failed attempts", slog.String("event", "account_locked_now"), slog.String("email_hash", key))
		return &LoginLockedError{Until: *la.LockedUntil}
	}
	return fmt.Errorf("%s. Te quedan %d intento(s) antes del bloqueo temporal", base, max-la.Failed)
}

func (s *authService) resetLoginFailures(ctx context.Context, key string) {
	if s.loginAttempts == nil {
		return
	}
	if err := s.loginAttempts.Reset(ctx, key); err != nil {
		slog.Error("login attempts: reset fallido", slog.String("error", err.Error()))
	}
}
