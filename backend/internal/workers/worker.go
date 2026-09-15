package workers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Saif724/STAQ/backend/internal/actions"
	"github.com/Saif724/STAQ/backend/internal/actions/email"
	httpaction "github.com/Saif724/STAQ/backend/internal/actions/http"
	"github.com/Saif724/STAQ/backend/internal/actions/reminder"
	"github.com/Saif724/STAQ/backend/internal/actions/shell"
	"github.com/Saif724/STAQ/backend/internal/broker"
	"github.com/Saif724/STAQ/backend/internal/executions"
	"github.com/Saif724/STAQ/backend/internal/tasks"
	"github.com/rs/zerolog"
)

type Worker struct {
	broker           *broker.Redis
	taskRepository   *tasks.Repository
	actionRepository *actions.Repository
	executionService *executions.Service
	registry         *actions.Registry
	logger           zerolog.Logger
}

func New(
	broker *broker.Redis,
	taskRepository *tasks.Repository,
	actionRepository *actions.Repository,
	executionService *executions.Service,
	registry *actions.Registry,
	logger zerolog.Logger,
) *Worker {
	return &Worker{
		broker:           broker,
		taskRepository:   taskRepository,
		actionRepository: actionRepository,
		executionService: executionService,
		registry:         registry,
		logger:           logger,
	}
}

func NewRegistry() *actions.Registry {
	registry := actions.NewRegistry()

	registry.Register(
		actions.TypeReminder,
		reminder.NewExecutor(),
	)

	registry.Register(
		actions.TypeEmail,
		email.NewExecutor(),
	)

	registry.Register(
		actions.TypeHTTP,
		httpaction.NewExecutor(
			&http.Client{
				Timeout: 30 * time.Second,
			},
		),
	)

	registry.Register(
		actions.TypeShell,
		shell.NewExecutor(
			[]string{},
		),
	)

	return registry
}

func (w *Worker) Run(ctx context.Context) error {
	w.logger.Info().Msg("worker started")

	for {
		if err := w.processOne(ctx); err != nil {
			if errors.Is(err, context.Canceled) ||
				errors.Is(err, context.DeadlineExceeded) {
				return err
			}

			w.logger.Error().
				Err(err).
				Msg("failed to process task job")
		}
	}
}

func (w *Worker) processOne(ctx context.Context) error {
	result, err := w.broker.Client().BRPop(
		ctx,
		0,
		broker.TaskQueue,
	).Result()

	if err != nil {
		return err
	}

	if len(result) != 2 {
		return fmt.Errorf("invalid redis task response")
	}

	var job broker.TaskJob

	if err := json.Unmarshal(
		[]byte(result[1]),
		&job,
	); err != nil {
		return fmt.Errorf("failed to decode task job: %w", err)
	}

	return w.executeJob(ctx, job)
}

func (w *Worker) executeJob(
	ctx context.Context,
	job broker.TaskJob,
) error {
	if strings.TrimSpace(job.TaskID) == "" {
		return errors.New("task job has empty task id")
	}

	if strings.TrimSpace(job.TriggerID) == "" {
		return errors.New("task job has empty trigger id")
	}

	task, err := w.taskRepository.FindByID(ctx, job.TaskID)
	if err != nil {
		return fmt.Errorf("failed to load task: %w", err)
	}

	if task.Status != tasks.StatusActive {
		w.logger.Warn().
			Str("task_id", task.ID).
			Str("status", task.Status).
			Msg("skipping inactive task")

		return nil
	}

	execution, err := w.executionService.Start(
		ctx,
		task.ID,
		job.TriggerID,
	)
	if err != nil {
		return fmt.Errorf("failed to start execution: %w", err)
	}

	w.logger.Info().
		Str("execution_id", execution.ID).
		Str("task_id", task.ID).
		Str("trigger_id", job.TriggerID).
		Msg("execution started")

	if err := w.executionService.Log(
		ctx,
		execution.ID,
		executions.LogInfo,
		"execution started",
	); err != nil {
		w.logger.Warn().
			Err(err).
			Str("execution_id", execution.ID).
			Msg("failed to create execution log")
	}

	if err := w.executeActions(
		ctx,
		task,
		execution,
	); err != nil {
		if updateErr := w.executionService.CompleteFailure(
			ctx,
			execution,
			err,
		); updateErr != nil {
			return fmt.Errorf("execution failed and status update failed: %w", updateErr)
		}

		_ = w.executionService.Log(
			ctx,
			execution.ID,
			executions.LogError,
			err.Error(),
		)

		w.logger.Error().
			Err(err).
			Str("execution_id", execution.ID).
			Msg("execution failed")

		return nil
	}

	if err := w.executionService.CompleteSuccess(
		ctx,
		execution,
	); err != nil {
		return fmt.Errorf("failed to complete execution: %w", err)
	}

	_ = w.executionService.Log(
		ctx,
		execution.ID,
		executions.LogInfo,
		"execution completed successfully",
	)

	w.logger.Info().
		Str("execution_id", execution.ID).
		Str("task_id", task.ID).
		Msg("execution completed successfully")

	return nil
}

func (w *Worker) executeActions(
	ctx context.Context,
	task *tasks.Task,
	execution *executions.Execution,
) error {
	actionsList, err := w.actionRepository.FindByTaskID(ctx, task.ID)
	if err != nil {
		return fmt.Errorf("failed to load actions: %w", err)
	}

	for _, action := range actionsList {
		actionContext := actions.ExecutionContext{
			TaskID:   task.ID,
			ActionID: action.ID,
			UserID:   task.UserID,
		}

		executor, err := w.registry.Get(action.ActionType)
		if err != nil {
			actionErr := fmt.Errorf("executor not found for action type %s: %w", action.ActionType, err)

			_ = w.executionService.Log(
				ctx,
				execution.ID,
				executions.LogError,
				actionErr.Error(),
			)

			if !action.ContinueOnFailure {
				return actionErr
			}

			continue
		}

		actionCtx, cancel := context.WithTimeout(
			ctx,
			time.Duration(task.TimeoutSeconds)*time.Second,
		)

		result, err := executor.Execute(
			actionCtx,
			actionContext,
			action.Configuration,
		)

		cancel()

		if err != nil {
			actionErr := fmt.Errorf("action %s required: %w", action.ID, err)

			_ = w.executionService.Log(
				ctx,
				execution.ID,
				executions.LogError,
				actionErr.Error(),
			)

			if !action.ContinueOnFailure {
				return actionErr
			}

			continue
		}

		message := result.Message
		if strings.TrimSpace(message) == "" {
			message = "action executed successfully"
		}

		_ = w.executionService.Log(
			ctx,
			execution.ID,
			executions.LogInfo,
			message,
		)

		w.logger.Info().
			Str("execution_id", execution.ID).
			Str("action_id", action.ID).
			Str("action_type", action.ActionType).
			Msg("action executed")
	}

	return nil
}
