package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"auth/internal/models"
	"auth/internal/repositories"
	"auth/internal/utils"
)

type VerificationService interface {
	GenerateAndSendVerification(ctx context.Context, user *models.User) error
	VerifyEmail(ctx context.Context, rawToken string) error
	ResendVerification(ctx context.Context, email string) error
}

type verificationService struct {
	evRepo           repositories.EmailVerificationRepository
	userRepo         repositories.UserRepository
	notificationSync NotificationSyncService
}

func NewVerificationService(
	evRepo repositories.EmailVerificationRepository,
	userRepo repositories.UserRepository,
	notificationSync NotificationSyncService,
) VerificationService {
	return &verificationService{
		evRepo:           evRepo,
		userRepo:         userRepo,
		notificationSync: notificationSync,
	}
}

func (s *verificationService) GenerateAndSendVerification(ctx context.Context, user *models.User) error {
	if user == nil {
		return errors.New("el usuario no puede ser nulo")
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return fmt.Errorf("error generando token criptográfico: %w", err)
	}
	rawToken := hex.EncodeToString(tokenBytes)

	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	ev := &models.EmailVerification{
		EmailVerificationID: "ev-" + utils.GenerateRandomString(10),
		UserID:              user.UserID,
		TokenHash:           tokenHash,
		ExpiresAt:           time.Now().Add(24 * time.Hour),
	}

	if err := s.evRepo.Create(ctx, ev); err != nil {
		return fmt.Errorf("error guardando token de verificación: %w", err)
	}

	frontendURL := strings.TrimSpace(os.Getenv("FRONTEND_URL"))
	if frontendURL == "" {
		frontendURL = "https://ceibuenvivir.org"
	}
	verifyPath := strings.TrimSpace(os.Getenv("EMAIL_VERIFICATION_PATH"))
	if verifyPath == "" {
		verifyPath = "/verify-email"
	}
	if !strings.HasPrefix(verifyPath, "/") {
		verifyPath = "/" + verifyPath
	}
	actionURL := fmt.Sprintf("%s%s?token=%s", strings.TrimRight(frontendURL, "/"), verifyPath, rawToken)

	fullName := strings.TrimSpace(user.FirstName + " " + user.LastName)
	if fullName == "" {
		fullName = user.Email
	}

	go func() {
		_ = s.notificationSync.SendVerificationEmail(context.Background(), user.UserID, user.Email, fullName, actionURL)
	}()

	return nil
}

func (s *verificationService) VerifyEmail(ctx context.Context, rawToken string) error {
	trimmed := strings.TrimSpace(rawToken)
	if trimmed == "" {
		return errors.New("el token de verificación es obligatorio")
	}

	hash := sha256.Sum256([]byte(trimmed))
	tokenHash := hex.EncodeToString(hash[:])

	ev, err := s.evRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return errors.New("el token de verificación es inválido o no existe")
	}

	if ev.UsedAt != nil {
		return errors.New("el token de verificación ya fue utilizado")
	}

	if time.Now().After(ev.ExpiresAt) {
		return errors.New("el token de verificación ha expirado")
	}

	if err := s.userRepo.UpdateEmailVerified(ctx, ev.UserID, true); err != nil {
		return fmt.Errorf("error al actualizar el estado de verificación del usuario: %w", err)
	}

	if err := s.evRepo.Delete(ctx, ev.EmailVerificationID); err != nil {
		return fmt.Errorf("error al eliminar token de verificación: %w", err)
	}

	return nil
}

func (s *verificationService) ResendVerification(ctx context.Context, email string) error {
	trimmedEmail := strings.TrimSpace(email)
	if trimmedEmail == "" {
		return errors.New("el correo electrónico es obligatorio")
	}

	user, err := s.userRepo.FindByEmail(ctx, trimmedEmail)
	if err != nil || user == nil {
		return errors.New("no existe ninguna cuenta registrada con este correo electrónico")
	}

	if user.IsEmailVerified {
		return errors.New("el correo electrónico de esta cuenta ya ha sido verificado")
	}

	return s.GenerateAndSendVerification(ctx, user)
}
