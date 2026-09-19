package email

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/Saif724/STAQ/backend/internal/actions"
)

type Sender interface {
	SendEmail(
		ctx context.Context,
		userID string,
		connectionID string,
		to string,
		subject string,
		body string,
	) error
}

type Executor struct {
	sender Sender
}

func NewExecutor(senders ...Sender) *Executor {
	var sender Sender

	if len(senders) > 0 {
		sender = senders[0]
	}
	return &Executor{
		sender: sender,
	}
}

type Configuration struct {
	ConnectionID string `json:"connection_id"`
	To           string `json:"to"`
	Subject      string `json:"subject"`
	Body         string `json:"body"`
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

	config.ConnectionID = strings.TrimSpace(config.ConnectionID)
	config.To = strings.TrimSpace(config.To)
	config.Subject = strings.TrimSpace(config.Subject)

	if config.ConnectionID == "" {
		return nil, errors.New("email connection id is required")
	}

	if config.To == "" {
		return nil, errors.New("email receipient is required")
	}

	if config.Subject == "" {
		return nil, errors.New("email subject is required")
	}

	if strings.TrimSpace(config.Body) == "" {
		return nil, errors.New("email body is required")
	}

	if e.sender == nil {
		return nil, errors.New("email sender is not configured")
	}

	err := e.sender.SendEmail(
		ctx,
		executionContext.UserID,
		config.ConnectionID,
		config.To,
		config.Subject,
		config.Body,
	)
	if err != nil {
		return nil, err
	}

	return &actions.ExecutionResult{
		Message: "email sent successfully",
		Data: map[string]any{
			"to":      config.To,
			"subject": config.Subject,
		},
	}, nil
}
