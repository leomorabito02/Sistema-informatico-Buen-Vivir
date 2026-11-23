package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
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

type VerificationService interface {
	GenerateAndSendVerification(ctx context.Context, user *models.User) error
	VerifyEmail(ctx context.Context, rawToken string) error
	ResendVerification(ctx context.Context, email string) error
	IssueActionLink(ctx context.Context, requestID, userID string) (string, error)
}

type verificationService struct {
	evRepo           repositories.EmailVerificationRepository
	userRepo         repositories.UserRepository
	notificationSync NotificationSyncService
	producer         producers.UsuariosProducer
}

func NewVerificationService(
	evRepo repositories.EmailVerificationRepository,
	userRepo repositories.UserRepository,
	notificationSync NotificationSyncService,
	producer ...producers.UsuariosProducer,
) VerificationService {
	svc := &verificationService{
		evRepo:           evRepo,
		userRepo:         userRepo,
		notificationSync: notificationSync,
	}
	if len(producer) > 0 {
		svc.producer = producer[0]
	}
	return svc
}

func (s *verificationService) GenerateAndSendVerification(ctx context.Context, user *models.User) error {
	if user == nil {
		return errors.New("el usuario no puede ser nulo")
	}

	requestID := "ev-" + utils.GenerateRandomString(10)

	if s.producer != nil {
		if err := s.producer.EmailSolicitado(ctx, requestID, user.UserID, "verificacion"); err != nil {
			slog.Error("verification_service: failed to enqueue verification email", "user_id", user.UserID, "error", err)
		}
	}

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

	user, err := s.userRepo.FindByID(ctx, ev.UserID)
	if err == nil && user != nil && user.Email != "" {
		if err := s.userRepo.UpdateEmailVerifiedByEmail(ctx, user.Email, true); err != nil {
			return fmt.Errorf("error al actualizar el estado de verificación del usuario: %w", err)
		}
	} else {
		if err := s.userRepo.UpdateEmailVerified(ctx, ev.UserID, true); err != nil {
			return fmt.Errorf("error al actualizar el estado de verificación del usuario: %w", err)
		}
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

	users, err := s.userRepo.FindAllByEmail(ctx, trimmedEmail)
	if err != nil || len(users) == 0 {
		return errors.New("no existe ninguna cuenta registrada con este correo electrónico")
	}

	allVerified := true
	for _, u := range users {
		if !u.IsEmailVerified {
			allVerified = false
			break
		}
	}
	if allVerified {
		return errors.New("el correo electrónico de esta cuenta ya ha sido verificado")
	}

	return s.GenerateAndSendVerification(ctx, &users[0])
}

func (s *verificationService) IssueActionLink(ctx context.Context, requestID, userID string) (string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("error generando token criptográfico: %w", err)
	}
	rawToken := hex.EncodeToString(tokenBytes)
	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	if requestID == "" {
		requestID = "ev-" + utils.GenerateRandomString(10)
	}

	ev := &models.EmailVerification{
		EmailVerificationID: requestID,
		UserID:              userID,
		TokenHash:           tokenHash,
		ExpiresAt:           time.Now().Add(24 * time.Hour),
	}

	_ = s.evRepo.Delete(ctx, requestID)
	if err := s.evRepo.Create(ctx, ev); err != nil {
		return "", fmt.Errorf("error guardando token de verificación: %w", err)
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
	return fmt.Sprintf("%s%s?token=%s", strings.TrimRight(frontendURL, "/"), verifyPath, rawToken), nil
}

