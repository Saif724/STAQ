package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)


type RefreshTokenRepository struct {
	db *pgxpool.Pool
}

func NewRefreshTokenRepository(db *pgxpool.Pool) *RefreshTokenRepository {
	return &RefreshTokenRepository{
		db: db,
	}
}

func (r *RefreshTokenRepository) Create(
	ctx context.Context,
	token *RefreshToken,
) error {
	query := `
		INSERT INTO refresh_tokens (
			id,
			user_id,
			token_hash,
			expires_at,
			created_at,
			last_used_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		token.ID,
		token.UserID,
		token.TokenHash,
		token.ExpiresAt,
		token.CreatedAt,
		token.LastUsedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create refresh token: %w", err)
	}

	return nil
}

func (r *RefreshTokenRepository) FindByHash(
	ctx context.Context,
	tokenHash string,
) (*RefreshToken, error) {
	query := `
		SELECT
			id,
			user_id,
			token_hash,
			expires_at,
			revoked_at,
			created_at,
			last_used_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`

	var token RefreshToken

	err := r.db.QueryRow(
		ctx,
		query,
		tokenHash,
	).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.RevokedAt,
		&token.CreatedAt,
		&token.LastUsedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRefreshTokenNotFound
		}

		return nil, fmt.Errorf("failed to find refresh token: %w", err)
	}

	return &token, nil
}

func (r *RefreshTokenRepository) Revoke(
	ctx context.Context,
	tokenID string,
) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE id = $1
			AND revoked_at IS NULL
	`

	_, err := r.db.Exec(ctx, query, tokenID)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	return nil
}

func (r *RefreshTokenRepository) UpdateLastUsed(
	ctx context.Context,
	tokenID string,
) error {
	query := `
		UPDATE refresh_tokens
		SET last_used_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, tokenID)
	if err != nil {
		return fmt.Errorf("failed to update refresh token usage: %w", err)
	}

	return nil
}
