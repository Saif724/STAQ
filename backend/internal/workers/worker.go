package workers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
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
	heartbeatInterval  = 10 * time.Second
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
	platformEmailSender emailAction.PlatformSender,
	gmailIntegrationService emailAction.UserSender,
) *actions.Registry {
	registry := actions.NewRegistry()

	registry.Register(
		actions.TypeReminder,
		reminder.NewExecutor(),
	)

	registry.Register(
		actions.TypeEmail,
		emailAction.NewExecutor(platformEmailSender, gmailIntegrationService),
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
	concurrency := workerConcurrency()

	w.logger.Info().
		Int("concurrency", concurrency).
		Msg("worker pool started")

	if err := w.broker.EnsureTaskConsumerGroup(ctx); err != nil {
		return fmt.Errorf("failed to initialize task consumer group: %w", err)
	}

	recoveryConsumer := fmt.Sprintf("recovery-%d", time.Now().UnixNano())

	go w.runPendingJobRecovery(
		ctx,
		recoveryConsumer,
	)

	var wg sync.WaitGroup

	for i := 0; i < concurrency; i++ {
		wg.Add(1)

		go func(workerID int) {
			defer wg.Done()

			consumer := fmt.Sprintf(
				"worker-%d",
				workerID,
			)

			w.logger.Info().
				Int("worker_id", workerID).
				Str("consumer", consumer).
				Msg("worker started")

			consecutiveErrors := 0

			for {
				if err := w.processOne(ctx, consumer); err != nil {
					if errors.Is(err, context.Canceled) ||
						errors.Is(err, context.DeadlineExceeded) {
						w.logger.Info().
							Int("worker_id", workerID).
							Msg("worker stopped")

						return
					}

					if errors.Is(err, redis.Nil) {
						consecutiveErrors = 0
						continue
					}

					consecutiveErrors++

					w.logger.Error().
						Int("worker_id", workerID).
						Int("consecutive_errors", consecutiveErrors).
						Err(err).
						Msg("failed to process task job")

					if !sleepContext(ctx, errorBackoff(consecutiveErrors)) {
						return
					}
				}
			}
		}(i + 1)
	}

	wg.Wait()

	w.logger.Info().Msg("worker pool stopped")

	return ctx.Err()

}

func (w *Worker) runPendingJobRecovery(
	ctx context.Context,
	consumer string,
) {
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			w.recoverPendingJobs(
				ctx,
				consumer,
			)
		}
	}
}

func (w *Worker) processOne(ctx context.Context, consumer string) error {
	messageID, job, err := w.broker.ReadTaskJob(
		ctx,
		consumer,
	)
	if err != nil {
		return err
	}

	if err := w.executeJob(ctx, job, messageID, consumer); err != nil {
		return err
	}

	if err := w.broker.AcknowledgeTaskJob(
		ctx,
		messageID,
	); err != nil {
		return err
	}

	return nil
}

func (w *Worker) executeJob(
	ctx context.Context,
	job broker.TaskJob,
	messageID string,
	consumer string,
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
		execution, findErr := w.executionService.FindExisting(
			ctx,
			job.TriggerID,
			job.ScheduledAt,
		)

		if findErr != nil {
			return fmt.Errorf("failed to find existing execution: %w", findErr)
		}
		if execution == nil {
			return fmt.Errorf("execution already exists but could not be found")
		}

		if execution.Status != executions.StatusRunning {
			w.logger.Info().
				Str("execution_id", execution.ID).
				Str("task_id", task.ID).
				Str("trigger_id", job.TriggerID).
				Time("scheduled_at", job.ScheduledAt).
				Str("status", execution.Status).
				Msg("duplicate scheduled job ignored")

			return nil
		}

		reclaimed, reclaimErr := w.executionService.Reclaim(
			ctx,
			execution,
		)
		if reclaimErr != nil {
			return fmt.Errorf("failed to reclaim existing execution: %w", reclaimErr)
		}

		if !reclaimed {
			w.logger.Info().
				Str("execution_id", execution.ID).
				Str("task_id", task.ID).
				Str("trigger_id", job.TriggerID).
				Msg("existing execution is still active; skipping")

			return nil
		}

		w.logger.Warn().
			Str("execution_id", execution.ID).
			Str("task_id", task.ID).
			Str("trigger_id", job.TriggerID).
			Msg("reclaimed stale execution")
	} else if err != nil {
		return fmt.Errorf("failed to start execution: %w", err)
	}

	heartbeatCtx, heartbeatCancel := context.WithCancel(ctx)
	defer heartbeatCancel()

	go w.runExecutionHeartbeat(
		heartbeatCtx,
		execution.ID,
	)

	go w.runTaskJobRenewal(
		heartbeatCtx,
		messageID,
		consumer,
	)

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

