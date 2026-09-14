package actions

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrActionNotFound = errors.New("action not found")

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
	action *Action,
) error {
	const query = `
		INSERT INTO  actions (
			id,
			task_id,
			action_type,
			execution_order,
			configuration,
			continue_on_failure,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, #8)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		action.ID,
		action.TaskID,
		action.ActionType,
		action.ExecutionOrder,
		action.Configuration,
		action.ContinueOnFailure,
		action.CreatedAt,
		action.UpdatedAt,
	)

	return err
}

func (r *Repository) FindByID(
	ctx context.Context,
	actionID string,
) (*Action, error) {
	const query = `
		SELECT
			id,
			task_id,
			action_type,
			execution_order,
			configuration,
			continue_on_failure,
			created_at,
			updated_at
		FROM actions
		WHERE id = $1
	`

	action := &Action{}

	err := r.db.QueryRow(
		ctx,
		query,
		actionID,
	).Scan(
		&action.ID,
		&action.TaskID,
		&action.ActionType,
		&action.ExecutionOrder,
		&action.Configuration,
		&action.ContinueOnFailure,
		&action.CreatedAt,
		&action.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrActionNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to find action: %w", err)
	}

	return action, nil
}

func (r *Repository) FindByTaskID(
	ctx context.Context,
	taskID string,
) ([]Action, error) {
	const query = `
		SELECT
			id,
			task_id,
			action_type,
			execution_order,
			configuration,
			continue_on_failure,
			created_at,
			updated_at
		FROM actions
		WHERE task_id = $1
		ORDER BY execution_order ASC
	`

	rows, err := r.db.Query(ctx, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to find actions: %w", err)
	}
	defer rows.Close()

	var actions []Action

	for rows.Next() {
		var action Action

		if err := rows.Scan(
			&action.ID,
			&action.TaskID,
			&action.ActionType,
			&action.ExecutionOrder,
			&action.Configuration,
			&action.ContinueOnFailure,
			&action.CreatedAt,
			&action.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan action: %w", err)
		}

		actions = append(actions, action)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate actions: %w", err)
	}

	return actions, nil
}

func (r *Repository) Update(
	ctx context.Context,
	action *Action,
) error {
	const query = `
		UPDATE actions
		SET
			task_id = $2,
			action_type = $3,
			execution_order = $4,
			configuration = $5,
			continue_on_failure = $6,
			updated_at = $7
		WHERE id = $1
	`

	tag, err := r.db.Exec(
		ctx,
		query,
		action.ID,
		action.TaskID,
		action.ActionType,
		action.ExecutionOrder,
		action.Configuration,
		action.ContinueOnFailure,
		action.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update action: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrActionNotFound
	}

	return nil
}

func (r *Repository) Delete(
	ctx context.Context,
	actionID string,
) error {
	const query = `
		DELETE FROM actions
		WHERE id = $1
	`

	tag, err := r.db.Exec(ctx, query, actionID)
	if err != nil {
		return fmt.Errorf("failed to delete action: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrActionNotFound
	}

	return nil
}
