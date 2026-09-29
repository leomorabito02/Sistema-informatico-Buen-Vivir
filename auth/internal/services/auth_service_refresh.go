package services

import (
	"context"
	"errors"
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

	if rt.RevokedAt != nil || rt.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("refresh token expired or revoked")
	}

	user, err := s.userRepo.FindByID(ctx, rt.UserID)
	if err != nil {
		return nil, errors.New("user not found")
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
		Token:        accessToken,
		RefreshToken: newRefreshToken,
		User:         user,
		UserStatus:   user.Status,
	}, nil
}
