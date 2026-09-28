package executions

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrExecutionNotFound     = errors.New("execution not found")
	ErrExecutionAlreadyExist = errors.New("execution already exists")
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
	execution *Execution,
) error {
	const query = `
		INSERT INTO  executions (
			id,
			task_id,
			trigger_id,
			scheduled_at,
			status,
			started_at,
			heartbeat_at,
			completed_at,
			duration_ms,
			retry_count,
			error_message,
			created_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11, $12
		)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		execution.ID,
		execution.TaskID,
		execution.TriggerID,
		execution.ScheduledAt,
		execution.Status,
		execution.StartedAt,
		execution.HeartbeatAt,
		execution.CompletedAt,
		execution.DurationMs,
		execution.RetryCount,
		execution.ErrorMessage,
		execution.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" &&
			pgErr.ConstraintName == "uq_executions_trigger_scheduled" {
			return ErrExecutionAlreadyExist
		}

		return fmt.Errorf("failed to create execution: %w", err)
	}

	return nil
}

func (r *Repository) FindByID(
	ctx context.Context,
	executionID string,
) (*Execution, error) {
	const query = `
		SELECT
			id,
			task_id,
			trigger_id,
			scheduled_at,
			status,
			started_at,
			heartbeat_at,
			completed_at,
			duration_ms,
			retry_count,
			error_message,
			created_at
		FROM executions
		WHERE id = $1
	`

	var execution Execution

	err := r.db.QueryRow(
		ctx,
		query,
		executionID,
	).Scan(
		&execution.ID,
		&execution.TaskID,
		&execution.TriggerID,
		&execution.ScheduledAt,
		&execution.Status,
		&execution.StartedAt,
		&execution.HeartbeatAt,
		&execution.CompletedAt,
		&execution.DurationMs,
		&execution.RetryCount,
		&execution.ErrorMessage,
		&execution.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrExecutionNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to find execution: %w", err)
	}

	return &execution, nil
}

func (r *Repository) FindByTriggerAndScheduledAt(
	ctx context.Context,
	triggerID string,
	scheduledAt interface{},
) (*Execution, error) {
	const query = `
		SELECT
			id,
			task_id,
			trigger_id,
			scheduled_at,
			status,
			started_at,
			heartbeat_at,
			completed_at,
			duration_ms,
			retry_count,
			error_message,
			created_at
		FROM executions
		WHERE trigger_id = $1
			AND scheduled_at = $2
	`

	var execution Execution

	err := r.db.QueryRow(
		ctx,
		query,
		triggerID,
		scheduledAt,
	).Scan(
		&execution.ID,
		&execution.TaskID,
		&execution.TriggerID,
		&execution.ScheduledAt,
		&execution.Status,
		&execution.StartedAt,
		&execution.HeartbeatAt,
		&execution.CompletedAt,
		&execution.DurationMs,
		&execution.RetryCount,
		&execution.ErrorMessage,
		&execution.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrExecutionNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to find execution by trigger and scheduled time: %w", err)
	}

	return &execution, nil
}

func (r *Repository) FindByIDAndUser(
	ctx context.Context,
	executionID string,
	userID string,
) (*Execution, error) {
	const query = `
		SELECT
			e.id,
			e.task_id,
			e.trigger_id,
			e.scheduled_at,
			e.status,
			e.started_at,
			e.heartbeat_at,
			e.completed_at,
			e.duration_ms,
			e.retry_count,
			e.error_message,
			e.created_at
		FROM executions e
		INNER JOIN tasks t ON t.id = e.task_id
		WHERE e.id = $1
			AND t.user_id = $2
	`

	var execution Execution

	err := r.db.QueryRow(
		ctx,
		query,
		executionID,
		userID,
	).Scan(
		&execution.ID,
		&execution.TaskID,
		&execution.TriggerID,
		&execution.ScheduledAt,
		&execution.Status,
		&execution.StartedAt,
		&execution.HeartbeatAt,
		&execution.CompletedAt,
		&execution.DurationMs,
		&execution.RetryCount,
		&execution.ErrorMessage,
		&execution.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrExecutionNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("failed to find execution: %w", err)
	}

	return &execution, nil
}

func (r *Repository) Update(
	ctx context.Context,
	execution *Execution,
) error {
	const query = `
		UPDATE executions
		SET
			status = $1,
			heartbeat_at = $2,
			completed_at = $3,
			duration_ms = $4,
			retry_count = $5,
			error_message = $6
		WHERE id = $7
	`

	result, err := r.db.Exec(
		ctx,
		query,
		execution.Status,
		execution.HeartbeatAt,
		execution.CompletedAt,
		execution.DurationMs,
		execution.RetryCount,
		execution.ErrorMessage,
		execution.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update execution: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrExecutionNotFound
	}

	return nil
}

func (r *Repository) IncrementRetryCount(
	ctx context.Context,
	executionID string,
) error {
	const query = `
		UPDATE executions
		SET retry_count = retry_count + 1
		WHERE id = $1
			AND status = 'RUNNING'
	`

	result, err := r.db.Exec(
		ctx,
		query,
		executionID,
	)

	if err != nil {
		return fmt.Errorf("failed to increment execution retry count: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrExecutionNotFound
	}

	return nil
}

func (r *Repository) CreateLog(
	ctx context.Context,
	log *ExecutionLog,
) error {
	const query = `
		INSERT INTO  execution_logs (
			id,
			execution_id,
			log_level,
			message,
			created_at
		)
		VALUES (
			$1, $2, $3, $4, $5
		)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		log.ID,
		log.ExecutionID,
		log.LogLevel,
		log.Message,
		log.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create execution log: %w", err)
	}

	return nil
}

