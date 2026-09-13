package queues

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrQueueNotFound = errors.New("queue not found")

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
	queue *Queue,
) error {
	query := `
		INSERT INTO queues (
			id,
			name,
			description,
			is_active,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		queue.ID,
		queue.Name,
		queue.Description,
		queue.IsActive,
		queue.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create queue: %w", err)
	}

	return nil
}

func (r *Repository) FindByID(
	ctx context.Context,
	queueID string,
) (*Queue, error) {
	query := `
		SELECT
			id,
			name,
			description,
			is_active,
			created_at
		FROM queues
		WHERE id = $1
	`

	var queue Queue

	err := r.db.QueryRow(ctx, query, queueID).Scan(
		&queue.ID,
		&queue.Name,
		&queue.Description,
		&queue.IsActive,
		&queue.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrQueueNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to find queue: %w", err)
	}

	return &queue, nil
}

func (r *Repository) FindAll(
	ctx context.Context,
) ([]Queue, error) {
	query := `
		SELECT
			id,
			name,
			description,
			is_active,
			created_at
		FROM queues
		ORDER BY name ASC
	`

	rows, err := r.db.Query(ctx, query)

	if err != nil {
		return nil, fmt.Errorf("failed to find queues: %w", err)
	}

	defer rows.Close()

	queues := make([]Queue, 0)

	for rows.Next() {
		var queue Queue

		if err := rows.Scan(
			&queue.ID,
			&queue.Name,
			&queue.Description,
			&queue.IsActive,
			&queue.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan queue: %w", err)
		}

		queues = append(queues, queue)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed while reading queues: %w", err)
	}

	return queues, nil
}

func (r *Repository) Update(
	ctx context.Context,
	queue *Queue,
) error {
	query := `
		UPDATE queues
		SET
			name = $1,
			description = $2,
			is_active = $3
		WHERE id = &4
	`

	result, err := r.db.Exec(
		ctx,
		query,
		queue.Name,
		queue.Description,
		queue.IsActive,
		queue.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update queue: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrQueueNotFound
	}

	return nil
}

func (r *Repository) Delete(
	ctx context.Context,
	queueID string,
) error {
	query := `
		DELETE FROM queues
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, queueID)

	if err != nil {
		return fmt.Errorf("failed to delete queue: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrQueueNotFound
	}

	return nil
}
