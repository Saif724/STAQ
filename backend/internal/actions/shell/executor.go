package shell

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/Saif724/STAQ/backend/internal/actions"
)

const maxOutputSize = 64 * 1024

type Executor struct {
	allowedCommands map[string]struct{}
}

func NewExecutor(allowedCommands []string) *Executor {
	commands := make(map[string]struct{})

	for _, command := range allowedCommands {
		command = strings.TrimSpace(command)
		if command != "" {
			commands[command] = struct{}{}
		}
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
		return nil, errors.New("invalid shell configuration")
	}

	command := strings.TrimSpace(config.Command)

	if command == "" {
		return nil, errors.New("shell command is required")
	}

	if _, allowed := e.allowedCommands[command]; !allowed {
		return nil, fmt.Errorf("shell command is not allowed: %s", config.Command)
	}

	cmd := exec.CommandContext(ctx, command, config.Args...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	output := stdout.String()
	errorOutput := stderr.String()

	if len(output) > maxOutputSize {
		output = output[:maxOutputSize]
	}

	if len(errorOutput) > maxOutputSize {
		errorOutput = errorOutput[:maxOutputSize]
	}

	resultData := map[string]any{
		"command": config.Command,
		"args":    config.Args,
		"stdout":  output,
		"stderr":  errorOutput,
	}

	if err != nil {
		return &actions.ExecutionResult{
			Message: "shell action failed",
			Data:    resultData,
		}, fmt.Errorf("shell command failed: %w", err)
	}

	return &actions.ExecutionResult{
		Message: "shell action validated",
		Data:    resultData,
	}, nil
}