func (r *Repository) FindLogs(
	ctx context.Context,
	executionID string,
) ([]ExecutionLog, error) {
	const query = `
		SELECT
			id,
			execution_id,
			log_level,
			message,
			created_at
		FROM execution_logs
		WHERE execution_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query, executionID)
	if err != nil {
		return nil, fmt.Errorf("failed to find execution logs: %w", err)
	}
	defer rows.Close()

	logs := make([]ExecutionLog, 0)

	for rows.Next() {
		var log ExecutionLog

		if err := rows.Scan(
			&log.ID,
			&log.ExecutionID,
			&log.LogLevel,
			&log.Message,
			&log.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan execution log: %w", err)
		}

		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate execution logs: %w", err)
	}

	return logs, nil
}

func (r *Repository) FindLogsByUser(
	ctx context.Context,
	executionID string,
	userID string,
) ([]ExecutionLog, error) {
	const query = `
		SELECT
			el.id,
			el.execution_id,
			el.log_level,
			el.message,
			el.created_at
		FROM execution_logs el
		INNER JOIN executions e ON e.id = el.execution_id
		INNER JOIN tasks t ON t.id = e.task_id
		WHERE el.execution_id = $1
			AND t.user_id = $2
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query, executionID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fing execution logs: %w", err)
	}
	defer rows.Close()

	logs := make([]ExecutionLog, 0)

	for rows.Next() {
		var log ExecutionLog

		if err := rows.Scan(
			&log.ID,
			&log.ExecutionID,
			&log.LogLevel,
			&log.Message,
			&log.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan execution log: %w", err)
		}

		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate execution logs: %w", err)
	}

	return logs, nil
}

func (r *Repository) Heartbeat(
	ctx context.Context,
	executionID string,
	heartbeatAt time.Time,
) error {
	const query = `
		UPDATE executions
		SET heartbeat_at = $1
		WHERE id = $2
			AND status = 'RUNNING'
	`

	result, err := r.db.Exec(
		ctx,
		query,
		heartbeatAt,
		executionID,
	)
	if err != nil {
		return fmt.Errorf("failed to update execution heartbeat: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrExecutionNotFound
	}

	return nil
}

func (r *Repository) Complete(
	ctx context.Context,
	execution *Execution,
) error {
	const query = `
		UPDATE executions
		SET
			status = $1,
			heartbeat_at = NULL,
			completed_at = $2,
			duration_ms = $3,
			retry_count = $4,
			error_message = $5
		WHERE id = $6
			AND status = 'RUNNING'
	`

	result, err := r.db.Exec(
		ctx,
		query,
		execution.Status,
		execution.CompletedAt,
		execution.DurationMs,
		execution.RetryCount,
		execution.ErrorMessage,
		execution.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to complete execution: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrExecutionNotFound
	}

	return nil
}

func (r *Repository) RecoveryStaleRunning(
	ctx context.Context,
	before time.Time,
) ([]string, error) {
	const query = `
		UPDATE executions
		SET
			status = 'FAILED',
			completed_at = $1,
			error_message = 'worker heartbeat expired',
			heartbeat_at = NULL,
			duration_ms = EXTRACT(
				EPOCH FROM ($1 - started_at)
			) * 1000
		WHERE status = 'RUNNING'
			AND heartbeat_at is NOT NULL
			AND heartbeat_at < $2
		RETURNING id
	`

	rows, err := r.db.Query(
		ctx,
		query,
		time.Now().UTC(),
		before,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to recover stale executions: %w", err)
	}
	defer rows.Close()

	var executionIDs []string

	for rows.Next() {
		var executionID string

		if err := rows.Scan(&executionID); err != nil {
			return nil, fmt.Errorf("failed to scan recovered execution: %w", err)
		}

		executionIDs = append(executionIDs, executionID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate recovered executions: %w", err)
	}

	return executionIDs, nil
}

func (r *Repository) ReclaimExecution(
	ctx context.Context,
	executionID string,
	heartbeatAt time.Time,
) (bool, error) {
	const query = `
		UPDATE executions
		SET
			heartbeat_at = $1
		WHERE id = $2
			AND status = 'RUNNING'
			AND heartbeat_at is NOT NULL
			AND heartbeat_at < $1 - INTERVAL '2 minutes'
	`
	result, err := r.db.Exec(
		ctx,
		query,
		heartbeatAt,
		executionID,
	)
	if err != nil {
		return false, fmt.Errorf("failed to reclaim execution: %w", err)
	}

	return result.RowsAffected() == 1, nil
}