func (w *Worker) runTaskJobRenewal(
	ctx context.Context,
	messageID string,
	consumer string,
) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			if err := w.broker.RenewTaskJob(
				ctx,
				messageID,
				consumer,
			); err != nil {
				if errors.Is(err, context.Canceled) ||
					errors.Is(err, context.DeadlineExceeded) {
					return
				}

				w.logger.Error().
					Err(err).
					Str("message_id", messageID).
					Str("consumer", consumer).
					Msg("failed to renew task job")
			}
		}
	}
}

func (w *Worker) runExecutionHeartbeat(
	ctx context.Context,
	executionID string,
) {
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			if err := w.executionService.Heartbeat(
				ctx,
				executionID,
			); err != nil {
				if errors.Is(err, context.Canceled) ||
					errors.Is(err, context.DeadlineExceeded) {
					return
				}

				w.logger.Error().
					Err(err).
					Str("execution_id", executionID).
					Msg("failed to update execution heartbeat")
			}
		}
	}
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

func (w *Worker) recoverPendingJobs(
	ctx context.Context,
	consumer string,
) {
	messages, err := w.broker.ClaimStaleTaskJobs(
		ctx,
		consumer,
	)
	if err != nil {
		w.logger.Error().
			Err(err).
			Msg("failed to claim stale task jobs")

		return
	}

	for _, message := range messages {
		rawJob, ok := message.Values["job"].(string)
		if !ok {
			w.logger.Error().
				Str("message_id", message.ID).
				Msg("claimed task job has invalid payload")

			continue
		}

		var job broker.TaskJob
		if err := json.Unmarshal(
			[]byte(rawJob),
			&job,
		); err != nil {
			w.logger.Error().
				Err(err).
				Str("message_id", message.ID).
				Msg("failed to decode claimed task job")

			continue
		}
		w.logger.Warn().
			Str("message_id", message.ID).
			Str("consumer", consumer).
			Str("task_id", job.TaskID).
			Str("trigger_id", job.TriggerID).
			Msg("claimed stale task job")

		if err := w.executeJob(ctx, job, message.ID, consumer); err != nil {
			w.logger.Error().
				Err(err).
				Str("message_id", message.ID).
				Msg("failed to execute claimed task job")

			continue
		}

		if err := w.broker.AcknowledgeTaskJob(
			ctx,
			message.ID,
		); err != nil {
			w.logger.Error().
				Err(err).
				Str("message_id", message.ID).
				Msg("failed to acknowledge recovered task job")
		}
	}
}

func errorBackoff(consecutive int) time.Duration {
	delay := 5 * time.Second
	for i := 1; i < consecutive && delay < 5*time.Minute; i++ {
		delay *= 2
	}
	if delay > 5*time.Minute {
		delay = 5 * time.Minute
	}
	return delay
}

func sleepContext(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func workerConcurrency() int {
	if v, err := strconv.Atoi(os.Getenv("WORKER_CONCURRENCY")); err == nil && v > 0 {
		return v
	}
	return defaultConcurrency
}
