package shell

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Saif724/STAQ/backend/internal/actions"
)

type Executor struct {
	allowedCommands map[string]struct{}
}

func NewExecutor(allowedCommands []string) *Executor {
	commands := make(map[string]struct{})

	for _, command := range allowedCommands {
		commands[command] = struct{}{}
	}
	return &Executor{
		allowedCommands: commands,
	}
}

type Configuration struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

func (e *Executor) Execute(
	ctx context.Context,
	executionContext actions.ExecutionContext,
	configuration []byte,
) (*actions.ExecutionResult, error) {
	var config Configuration

	if err := json.Unmarshal(configuration, &config); err != nil {
		return nil, errors.New("invalid shell configuraion")
	}

	if config.Command == "" {
		return nil, errors.New("shell command is required")
	}

	if _, allowed := e.allowedCommands[config.Command]; !allowed {
		return nil, fmt.Errorf("shell command is not allowed: %s", config.Command)
	}

	return &actions.ExecutionResult{
		Message: "shell action validated",
		Data: map[string]any{
			"command": config.Command,
			"args":    config.Args,
		},
	}, nil
}
