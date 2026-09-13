package tasks

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrTaskNotFound = errors.New("task not found")

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
	task *Task,
) error {
	query := `
		INSERT INTO tasks (
			id,
			user_id,
			queue_id,
			name,
			description,
			status,
			timeout_seconds,
			max_retries,
			created_at,
			updated_at
		)
		VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10
		)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		task.ID,
		task.UserID,
		task.QueueID,
		task.Name,
		task.Description,
		task.Status,
		task.TimeoutSeconds,
		task.MaxRetries,
		task.CreatedAt,
		task.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	return nil
}

func (r *Repository) FindByID(
	ctx context.Context,
	taskID string,
) (*Task, error) {
	query := `
		SELECT
			id,
			user_id,
			queue_id,
			name,
			description,
			status,
			timeout_seconds,
			max_retries,
			created_at,
			updated_at
		FROM tasks
		WHERE id = $1
	`

	var task Task

	err := r.db.QueryRow(ctx, query, taskID).Scan(
		&task.ID,
		&task.UserID,
		&task.QueueID,
		&task.Name,
		&task.Description,
		&task.Status,
		&task.TimeoutSeconds,
		&task.MaxRetries,
		&task.CreatedAt,
		&task.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrTaskNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to find task: %w", err)
	}

	return &task, nil
}

func (r *Repository) FindByIDAndUser(
	ctx context.Context,
	taskID string,
	userID string,
) (*Task, error) {
	query := `
		SELECT
			id,
			user_id,
			queue_id,
			name,
			description,
			status,
			timeout_seconds,
			max_retries,
			created_at,
			updated_at
		FROM tasks
		WHERE id = $1
			AND user_id = $2
	`

	var task Task

	err := r.db.QueryRow(ctx, query, taskID, userID).Scan(
		&task.ID,
		&task.UserID,
		&task.QueueID,
		&task.Name,
		&task.Description,
		&task.Status,
		&task.TimeoutSeconds,
		&task.MaxRetries,
		&task.CreatedAt,
		&task.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrTaskNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to find task: %w", err)
	}

	return &task, nil
}

func (r *Repository) FindByUser(
	ctx context.Context,
	userID string,
) ([]Task, error) {
	query := `
		SELECT
			id,
			user_id,
			queue_id,
			name,
			description,
			status,
			timeout_seconds,
			max_retries,
			created_at,
			updated_at
		FROM tasks
		WHERE user_id = $1
	`

	rows, err := r.db.Query(ctx, query, userID)

	if err != nil {
		return nil, fmt.Errorf("failed to find user tasks: %w", err)
	}
	defer rows.Close()

	tasks := make([]Task, 0)

	for rows.Next() {
		var task Task

		if err := rows.Scan(
			&task.ID,
			&task.UserID,
			&task.QueueID,
			&task.Name,
			&task.Description,
			&task.Status,
			&task.TimeoutSeconds,
			&task.MaxRetries,
			&task.CreatedAt,
			&task.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}

		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed while reading tasks: %w", err)
	}

	return tasks, nil
}

func (r *Repository) Update(
	ctx context.Context,
	task *Task,
) error {
	query := `
		UPDATE tasks
		SET
			queue_id = $1,
			name = $2,
			description = $3,
			status = $4,
			timeout_seconds = $5,
			max_retries = $6,
			updated_at = $7
		WHERE id = $8
			AND user_id = $9
	`

	result, err := r.db.Exec(
		ctx,
		query,
		task.QueueID,
		task.Name,
		task.Description,
		task.Status,
		task.TimeoutSeconds,
		task.MaxRetries,
		task.UpdatedAt,
		task.ID,
		task.UserID,
	)

	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrTaskNotFound
	}

	return nil
}

func (r *Repository) Archive(
	ctx context.Context,
	taskID string,
	userID string,
) error {
	query := `
		UPDATE tasks
		SET
			status = $1,
			updated_at = NOW()
		WHERE id = $2
			AND user_id = $3
	`

	result, err := r.db.Exec(
		ctx,
		query,
		StatusArchived,
		taskID,
		userID,
	)

	if err != nil {
		return fmt.Errorf("failed to archive task: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrTaskNotFound
	}

	return nil
}
