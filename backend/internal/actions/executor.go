package actions

import "context"

type ExecutionContext struct {
	TaskID   string
	ActionID string
	UserID   string
}

type ExecutionResult struct {
	Message string
	Data    any
}

type Executor interface {
	Execute(
		ctx context.Context,
		executionContext ExecutionContext,
		configuration []byte,
	) (*ExecutionResult, error)
}
