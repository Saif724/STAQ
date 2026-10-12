package triggers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
	return r.create(ctx, r.db, trigger)
}

func (r *Repository) CreateWithTx(
	ctx context.Context,
	tx pgx.Tx,
	trigger *Trigger,
) error {
	return r.create(ctx, tx, trigger)
}

type triggerExecer interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}

func (r *Repository) create(
	ctx context.Context,
	execer triggerExecer,
	trigger *Trigger,
) error {
	const query = `
		INSERT INTO triggers (
			id,
			task_id,
			trigger_type,
			cron_expression,
			timezone,
			start_at,
			next_run_at,
			last_run_at,
			is_active,
			created_at,
			updated_at
		)
		VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10, $11
		)
	`

	_, err := execer.Exec(
		ctx,
		query,
		trigger.ID,
		trigger.TaskID,
		trigger.TriggerType,
		trigger.CronExpression,
		trigger.TimeZone,
		trigger.StartAt,
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
			start_at,
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
		&trigger.StartAt,
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
			start_at,
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

	triggers := make([]Trigger, 0)

	for rows.Next() {
		var trigger Trigger
		if err := rows.Scan(
			&trigger.ID,
			&trigger.TaskID,
			&trigger.TriggerType,
			&trigger.CronExpression,
			&trigger.TimeZone,
			&trigger.StartAt,
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
			start_at,
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

	triggers := make([]Trigger, 0)

	for rows.Next() {
		var trigger Trigger
		if err := rows.Scan(
			&trigger.ID,
			&trigger.TaskID,
			&trigger.TriggerType,
			&trigger.CronExpression,
			&trigger.TimeZone,
			&trigger.StartAt,
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
			trigger_type = $1,
			cron_expression = $2,
			timezone = $3,
			start_at = $4,
			next_run_at = $5,
			last_run_at = $6,
			is_active = $7,
			updated_at = $8
		WHERE id = $9
	`

	result, err := r.db.Exec(
		ctx,
		query,
		trigger.TriggerType,
		trigger.CronExpression,
		trigger.TimeZone,
		trigger.StartAt,
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

func (r *Repository) FindDueForUpdate(
	ctx context.Context,
	tx pgx.Tx,
	now time.Time,
) (*Trigger, string, error) {
	const query = `
		SELECT
			triggers.id,
			triggers.task_id,
			triggers.trigger_type,
			triggers.cron_expression,
			triggers.timezone,
			triggers.start_at,
			triggers.next_run_at,
			triggers.last_run_at,
			triggers.is_active,
			triggers.created_at,
			triggers.updated_at,
			tasks.status
		FROM triggers
		INNER JOIN tasks ON tasks.id = triggers.task_id
		WHERE triggers.is_active = TRUE
			AND triggers.next_run_at <= $1
		ORDER BY triggers.next_run_at ASC
		LIMIT 1
		FOR UPDATE OF triggers SKIP LOCKED
	`

	trigger := &Trigger{}
	var taskStatus string

	err := tx.QueryRow(
		ctx,
		query,
		now,
	).Scan(
		&trigger.ID,
		&trigger.TaskID,
		&trigger.TriggerType,
		&trigger.CronExpression,
		&trigger.TimeZone,
		&trigger.StartAt,
		&trigger.NextRunAt,
		&trigger.LastRunAt,
		&trigger.IsActive,
		&trigger.CreatedAt,
		&trigger.UpdatedAt,
		&taskStatus,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, "", ErrTriggerNotFound
	}

	if err != nil {
		return nil, "", fmt.Errorf("failed to find due trigger for update: %w", err)
	}

	return trigger, taskStatus, nil
}
