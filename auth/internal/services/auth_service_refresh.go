package services

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"auth/internal/dto"
	"auth/internal/utils"
)

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*dto.LoginResponse, error) {
	hash := utils.HashToken(refreshToken)
	rt, err := s.refreshTokenRepo.FindByTokenHash(ctx, hash)
	if err != nil {
		return nil, errors.New("refresh token not found")
	}

	if rt.RevokedAt != nil {
		// HU-04: un refresh ya usado que vuelve a aparecer indica que
		// pudo ser robado. Se cierran todas las sesiones del usuario: el
		// atacante y el dueño quedan afuera y el dueño vuelve a iniciar
		// sesión.
		slog.Warn("security: refresh token reuse detected", slog.String("event", "refresh_token_reuse"), slog.String("user_id", rt.UserID))
		_ = s.refreshTokenRepo.RevokeAllForUser(ctx, rt.UserID)
		return nil, ErrRefreshTokenReused
	}
	if rt.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("refresh token expired or revoked")
	}

	user, err := s.userRepo.FindByID(ctx, rt.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	statusLower := strings.ToLower(strings.TrimSpace(user.Status))
	if statusLower != "active" {
		return nil, ErrAccountBlocked
	}

	approvalStatus, err := s.validateUserAccess(ctx, user)
	if err != nil {
		return nil, err
	}

	// Revoke old token
	_ = s.refreshTokenRepo.Revoke(ctx, rt.RefreshTokenID)

	// Generate new token pair
	accessToken, newRefreshToken, err := s.jwtService.GenerateTokenPair(user.UserID, user.RoleID, false)
	if err != nil {
		return nil, err
	}

	if err := s.saveRefreshToken(ctx, user.UserID, newRefreshToken); err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		Token:          accessToken,
		RefreshToken:   newRefreshToken,
		User:           user,
		ApprovalStatus: approvalStatus,
		UserStatus:     user.Status,
	}, nil
}
