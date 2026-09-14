package email

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
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

func (e *Executor) Execute(
	ctx context.Context,
	executionContext actions.ExecutionContext,
	configuration []byte,
) (*actions.ExecutionResult, error) {
	var config Configuration

	if err := json.Unmarshal(configuration, &config); err != nil {
		return nil, errors.New("invalid email configuraion")
	}

	if config.To == "" {
		return nil, errors.New("email receipient is required")
	}

	if config.Subject == "" {
		return nil, errors.New("email subject is required")
	}

	if config.Body == "" {
		return nil, errors.New("email body is required")
	}

	return &actions.ExecutionResult{
		Message: "email action prepared",
		Data: map[string]any{
			"to":      config.To,
			"subject": config.Subject,
		},
	}, nil
}
