package workers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/Saif724/STAQ/backend/internal/actions"
	emailAction "github.com/Saif724/STAQ/backend/internal/actions/email"
	httpaction "github.com/Saif724/STAQ/backend/internal/actions/http"
	"github.com/Saif724/STAQ/backend/internal/actions/reminder"
	"github.com/Saif724/STAQ/backend/internal/actions/shell"
	"github.com/Saif724/STAQ/backend/internal/broker"
	"github.com/Saif724/STAQ/backend/internal/executions"
	"github.com/Saif724/STAQ/backend/internal/tasks"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

const (
	initialRetryDelay  = 1 * time.Second
	maxRetryDelay      = 30 * time.Second
	defaultConcurrency = 4
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

func NewRegistry(
	emailSender emailAction.Sender,
) *actions.Registry {
	registry := actions.NewRegistry()

	registry.Register(
		actions.TypeReminder,
		reminder.NewExecutor(),
	)

	registry.Register(
		actions.TypeEmail,
		emailAction.NewExecutor(emailSender),
	)

	registry.Register(
		actions.TypeHTTP,
		httpaction.NewExecutor(
			&http.Client{
				Timeout: 30 * time.Second,
			},
		),
	)

	echoPath, err := exec.LookPath("echo")
	if err != nil {
		panic("echo command not found")
	}

	registry.Register(
		actions.TypeShell,
		shell.NewExecutor(
			[]string{echoPath},
		),
	)

	return registry
}

func (w *Worker) Run(ctx context.Context) error {
	w.logger.Info().
		Int("concurrency", defaultConcurrency).
		Msg("worker pool started")

	var wg sync.WaitGroup

	for i := 0; i < defaultConcurrency; i++ {
		wg.Add(1)

		go func(workerID int) {
			defer wg.Done()

			w.logger.Info().
				Int("worker_id", workerID).
				Msg("worker started")

			for {
				if err := w.processOne(ctx); err != nil {
					if errors.Is(err, context.Canceled) ||
						errors.Is(err, context.DeadlineExceeded) {
						w.logger.Info().
							Int("worker_id", workerID).
							Msg("worker stopped")

						return
					}

					if errors.Is(err, redis.Nil) {
						continue
					}

					w.logger.Error().
						Int("worker_id", workerID).
						Err(err).
						Msg("failed to process task job")
				}
			}
		}(i + 1)
	}

	wg.Wait()

	w.logger.Info().Msg("worker pool stopped")

	return ctx.Err()

}

func (w *Worker) processOne(ctx context.Context) error {
	result, err := w.broker.Client().BRPop(
		ctx,
		2*time.Second,
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
		job.ScheduledAt,
	)

	if errors.Is(err, executions.ErrExecutionAlreadyExist) {
		w.logger.Info().
			Str("task_id", task.ID).
			Str("trigger_id", job.TriggerID).
			Time("scheduled_at", job.ScheduledAt).
			Msg("duplicate scheduled job ignored")

		return nil
	}
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
		var completionErr error

		if errors.Is(err, context.DeadlineExceeded) {
			completionErr = w.executionService.CompleteTimeout(
				ctx,
				execution,
				err,
			)
		} else {
			completionErr = w.executionService.CompleteFailure(
				ctx,
				execution,
				err,
			)
		}

		if completionErr != nil {
			return fmt.Errorf("execution failed and status update failed: %w", completionErr)
		}

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

	if err := w.executionService.Log(
		ctx,
		execution.ID,
		executions.LogInfo,
		"execution completed successfully",
	); err != nil {
		w.logger.Warn().
			Err(err).
			Str("execution_id", execution.ID).
			Msg("failed to persist completion log")
	}

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

			w.logActionError(ctx, execution.ID, actionErr)

			if !action.ContinueOnFailure {
				return actionErr
			}

			continue
		}

		var result *actions.ExecutionResult

		actionRetryCount := 0

		for {
			actionCtx, cancel := context.WithTimeout(
				ctx,
				time.Duration(task.TimeoutSeconds)*time.Second,
			)

			result, err = executor.Execute(
				actionCtx,
				actionContext,
				action.Configuration,
			)

			cancel()

			if err == nil {
				break
			}

			actionErr := fmt.Errorf("action %s failed: %w", action.ID, err)

			result = nil

			if ctx.Err() != nil {
				return ctx.Err()
			}

			if errors.Is(err, context.DeadlineExceeded) ||
				errors.Is(err, context.Canceled) {
				w.logActionError(ctx, execution.ID, actionErr)

				if !action.ContinueOnFailure {
					return actionErr
				}

				break
			}

			if !actions.IsRetryable(err) {
				w.logActionError(ctx, execution.ID, actionErr)

				if !action.ContinueOnFailure {
					return actionErr
				}

				break
			}

			if actionRetryCount >= task.MaxRetries {
				w.logActionError(ctx, execution.ID, actionErr)

				if !action.ContinueOnFailure {
					return actionErr
				}

				break
			}

			actionRetryCount++

			if err := w.executionService.IncrementRetryCount(
				ctx,
				execution,
			); err != nil {
				return fmt.Errorf("failed to update retry count: %w", err)
			}

			delay := retryDelay(actionRetryCount)

			retryMessage := fmt.Sprintf("action %s failed; retrying in %s (%d/%d)", action.ID, delay, actionRetryCount, task.MaxRetries)

			if err := w.executionService.Log(
				ctx,
				execution.ID,
				executions.LogWarning,
				retryMessage,
			); err != nil {
				w.logger.Warn().
					Err(err).
					Str("execution_id", execution.ID).
					Msg("failed to persist retry log")
			}

			timer := time.NewTimer(delay)

			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()

			case <-timer.C:
			}
		}

		if result == nil {
			continue
		}

		message := result.Message
		if strings.TrimSpace(message) == "" {
			message = "action executed successfully"
		}

		if result.Data != nil {
			resultJSON, err := json.Marshal(map[string]any{
				"message": message,
				"data":    result.Data,
			})

			if err != nil {
				w.logger.Warn().
					Err(err).
					Str("execution_id", execution.ID).
					Str("action_id", action.ID).
					Msg("failed to serialize action result")
			} else {
				message = string(resultJSON)
			}
		}

		if err := w.executionService.Log(
			ctx,
			execution.ID,
			executions.LogInfo,
			message,
		); err != nil {
			w.logger.Warn().
				Err(err).
				Str("execution_id", execution.ID).
				Str("action_id", action.ID).
				Msg("failed to persist action result")
		}

		w.logger.Info().
			Str("execution_id", execution.ID).
			Str("action_id", action.ID).
			Str("action_type", action.ActionType).
			Interface("result", result.Data).
			Msg("action executed")
	}

	return nil
}

func retryDelay(retryCount int) time.Duration {
	if retryCount <= 0 {
		return initialRetryDelay
	}

	delay := initialRetryDelay

	for i := 1; i < retryCount; i++ {
		delay *= 2

		if delay >= maxRetryDelay {
			return maxRetryDelay
		}
	}

	return delay
}

func (w *Worker) logActionError(
	ctx context.Context,
	executionID string,
	err error,
) {
	if logErr := w.executionService.Log(
		ctx,
		executionID,
		executions.LogError,
		err.Error(),
	); logErr != nil {
		w.logger.Warn().
			Err(logErr).
			Str("execution_id", executionID).
			Msg("failed to persist action error")
	}
}
