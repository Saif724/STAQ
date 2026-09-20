package connections

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrConnectionNotFound = errors.New("connection not found")
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) Create(
	ctx context.Context,
	connection *Connection,
) error {
	const query = `
		INSERT INTO connections (
			id,
			user_id,
			provider,
			provider_account_id,
			account_email,
			access_token_encrypted,
			refresh_token_encrypted,
			token_expires_at,
			scopes,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		connection.ID,
		connection.UserID,
		connection.Provider,
		connection.ProviderAccountID,
		connection.AccountEmail,
		connection.AccessTokenEncrypted,
		connection.RefreshTokenEncrypted,
		connection.TokenExpiresAt,
		connection.Scopes,
		connection.CreatedAt,
		connection.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create connection: %w", err)
	}

	return nil
}

func (r *Repository) FindByID(
	ctx context.Context,
	id string,
) (*Connection, error) {
	const query = `
		SELECT
			id,
			user_id,
			provider,
			provider_account_id,
			account_email,
			access_token_encrypted,
			refresh_token_encrypted,
			token_expires_at,
			scopes,
			created_at,
			updated_at
		FROM connections
		WHERE id = $1
	`

	var connection Connection

	err := r.db.QueryRow(ctx, query, id).Scan(
		&connection.ID,
		&connection.UserID,
		&connection.Provider,
		&connection.ProviderAccountID,
		&connection.AccountEmail,
		&connection.AccessTokenEncrypted,
		&connection.RefreshTokenEncrypted,
		&connection.TokenExpiresAt,
		&connection.Scopes,
		&connection.CreatedAt,
		&connection.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrConnectionNotFound
		}

		return nil, fmt.Errorf("failed to find connection: %w", err)
	}

	return &connection, nil
}

func (r *Repository) FindByIDAndUser(
	ctx context.Context,
	id string,
	userID string,
) (*Connection, error) {
	const query = `
		SELECT
			id,
			user_id,
			provider,
			provider_account_id,
			account_email,
			access_token_encrypted,
			refresh_token_encrypted,
			token_expires_at,
			scopes,
			created_at,
			updated_at
		FROM connections
		WHERE id = $1
			AND user_id = $2
	`

	var connection Connection

	err := r.db.QueryRow(ctx, query, id, userID).Scan(
		&connection.ID,
		&connection.UserID,
		&connection.Provider,
		&connection.ProviderAccountID,
		&connection.AccountEmail,
		&connection.AccessTokenEncrypted,
		&connection.RefreshTokenEncrypted,
		&connection.TokenExpiresAt,
		&connection.Scopes,
		&connection.CreatedAt,
		&connection.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrConnectionNotFound
		}

		return nil, fmt.Errorf("failed to find user connection: %w", err)
	}

	return &connection, nil
}

func (r *Repository) FindByProviderAccount(
	ctx context.Context,
	provider string,
	providerAccountID string,
) (*Connection, error) {
	const query = `
		SELECT
			id,
			user_id,
			provider,
			provider_account_id,
			account_email,
			access_token_encrypted,
			refresh_token_encrypted,
			token_expires_at,
			scopes,
			created_at,
			updated_at
		FROM connections
		WHERE provider = $1
			AND provider_account_id = $2
	`

	var connection Connection

	err := r.db.QueryRow(ctx, query, provider, providerAccountID).Scan(
		&connection.ID,
		&connection.UserID,
		&connection.Provider,
		&connection.ProviderAccountID,
		&connection.AccountEmail,
		&connection.AccessTokenEncrypted,
		&connection.RefreshTokenEncrypted,
		&connection.TokenExpiresAt,
		&connection.Scopes,
		&connection.CreatedAt,
		&connection.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrConnectionNotFound
		}

		return nil, fmt.Errorf("failed to find provider connection: %w", err)
	}

	return &connection, nil
}

func (r *Repository) FindByUser(
	ctx context.Context,
	userID string,
) ([]*Connection, error) {
	const query = `
		SELECT
			id,
			user_id,
			provider,
			provider_account_id,
			account_email,
			access_token_encrypted,
			refresh_token_encrypted,
			token_expires_at,
			scopes,
			created_at,
			updated_at
		FROM connections
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)

	if err != nil {
		return nil, fmt.Errorf("failed to list user connections: %w", err)
	}

	defer rows.Close()

	var connections []*Connection

	for rows.Next() {
		var connection Connection

		if err := rows.Scan(
			&connection.ID,
			&connection.UserID,
			&connection.Provider,
			&connection.ProviderAccountID,
			&connection.AccountEmail,
			&connection.AccessTokenEncrypted,
			&connection.RefreshTokenEncrypted,
			&connection.TokenExpiresAt,
			&connection.Scopes,
			&connection.CreatedAt,
			&connection.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan connection: %w", err)
		}

		connections = append(connections, &connection)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate connections: %w", err)
	}

	return connections, nil
}

func (r *Repository) UpdateTokens(
	ctx context.Context,
	id string,
	accessTokenEncrypted string,
	refreshTokenEncrypted *string,
	tokenExpiresAt *time.Time,
) error {
	const query = `
		UPDATE connections
		SET
			access_token_encrypted = $1,
			refresh_token_encrypted = COALESCE($2, refresh_token_encrypted),
			token_expires_at = $3,
			updated_at = $4
		WHERE id = $5
	`

	result, err := r.db.Exec(
		ctx,
		query,
		accessTokenEncrypted,
		refreshTokenEncrypted,
		tokenExpiresAt,
		time.Now().UTC(),
		id,
	)

	if err != nil {
		return fmt.Errorf("failed to update connection tokens: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrConnectionNotFound
	}

	return nil
}

func (r *Repository) Delete(
	ctx context.Context,
	id string,
	userID string,
) error {
	query := `
		DELETE FROM Connections
		WHERE id = $1
			AND user_id = $2
	`

	result, err := r.db.Exec(ctx, query, id, userID)

	if err != nil {
		return fmt.Errorf("failed to delete connection: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrConnectionNotFound
	}

	return nil
}
