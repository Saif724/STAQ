package actions

import "errors"

var ErrExecutorNotFound = errors.New("executor not found")

type Registry struct {
	executors map[string]Executor
}

func NewRegistry() *Registry {
	return &Registry{
		executors: make(map[string]Executor),
	}
}

func (r *Registry) Register(
	actionType string,
	executor Executor,
) {
	r.executors[actionType] = executor
}

func (r *Registry) Get(
	actionType string,
) (Executor, error) {
	executor, ok := r.executors[actionType]
	if !ok {
		return nil, ErrExecutorNotFound
	}

	return executor, nil
}
