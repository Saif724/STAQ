package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrEmailVerificationNotFound = errors.New("email verification not found")

type EmailVerificationRepository struct {
	db *pgxpool.Pool
}

func NewEmailVerificationRepository(
	db *pgxpool.Pool,
) *EmailVerificationRepository {
	return &EmailVerificationRepository{
		db: db,
	}
}

func (r *EmailVerificationRepository) Create(
	ctx context.Context,
	verification *EmailVerification,
) error {
	query := `
		INSERT INTO email_verifications (
			id,
			user_id,
			token,
			expires_at,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		verification.ID,
		verification.UserID,
		verification.Token,
		verification.ExpiresAt,
		verification.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create email verification: %w", err)
	}

	return nil
}

func (r *EmailVerificationRepository) FindByToken(
	ctx context.Context,
	token string,
) (*EmailVerification, error) {
	query := `
		SELECT
			id,
			user_id,
			token,
			expires_at,
			verified_at,
			revoked_at,
			created_at
		FROM email_verifications
		WHERE token = $1
	`

	var verification EmailVerification

	err := r.db.QueryRow(
		ctx,
		query,
		token,
	).Scan(
		&verification.ID,
		&verification.UserID,
		&verification.Token,
		&verification.ExpiresAt,
		&verification.VerifiedAt,
		&verification.RevokedAt,
		&verification.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrEmailVerificationNotFound
		}

		return nil, fmt.Errorf(
			"failed to find email verification: %w",
			err,
		)
	}

	return &verification, nil
}

func (r *EmailVerificationRepository) MarkVerified(
	ctx context.Context,
	id string,
) error {
	query := `
		UPDATE email_verifications
		SET verified_at = NOW()
		WHERE id = $1
			AND verified_at IS NULL
	`

	_, err := r.db.Exec(ctx, query, id)

	if err != nil {
		return fmt.Errorf(
			"failed to mark email verification as verified: %w",
			err,
		)
	}

	return nil
}

func (r *EmailVerificationRepository) RevokeForUser(
	ctx context.Context,
	userID string,
) error {
	query := `
		UPDATE email_verifications
		SET revoked_at = NOW()
		WHERE user_id = $1
			AND revoked_at IS NULL
	`

	_, err := r.db.Exec(ctx, query, userID)

	if err != nil {
		return fmt.Errorf(
			"failed to revoke email verifications: %w",
			err,
		)
	}

	return nil
}
