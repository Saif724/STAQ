package notifications

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// Create inserts a notification. A repeated (execution_id, action_id) pair is
// ignored, so a retried or reclaimed execution cannot remind twice.
func (r *Repository) Create(ctx context.Context, n *Notification) error {
	const query = `
		INSERT INTO notifications (
			id, user_id, task_id, execution_id, action_id,
			type, message, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (execution_id, action_id) DO NOTHING
	`

	_, err := r.db.Exec(
		ctx, query,
		n.ID, n.UserID, n.TaskID, n.ExecutionID, n.ActionID,
		n.Type, n.Message, n.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	return nil
}

func (r *Repository) ListByUser(
	ctx context.Context,
	userID string,
	limit int,
) ([]Notification, error) {
	const query = `
		SELECT id, user_id, task_id, execution_id, action_id,
		       type, message, read_at, created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list notifications: %w", err)
	}
	defer rows.Close()

	items := make([]Notification, 0)

	for rows.Next() {
		var n Notification

		if err := rows.Scan(
			&n.ID, &n.UserID, &n.TaskID, &n.ExecutionID, &n.ActionID,
			&n.Type, &n.Message, &n.ReadAt, &n.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}

		items = append(items, n)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate notifications: %w", err)
	}

	return items, nil
}

func (r *Repository) CountUnread(ctx context.Context, userID string) (int, error) {
	const query = `
		SELECT COUNT(*)
		FROM notifications
		WHERE user_id = $1 AND read_at IS NULL
	`

	var count int
	if err := r.db.QueryRow(ctx, query, userID).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count unread notifications: %w", err)
	}

	return count, nil
}

// MarkRead is idempotent: marking an already-read notification still succeeds.
func (r *Repository) MarkRead(ctx context.Context, id, userID string) (bool, error) {
	const query = `
		UPDATE notifications
		SET read_at = COALESCE(read_at, $1)
		WHERE id = $2 AND user_id = $3
	`

	result, err := r.db.Exec(ctx, query, time.Now().UTC(), id, userID)
	if err != nil {
		return false, fmt.Errorf("failed to mark notification read: %w", err)
	}

	return result.RowsAffected() == 1, nil
}

func (r *Repository) MarkAllRead(ctx context.Context, userID string) error {
	const query = `
		UPDATE notifications
		SET read_at = $1
		WHERE user_id = $2 AND read_at IS NULL
	`

	if _, err := r.db.Exec(ctx, query, time.Now().UTC(), userID); err != nil {
		return fmt.Errorf("failed to mark notifications read: %w", err)
	}

	return nil
}
