package tasksetups

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Saif724/STAQ/backend/internal/actions"
	"github.com/Saif724/STAQ/backend/internal/tasks"
	"github.com/Saif724/STAQ/backend/internal/triggers"
	"github.com/jackc/pgx/v5/pgxpool"
)

const maxSetupActions = 50

var ErrInvalidSetup = errors.New("invalid task setup")

type ActionInput struct {
	ActionType        string          `json:"action_type"`
	ExecutionOrder    int             `json:"execution_order"`
	Configuration     json.RawMessage `json:"configuration"`
	ContinueOnFailure bool            `json:"continue_on_failure"`
}

type CreateInput struct {
	Name           string        `json:"name"`
	Description    *string       `json:"description"`
	TimeoutSeconds int           `json:"timeout_seconds"`
	MaxRetries     int           `json:"max_retries"`
	TriggerType    string        `json:"trigger_type"`
	CronExpression *string       `json:"cron_expression"`
	Timezone       string        `json:"timezone"`
	StartAt        time.Time     `json:"start_at"`
	Actions        []ActionInput `json:"actions"`
}

type Result struct {
	Task    *tasks.Task       `json:"task"`
	Trigger *triggers.Trigger `json:"trigger"`
	Actions []*actions.Action `json:"actions"`
}

type Service struct {
	db             *pgxpool.Pool
	taskService    *tasks.Service
	triggerService *triggers.Service
	actionService  *actions.Service
}

func NewService(
	db *pgxpool.Pool,
	taskService *tasks.Service,
	triggerService *triggers.Service,
	actionService *actions.Service,
) *Service {
	return &Service{
		db:             db,
		taskService:    taskService,
		triggerService: triggerService,
		actionService:  actionService,
	}
}

func (s *Service) Create(ctx context.Context, userID string, input CreateInput) (*Result, error) {
	if len(input.Actions) == 0 {
		return nil, fmt.Errorf("%w: at least one action is required", ErrInvalidSetup)
	}
	if len(input.Actions) > maxSetupActions {
		return nil, fmt.Errorf("%w: no more than %d actions are allowed", ErrInvalidSetup, maxSetupActions)
	}

	task, err := s.taskService.PrepareCreate(ctx, tasks.CreateTaskInput{
		UserID:         userID,
		Name:           input.Name,
		Description:    input.Description,
		TimeoutSeconds: input.TimeoutSeconds,
		MaxRetries:     input.MaxRetries,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: task: %v", ErrInvalidSetup, err)
	}

	trigger, err := s.triggerService.PrepareCreateForTask(userID, triggers.CreateTriggerInput{
		TaskID:         task.ID,
		TriggerType:    input.TriggerType,
		CronExpression: input.CronExpression,
		Timezone:       input.Timezone,
		StartAt:        input.StartAt,
	}, task)
	if err != nil {
		return nil, fmt.Errorf("%w: trigger: %v", ErrInvalidSetup, err)
	}

	preparedActions := make([]*actions.Action, 0, len(input.Actions))
	seenOrders := make(map[int]struct{}, len(input.Actions))
	for _, actionInput := range input.Actions {
		if _, exists := seenOrders[actionInput.ExecutionOrder]; exists {
			return nil, fmt.Errorf("%w: action execution orders must be unique", ErrInvalidSetup)
		}
		seenOrders[actionInput.ExecutionOrder] = struct{}{}

		action, err := s.actionService.PrepareCreateForTask(userID, actions.CreateActionInput{
			TaskID:            task.ID,
			ActionType:        actionInput.ActionType,
			ExecutionOrder:    actionInput.ExecutionOrder,
			Configuration:     actionInput.Configuration,
			ContinueOnFailure: actionInput.ContinueOnFailure,
		}, task)
		if err != nil {
			return nil, fmt.Errorf("%w: action at order %d: %v", ErrInvalidSetup, actionInput.ExecutionOrder, err)
		}
		preparedActions = append(preparedActions, action)
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin task setup transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := s.taskService.CreateWithTx(ctx, tx, task); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}
	if err := s.triggerService.CreateWithTx(ctx, tx, trigger); err != nil {
		return nil, fmt.Errorf("failed to create trigger: %w", err)
	}
	for _, action := range preparedActions {
		if err := s.actionService.CreateWithTx(ctx, tx, action); err != nil {
			return nil, fmt.Errorf("failed to create action at order %d: %w", action.ExecutionOrder, err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit task setup: %w", err)
	}

	return &Result{
		Task:    task,
		Trigger: trigger,
		Actions: preparedActions,
	}, nil
}
