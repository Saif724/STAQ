package triggers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrTriggerNotFound = errors.New("trigger not found")

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
	trigger *Trigger,
) error {
	const query = `
		INSERT INTO triggers (
			id,
			task_id,
			trigger_type,
			cron_expression,
			timezone,
			next_run_at,
			last_run_at,
			is_active,
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
		trigger.ID,
		trigger.TaskID,
		trigger.TriggerType,
		trigger.CronExpression,
		trigger.TimeZone,
		trigger.NextRunAt,
		trigger.LastRunAt,
		trigger.IsActive,
		trigger.CreatedAt,
		trigger.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create trigger: %w", err)
	}

	return nil
}

func (r *Repository) FindByID(
	ctx context.Context,
	triggerID string,
) (*Trigger, error) {
	const query = `
		SELECT
			id,
			task_id,
			trigger_type,
			cron_expression,
			timezone,
			next_run_at,
			last_run_at,
			is_active,
			created_at,
			updated_at
		FROM triggers
		WHERE id = $1
	`

	trigger := &Trigger{}

	err := r.db.QueryRow(ctx, query, triggerID).Scan(
		&trigger.ID,
		&trigger.TaskID,
		&trigger.TriggerType,
		&trigger.CronExpression,
		&trigger.TimeZone,
		&trigger.NextRunAt,
		&trigger.LastRunAt,
		&trigger.IsActive,
		&trigger.CreatedAt,
		&trigger.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrTriggerNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to find trigger: %w", err)
	}

	return trigger, nil
}

func (r *Repository) FindByTaskID(
	ctx context.Context,
	taskID string,
) ([]Trigger, error) {
	const query = `
		SELECT
			id,
			task_id,
			trigger_type,
			cron_expression,
			timezone,
			next_run_at,
			last_run_at,
			is_active,
			created_at,
			updated_at
		FROM triggers
		WHERE task_id = $1
		ORDER BY next_run_at ASC
	`

	rows, err := r.db.Query(ctx, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to find task triggers: %w", err)
	}
	defer rows.Close()

	var triggers []Trigger

	for rows.Next() {
		var trigger Trigger
		if err := rows.Scan(
			&trigger.ID,
			&trigger.TaskID,
			&trigger.TriggerType,
			&trigger.CronExpression,
			&trigger.TimeZone,
			&trigger.NextRunAt,
			&trigger.LastRunAt,
			&trigger.IsActive,
			&trigger.CreatedAt,
			&trigger.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan trigger: %w", err)
		}

		triggers = append(triggers, trigger)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate triggers: %w", err)
	}

	return triggers, nil
}

func (r *Repository) FindDue(
	ctx context.Context,
	now time.Time,
	limit int,
) ([]Trigger, error) {
	const query = `
		SELECT
			id,
			task_id,
			trigger_type,
			cron_expression,
			timezone,
			next_run_at,
			last_run_at,
			is_active,
			created_at,
			updated_at
		FROM triggers
		WHERE is_active = TRUE
			AND next_run_at <= $1
		ORDER BY next_run_at ASC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, now, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to find due triggers: %w", err)
	}
	defer rows.Close()

	var triggers []Trigger

	for rows.Next() {
		var trigger Trigger
		if err := rows.Scan(
			&trigger.ID,
			&trigger.TaskID,
			&trigger.TriggerType,
			&trigger.CronExpression,
			&trigger.TimeZone,
			&trigger.NextRunAt,
			&trigger.LastRunAt,
			&trigger.IsActive,
			&trigger.CreatedAt,
			&trigger.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan trigger: %w", err)
		}

		triggers = append(triggers, trigger)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate due triggers: %w", err)
	}

	return triggers, nil
}

func (r *Repository) Update(
	ctx context.Context,
	trigger *Trigger,
) error {
	const query = `
		UPDATE triggers
		SET
			triggers_type = $1
			cron_expression = $2
			timezone = $3
			next_run_at = $4
			last_run_at = $5
			is_active = $6
			updated_at = $7
		WHERE id = $8
	`

	result, err := r.db.Exec(
		ctx,
		query,
		trigger.TriggerType,
		trigger.CronExpression,
		trigger.TimeZone,
		trigger.NextRunAt,
		trigger.LastRunAt,
		trigger.IsActive,
		trigger.UpdatedAt,
		trigger.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update trigger: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrTriggerNotFound
	}

	return nil
}

func (r *Repository) Delete(
	ctx context.Context,
	triggerID string,
) error {
	const query = `
		DELETE FROM triggers
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, triggerID)

	if err != nil {
		return fmt.Errorf("failed to delete trigger: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrTriggerNotFound
	}

	return nil
}
