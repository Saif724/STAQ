package executions

import "time"

const (
	StatusPending   = "PENDING"
	StatusRunning   = "RUNNING"
	StatusSuccess   = "SUCCESS"
	StatusFailed    = "FAILED"
	StatusCancelled = "CANCELLED"
	StatusTimedOut  = "TIMED_OUT"
)

type Execution struct {
	ID           string     `json:"id"`
	TaskID       string     `json:"task_id"`
	TriggerID    string     `json:"trigger_id"`
	ScheduledAt  time.Time  `json:"scheduled_at"`
	Status       string     `json:"status"`
	StartedAt    time.Time  `json:"started_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	DurationMs   *int64     `json:"duration_ms,omitempty"`
	RetryCount   int        `json:"retry_count"`
	ErrorMessage *string    `json:"error_message,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

type ExecutionLog struct {
	ID          string    `json:"id"`
	ExecutionID string    `json:"execution_id"`
	LogLevel    string    `json:"log_level"`
	Message     string    `json:"message"`
	CreatedAt   time.Time `json:"created_at"`
}

const (
	LogInfo    = "INFO"
	LogWarning = "WARNING"
	LogError   = "ERROR"
)
