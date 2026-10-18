package reminder

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Saif724/STAQ/backend/internal/actions"
)

type Notifier interface {
	CreateReminder(ctx context.Context, userID, taskID, executionID, actionID, message string) error
}

type Executor struct {
	notifier Notifier
}

func NewExecutor(notifier Notifier) *Executor {
	return &Executor{notifier: notifier}
}

type Configuration struct {
	Message string `json:"message"`
}

func (e *Executor) Execute(
	ctx context.Context,
	executionContext actions.ExecutionContext,
	configuration []byte,
) (*actions.ExecutionResult, error) {
	var config Configuration

	if err := json.Unmarshal(configuration, &config); err != nil {
		return nil, errors.New("invalid reminder configuration")
	}

	message := strings.TrimSpace(config.Message)
	if message == "" {
		return nil, errors.New("reminder message is required")
	}

	if e.notifier == nil {
		return nil, errors.New("reminder notifier is not configured")
	}

	if err := e.notifier.CreateReminder(
		ctx,
		executionContext.UserID,
		executionContext.TaskID,
		executionContext.ExecutionID,
		executionContext.ActionID,
		message,
	); err != nil {
		return nil, actions.Retryable(fmt.Errorf("failed to save reminder: %w", err))
	}

	return &actions.ExecutionResult{
		Message: "reminder delivered",
		Data: map[string]any{
			"message": message,
		},
	}, nil
}
