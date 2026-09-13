package tasks

import "time"

const (
	StatusActive   = "Active"
	StatusPaused   = "Paused"
	StatusArchived = "Archived"
)

type Task struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	QueueID        string    `json:"queue_id"`
	Name           string    `json:"name"`
	Description    *string   `json:"description,omitempty"`
	Status         string    `json:"status"`
	TimeoutSeconds int       `json:"timeout_seconds"`
	MaxRetries     int       `json:"max_retries"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
