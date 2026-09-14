package actions

import "time"

const (
	TypeReminder = "REMINDER"
	TypeEmail    = "EMAIL"
	TypeHTTP     = "HTTP"
	TypeShell    = "SHELL"
)

type Action struct {
	ID                string    `json:"id"`
	TaskID            string    `json:"task_id"`
	ActionType        string    `json:"action_type"`
	ExecutionOrder    int       `json:"execution_order"`
	Configuration     []byte    `json:"configuration"`
	ContinueOnFailure bool      `json:"continue_on_failure"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
