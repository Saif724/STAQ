package users

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrEmailExists  = errors.New("email already exists")
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) Create(ctx context.Context, user *User) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin user creation transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO users (
			id, full_name, email, password_hash, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`,
		user.ID,
		user.FullName,
		user.Email,
		user.PasswordHash,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	result, err := tx.Exec(ctx, `
		INSERT INTO user_settings (
			id, user_id, timezone, default_queue_id,
			created_at, updated_at
		)
		SELECT
			gen_random_uuid(), $1, 'UTC', q.id, NOW(), NOW()
		FROM queues q
		WHERE q.name = 'default'
			AND q.is_active = TRUE
	`, user.ID)
	if err != nil {
		return fmt.Errorf("failed to create user settings: %w", err)
	}

	if result.RowsAffected() != 1 {
		return errors.New("cannot create user settings: active default queue not found")
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit user creation: %w", err)
	}

	return nil
}

func (r *Repository) FindByID(
	ctx context.Context,
	id string,
) (*User, error) {
	query := `
		SELECT
			u.id,
			u.full_name,
			u.email,
			u.password_hash,
			u.email_verified,
			u.is_active,
			u.created_at,
			u.updated_at,
			COALESCE(us.timezone, 'UTC')
		FROM users u
		LEFT JOIN user_settings us ON us.user_id = u.id
		WHERE u.id = $1
	`

	user := &User{}

	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.FullName,
		&user.Email,
		&user.PasswordHash,
		&user.EmailVerified,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Timezone,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to find user by id: %w", err)
	}

	return user, nil
}

func (r *Repository) UpdateProfile(
	ctx context.Context,
	userID string,
	fullName string,
	timezone string,
) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin profile update: %w", err)
	}
	defer tx.Rollback(ctx)

	result, err := tx.Exec(ctx, `
		UPDATE users
		SET full_name = $1, updated_at = NOW()
		WHERE id = $2
	`, fullName, userID)
	if err != nil {
		return fmt.Errorf("failed to update profile name: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	result, err = tx.Exec(ctx, `
		UPDATE user_settings
		SET timezone = $1, updated_at = NOW()
		WHERE user_id = $2
	`, timezone, userID)
	if err != nil {
		return fmt.Errorf("failed to update user timezone: %w", err)
	}
	if result.RowsAffected() == 0 {
		return errors.New("user settings not found")
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit profile update: %w", err)
	}
	return nil
}

func (r *Repository) FindByEmail(
	ctx context.Context,
	email string,
) (*User, error) {
	query := `
		SELECT
			id,
			full_name,
			email,
			password_hash,
			email_verified,
			is_active,
			created_at,
			updated_at
		FROM users
		WHERE email=$1
	`

	user := &User{}

	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.FullName,
		&user.Email,
		&user.PasswordHash,
		&user.EmailVerified,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}

	return user, nil
}

func (r *Repository) MarkEmailVerified(
	ctx context.Context,
	userID string,
) error {
	query := `
		UPDATE users
		SET
			email_verified = TRUE,
			updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to mark email as verified: %w", err)
	}

	return nil
}

func (r *Repository) UpdateForReRegistration(
	ctx context.Context,
	userID string,
	fullName string,
	passwordHash string,
) error {
	query := `
		UPDATE users
		SET
			full_name = $1,
			password_hash = $2,
			updated_at = NOW()
		WHERE id = $3
			AND email_verified = FALSE
	`

	result, err := r.db.Exec(
		ctx,
		query,
		fullName,
		passwordHash,
		userID,
	)

	if err != nil {
		return fmt.Errorf("failed to update user for re-registration: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}
