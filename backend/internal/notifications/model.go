package notifications

import "time"

const TypeReminder = "REMINDER"

type Notification struct {
	ID          string     `json:"id"`
	UserID      string     `json:"-"`
	TaskID      *string    `json:"task_id,omitempty"`
	ExecutionID *string    `json:"execution_id,omitempty"`
	ActionID    *string    `json:"action_id,omitempty"`
	Type        string     `json:"type"`
	Message     string     `json:"message"`
	ReadAt      *time.Time `json:"read_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}
