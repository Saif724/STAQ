package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrOAuthAccountNotFound = errors.New("oauth account not found")

type OAuthAccount struct {
	ID             string
	UserID         string
	Provider       string
	ProviderUserID string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type OAuthAccountRepository struct {
	db *pgxpool.Pool
}

func NewOAuthAccountRepository(db *pgxpool.Pool) *OAuthAccountRepository {
	return &OAuthAccountRepository{
		db: db,
	}
}

func (r *OAuthAccountRepository) FindByProvider(
	ctx context.Context,
	provider string,
	providerUserID string,
) (*OAuthAccount, error) {
	query := `
		SELECT 
			id,
			user_id,
			provider,
			provider_user_id,
			created_at,
			updated_at
		FROM oauth_accounts
		WHERE provider = $1
			AND provider_user_id = $2
	`

	var account OAuthAccount
	err := r.db.QueryRow(
		ctx,
		query,
		provider,
		providerUserID,
	).Scan(
		&account.ID,
		&account.UserID,
		&account.Provider,
		&account.ProviderUserID,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrOAuthAccountNotFound
		}
		return nil, fmt.Errorf("failed to find oauth account: %w", err)
	}

	return &account, nil
}

func (r *OAuthAccountRepository) Create(
	ctx context.Context,
	account *OAuthAccount,
) error {
	query := `
		INSERT INTO oauth_accounts (
			id,
			user_id,
			provider,
			provider_user_id,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		account.ID,
		account.UserID,
		account.Provider,
		account.ProviderUserID,
		account.CreatedAt,
		account.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create oauth account: %w", err)
	}

	return nil
}
