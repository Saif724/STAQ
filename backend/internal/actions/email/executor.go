package email

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/Saif724/STAQ/backend/internal/actions"
)

const (
	SenderSTAQ      = "STAQ"
	SenderUserGmail = "USER_GMAIL"
)

type PlatformSender interface {
	Send(
		to string,
		subject string,
		body string,
	) error
}

type UserSender interface {
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
	platformSender PlatformSender
	userSender     UserSender
}

func NewExecutor(
	platformSender PlatformSender,
	userSender UserSender,
) *Executor {
	return &Executor{
		platformSender: platformSender,
		userSender:     userSender,
	}
}

type Configuration struct {
	Sender       string `json:"sender"`
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
		return nil, errors.New("invalid email configuration")
	}

	config.Sender = strings.ToUpper(strings.TrimSpace(config.Sender))
	config.ConnectionID = strings.TrimSpace(config.ConnectionID)
	config.To = strings.TrimSpace(config.To)
	config.Subject = strings.TrimSpace(config.Subject)

	if config.Sender == "" {
		return nil, errors.New("email sender is required")
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

	switch config.Sender {
	case SenderSTAQ:
		if e.platformSender == nil {
			return nil, errors.New("STAQ email sender is not configured")
		}

		if err := e.platformSender.Send(
			config.To,
			config.Subject,
			config.Body,
		); err != nil {
			return nil, err
		}

	case SenderUserGmail:
		if e.userSender == nil {
			return nil, errors.New("user Gmail sender is not configured")
		}

		if config.ConnectionID == "" {
			return nil, errors.New("email connection id is required for USER_GMAIL")
		}

		if err := e.userSender.SendEmail(
			ctx,
			executionContext.UserID,
			config.ConnectionID,
			config.To,
			config.Subject,
			config.Body,
		); err != nil {
			return nil, err
		}

	default:
		return nil, errors.New("unsupported email sender")
	}

	return &actions.ExecutionResult{
		Message: "email sent successfully",
		Data: map[string]any{
			"sender":  config.Sender,
			"to":      config.To,
			"subject": config.Subject,
		},
	}, nil
}
