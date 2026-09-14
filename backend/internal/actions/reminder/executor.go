package reminder

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/Saif724/STAQ/backend/internal/actions"
)

type Executor struct{}

func NewExecutor() *Executor {
	return &Executor{}
}

type Configuration struct {
	Message string `json:"message"`
}

func (s *Executor) Execute(
	ctx context.Context,
	executionContext actions.ExecutionContext,
	configuration []byte,
) (*actions.ExecutionResult, error) {
	var config Configuration

	if err := json.Unmarshal(configuration, &config); err != nil {
		return nil, errors.New("invalid reminder configuraion")
	}

	if config.Message == "" {
		return nil, errors.New("reminder message is required")
	}

	return &actions.ExecutionResult{
		Message: "reminder action executed",
		Data: map[string]any{
			"message": config.Message,
		},
	}, nil
}
