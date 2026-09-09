package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Saif724/STAQ/backend/internal/auth/dto"
	"github.com/google/uuid"
)

const refreshTokenLifeTime = 30 * 24 * time.Hour

func (s *Service) createRefreshToken(
	ctx context.Context,
	userID string,
) (string, error) {
	rawToken, err := generateRefreshToken()
	if err != nil {
		return "", err
	}

	token := &RefreshToken{
		ID:        uuid.NewString(),
		UserID:    userID,
		TokenHash: hashRefreshToken(rawToken),
		ExpiresAt: time.Now().UTC().Add(refreshTokenLifeTime),
		CreatedAt: time.Now().UTC(),
	}

	if err := s.refreshTokenRepository.Create(ctx, token); err != nil {
		return "", fmt.Errorf("failed to create refresh token: %w", err)
	}

	return rawToken, nil
}

func (s *Service) RefreshAccessToken(
	ctx context.Context,
	rawRefreshToken string,
) (*dto.RefreshResponse, error) {
	if rawRefreshToken == "" {
		return nil, ErrInvalidRefreshToken
	}

	tokenHash := hashRefreshToken(rawRefreshToken)

	token, err := s.refreshTokenRepository.FindByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, ErrRefreshTokenNotFound) {
			return nil, ErrInvalidRefreshToken
		}

		return nil, fmt.Errorf("failed to validate refresh token: %w", err)
	}

	now := time.Now().UTC()

	if token.RevokedAt != nil {
		return nil, ErrInvalidRefreshToken
	}

	if !now.Before(token.ExpiresAt) {
		return nil, ErrInvalidRefreshToken
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(token.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	if err := s.refreshTokenRepository.UpdateLastUsed(ctx, token.ID); err != nil {
		return nil, fmt.Errorf("failed to update refresh token usage: %w", err)
	}

	return &dto.RefreshResponse{
		AccessToken: accessToken,
	}, nil
}

func (s *Service) Logout(
	ctx context.Context,
	rawRefreshToken string,
) error {
	if rawRefreshToken == "" {
		return ErrInvalidRefreshToken
	}

	tokenHash := hashRefreshToken(rawRefreshToken)

	token, err := s.refreshTokenRepository.FindByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, ErrRefreshTokenNotFound) {
			return ErrInvalidRefreshToken
		}

		return fmt.Errorf("failed to find refresh token: %w", err)
	}

	if token.RevokedAt != nil {
		return ErrInvalidRefreshToken
	}

	if err := s.refreshTokenRepository.Revoke(ctx, token.ID); err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	return nil
}
