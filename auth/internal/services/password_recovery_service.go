package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"auth/internal/events/producers"
	"auth/internal/models"
	"auth/internal/repositories"
	"auth/internal/utils"
)

type PasswordRecoveryService interface {
	RequestReset(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, rawToken, newPassword string) error
	IssueActionLink(ctx context.Context, requestID, userID string) (string, error)
}

type passwordRecoveryService struct {
	recoveryRepo     repositories.PasswordRecoveryRepository
	userRepo         repositories.UserRepository
	refreshTokenRepo repositories.RefreshTokenRepository
	notificationSync NotificationSyncService
	producer         producers.UsuariosProducer
}

func NewPasswordRecoveryService(
	recoveryRepo repositories.PasswordRecoveryRepository,
	userRepo repositories.UserRepository,
	refreshTokenRepo repositories.RefreshTokenRepository,
	notificationSync NotificationSyncService,
	producer ...producers.UsuariosProducer,
) PasswordRecoveryService {
	svc := &passwordRecoveryService{
		recoveryRepo:     recoveryRepo,
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		notificationSync: notificationSync,
	}
	if len(producer) > 0 {
		svc.producer = producer[0]
	}
	return svc
}

func (s *passwordRecoveryService) RequestReset(ctx context.Context, email string) error {
	trimmedEmail := strings.TrimSpace(email)
	if trimmedEmail == "" {
		return errors.New("el correo electrónico es obligatorio")
	}

	users, err := s.userRepo.FindAllByEmail(ctx, trimmedEmail)
	if err != nil || len(users) == 0 {
		slog.Info("password recovery requested for non-existent email", slog.String("email_hash", utils.HashToken(trimmedEmail)))
		// Anti-enumeration: return nil so attacker cannot discern registered emails
		return nil
	}

	primaryUser := &users[0]
	_ = s.recoveryRepo.DeleteByUserID(ctx, primaryUser.UserID)

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return fmt.Errorf("error generando token criptográfico: %w", err)
	}
	rawToken := hex.EncodeToString(tokenBytes)
	tokenHash := utils.HashToken(rawToken)

	recovery := &models.PasswordRecovery{
		PasswordRecoveryID: "pr-" + utils.GenerateRandomString(10),
		UserID:             primaryUser.UserID,
		TokenHash:          tokenHash,
		ExpiresAt:          time.Now().Add(30 * time.Minute),
	}

	if err := s.recoveryRepo.Create(ctx, recovery); err != nil {
		return fmt.Errorf("error guardando solicitud de recuperación: %w", err)
	}

	if s.producer != nil {
		_ = s.producer.EmailSolicitado(ctx, recovery.PasswordRecoveryID, primaryUser.UserID, "recuperacion_password")
	}

	return nil
}

func validateRecoveryToken(pr *models.PasswordRecovery) error {
	if pr.UsedAt != nil {
		return errors.New("el enlace de restablecimiento ya fue utilizado")
	}
	if time.Now().After(pr.ExpiresAt) {
		return errors.New("el enlace de restablecimiento ha expirado")
	}
	return nil
}

func (s *passwordRecoveryService) revokeUserSessionsAsync(email, primaryUserID string) {
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		accounts, err := s.userRepo.FindAllByEmail(bgCtx, email)
		if err == nil {
			for _, acc := range accounts {
				_ = s.refreshTokenRepo.RevokeAllForUser(bgCtx, acc.UserID)
			}
			return
		}
		_ = s.refreshTokenRepo.RevokeAllForUser(bgCtx, primaryUserID)
	}()
}

func (s *passwordRecoveryService) ResetPassword(ctx context.Context, rawToken, newPassword string) error {
	trimmedToken := strings.TrimSpace(rawToken)
	if trimmedToken == "" {
		return errors.New("el token de recuperación es obligatorio")
	}

	if err := utils.ValidatePasswordComplexity(newPassword); err != nil {
		return err
	}

	tokenHash := utils.HashToken(trimmedToken)
	pr, err := s.recoveryRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return errors.New("el enlace de restablecimiento es inválido, ya fue utilizado o ha expirado")
	}

	if err := validateRecoveryToken(pr); err != nil {
		return err
	}

	user, err := s.userRepo.FindByID(ctx, pr.UserID)
	if err != nil || user == nil {
		return errors.New("usuario no encontrado")
	}

	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return errors.New("error al procesar la contraseña")
	}

	// Update password and mark email as verified across all accounts with this email in a single query
	if err := s.userRepo.UpdatePasswordAndVerifiedByEmail(ctx, user.Email, hashedPassword, true); err != nil {
		return fmt.Errorf("error al actualizar la contraseña: %w", err)
	}

	if err := s.recoveryRepo.MarkAsUsed(ctx, pr.PasswordRecoveryID); err != nil {
		return fmt.Errorf("error al marcar el token como utilizado: %w", err)
	}

	s.revokeUserSessionsAsync(user.Email, user.UserID)

	return nil
}

func (s *passwordRecoveryService) IssueActionLink(ctx context.Context, requestID, userID string) (string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("error generando token criptográfico: %w", err)
	}
	rawToken := hex.EncodeToString(tokenBytes)
	tokenHash := utils.HashToken(rawToken)

	if requestID == "" {
		requestID = "pr-" + utils.GenerateRandomString(10)
	}

	recovery := &models.PasswordRecovery{
		PasswordRecoveryID: requestID,
		UserID:             userID,
		TokenHash:          tokenHash,
		ExpiresAt:          time.Now().Add(30 * time.Minute),
	}

	_ = s.recoveryRepo.DeleteByUserID(ctx, userID)
	if err := s.recoveryRepo.Create(ctx, recovery); err != nil {
		return "", fmt.Errorf("error guardando solicitud de recuperación: %w", err)
	}

	frontendURL := strings.TrimSpace(os.Getenv("FRONTEND_URL"))
	if frontendURL == "" {
		frontendURL = "https://ceibuenvivir.org"
	}
	resetPath := strings.TrimSpace(os.Getenv("RESET_PASSWORD_PATH"))
	if resetPath == "" {
		resetPath = "/change-password"
	}
	if !strings.HasPrefix(resetPath, "/") {
		resetPath = "/" + resetPath
	}
	return fmt.Sprintf("%s%s?token=%s", strings.TrimRight(frontendURL, "/"), resetPath, rawToken), nil
}

