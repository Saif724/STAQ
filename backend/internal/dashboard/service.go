package dashboard

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Saif724/STAQ/backend/internal/executions"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	SummaryWindow = 30 * 24 * time.Hour
	SummaryLimit  = 5
)

type Summary struct {
	WindowStartsAt     time.Time         `json:"window_starts_at"`
	WindowDays         int               `json:"window_days"`
	SuccessfulCount    int64             `json:"successful_count"`
	FailedCount        int64             `json:"failed_count"`
	TimedOutCount      int64             `json:"timed_out_count"`
	CancelledCount     int64             `json:"cancelled_count"`
	PendingCount       int64             `json:"pending_count"`
	RunningCount       int64             `json:"running_count"`
	SuccessRate        float64           `json:"success_rate"`
	RecentExecutions   []RecentExecution `json:"recent_executions"`
	RecentFailures     []RecentExecution `json:"recent_failures"`
	UpcomingExecutions []UpcomingRun     `json:"upcoming_executions"`
}

type RecentExecution struct {
	executions.Execution
	TaskName string `json:"task_name"`
}

type UpcomingRun struct {
	TaskID      string    `json:"task_id"`
	TaskName    string    `json:"task_name"`
	TriggerID   string    `json:"trigger_id"`
	TriggerType string    `json:"trigger_type"`
	NextRunAt   time.Time `json:"next_run_at"`
	Timezone    string    `json:"timezone"`
}

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

func (s *Service) GetForUser(ctx context.Context, userID string) (*Summary, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, errors.New("user id is required")
	}

	now := time.Now().UTC()
	windowStartsAt := now.Add(-SummaryWindow)
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadOnly,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to begin dashboard snapshot: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	summary := &Summary{
		WindowStartsAt:     windowStartsAt,
		WindowDays:         30,
		RecentExecutions:   make([]RecentExecution, 0),
		RecentFailures:     make([]RecentExecution, 0),
		UpcomingExecutions: make([]UpcomingRun, 0),
	}

	const countsQuery = `
		SELECT
			COUNT(*) FILTER (WHERE e.status = 'SUCCESS'),
			COUNT(*) FILTER (WHERE e.status = 'FAILED'),
			COUNT(*) FILTER (WHERE e.status = 'TIMED_OUT'),
			COUNT(*) FILTER (WHERE e.status = 'CANCELLED'),
			COUNT(*) FILTER (WHERE e.status = 'PENDING'),
			COUNT(*) FILTER (WHERE e.status = 'RUNNING')
		FROM executions e
		INNER JOIN tasks t ON t.id = e.task_id
		WHERE t.user_id = $1
			AND e.created_at >= $2
	`
	if err := tx.QueryRow(ctx, countsQuery, userID, windowStartsAt).Scan(
		&summary.SuccessfulCount,
		&summary.FailedCount,
		&summary.TimedOutCount,
		&summary.CancelledCount,
		&summary.PendingCount,
		&summary.RunningCount,
	); err != nil {
		return nil, fmt.Errorf("failed to read dashboard execution counts: %w", err)
	}
	summary.SuccessRate = calculateSuccessRate(
		summary.SuccessfulCount,
		summary.FailedCount,
		summary.TimedOutCount,
	)

	if summary.RecentExecutions, err = listRecent(ctx, tx, userID, false); err != nil {
		return nil, err
	}
	if summary.RecentFailures, err = listRecent(ctx, tx, userID, true); err != nil {
		return nil, err
	}
	if summary.UpcomingExecutions, err = listUpcoming(ctx, tx, userID, now); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to finish dashboard snapshot: %w", err)
	}
	return summary, nil
}

func listRecent(
	ctx context.Context,
	tx pgx.Tx,
	userID string,
	failuresOnly bool,
) ([]RecentExecution, error) {
	where := "t.user_id = $1"
	if failuresOnly {
		where += " AND e.status IN ('FAILED', 'TIMED_OUT')"
	}
	query := `
		SELECT
			e.id, e.task_id, e.trigger_id, e.scheduled_at, e.status,
			e.started_at, e.heartbeat_at, e.completed_at, e.duration_ms,
			e.retry_count, e.error_message, e.created_at, t.name
		FROM executions e
		INNER JOIN tasks t ON t.id = e.task_id
		WHERE ` + where + `
		ORDER BY e.created_at DESC, e.id DESC
		LIMIT $2
	`
	rows, err := tx.Query(ctx, query, userID, SummaryLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to list recent dashboard executions: %w", err)
	}
	defer rows.Close()

	items := make([]RecentExecution, 0, SummaryLimit)
	for rows.Next() {
		var item RecentExecution
		if err := rows.Scan(
			&item.ID,
			&item.TaskID,
			&item.TriggerID,
			&item.ScheduledAt,
			&item.Status,
			&item.StartedAt,
			&item.HeartbeatAt,
			&item.CompletedAt,
			&item.DurationMs,
			&item.RetryCount,
			&item.ErrorMessage,
			&item.CreatedAt,
			&item.TaskName,
		); err != nil {
			return nil, fmt.Errorf("failed to scan recent dashboard execution: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate recent dashboard executions: %w", err)
	}
	return items, nil
}

func listUpcoming(
	ctx context.Context,
	tx pgx.Tx,
	userID string,
	now time.Time,
) ([]UpcomingRun, error) {
	const query = `
		SELECT t.id, t.name, tr.id, tr.trigger_type, tr.next_run_at, tr.timezone
		FROM triggers tr
		INNER JOIN tasks t ON t.id = tr.task_id
		WHERE t.user_id = $1
			AND t.status = 'Active'
			AND tr.is_active = TRUE
			AND tr.next_run_at > $2
		ORDER BY tr.next_run_at ASC, tr.id ASC
		LIMIT $3
	`
	rows, err := tx.Query(ctx, query, userID, now, SummaryLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to list upcoming dashboard runs: %w", err)
	}
	defer rows.Close()

	items := make([]UpcomingRun, 0, SummaryLimit)
	for rows.Next() {
		var item UpcomingRun
		if err := rows.Scan(
			&item.TaskID,
			&item.TaskName,
			&item.TriggerID,
			&item.TriggerType,
			&item.NextRunAt,
			&item.Timezone,
		); err != nil {
			return nil, fmt.Errorf("failed to scan upcoming dashboard run: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate upcoming dashboard runs: %w", err)
	}
	return items, nil
}

func calculateSuccessRate(successful, failed, timedOut int64) float64 {
	completed := successful + failed + timedOut
	if completed == 0 {
		return 0
	}
	return float64(successful) / float64(completed) * 100
}
