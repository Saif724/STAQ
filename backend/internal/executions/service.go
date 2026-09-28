package executions

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const staleExecutionTimeout = 2 * time.Minute

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) Start(
	ctx context.Context,
	taskID string,
	triggerID string,
	scheduledAt time.Time,
) (*Execution, error) {
	if strings.TrimSpace(taskID) == "" {
		return nil, errors.New("task id is required")
	}

	if strings.TrimSpace(triggerID) == "" {
		return nil, errors.New("trigger id is required")
	}

	if scheduledAt.IsZero() {
		return nil, errors.New("scheduled time is required")
	}

	now := time.Now().UTC()

	execution := &Execution{
		ID:          uuid.NewString(),
		TaskID:      taskID,
		TriggerID:   triggerID,
		ScheduledAt: scheduledAt.UTC(),
		Status:      StatusRunning,
		StartedAt:   now,
		HeartbeatAt: &now,
		RetryCount:  0,
		CreatedAt:   now,
	}

	err := s.repository.Create(ctx, execution)

	if err == nil {
		return execution, nil
	}

	if errors.Is(err, ErrExecutionAlreadyExist) {
		existing, findErr := s.repository.FindByTriggerAndScheduledAt(
			ctx,
			triggerID,
			scheduledAt.UTC(),
		)
		if findErr != nil {
			return nil, findErr
		}

		return existing, ErrExecutionAlreadyExist
	}

	return nil, err
}

func (s *Service) IncrementRetryCount(
	ctx context.Context,
	execution *Execution,
) error {
	if execution == nil {
		return errors.New("execution is required")
	}

	if strings.TrimSpace(execution.ID) == "" {
		return errors.New("execution id is required")
	}

	if err := s.repository.IncrementRetryCount(
		ctx,
		execution.ID,
	); err != nil {
		return err
	}

	execution.RetryCount++

	return nil
}

func (s *Service) CompleteSuccess(
	ctx context.Context,
	execution *Execution,
) error {
	now := time.Now().UTC()

	execution.Status = StatusSuccess
	execution.CompletedAt = &now
	execution.HeartbeatAt = nil
	execution.DurationMs = durationMs(
		execution.StartedAt,
		now,
	)
	execution.ErrorMessage = nil

	return s.repository.Complete(ctx, execution)
}

func (s *Service) CompleteFailure(
	ctx context.Context,
	execution *Execution,
	err error,
) error {
	now := time.Now().UTC()

	execution.Status = StatusFailed
	execution.CompletedAt = &now
	execution.HeartbeatAt = nil
	execution.DurationMs = durationMs(
		execution.StartedAt,
		now,
	)

	if err != nil {
		message := err.Error()
		execution.ErrorMessage = &message
	}

	return s.repository.Complete(ctx, execution)

}

func (s *Service) CompleteTimeout(
	ctx context.Context,
	execution *Execution,
	err error,
) error {
	now := time.Now().UTC()

	execution.Status = StatusTimedOut
	execution.CompletedAt = &now
	execution.HeartbeatAt = nil
	execution.DurationMs = durationMs(execution.StartedAt, now)

	if err != nil {
		message := err.Error()
		execution.ErrorMessage = &message
	}

	return s.repository.Complete(ctx, execution)
}

func (s *Service) Log(
	ctx context.Context,
	executionID string,
	level string,
	message string,
) error {
	if strings.TrimSpace(executionID) == "" {
		return errors.New("execution id is required")
	}

	if strings.TrimSpace(message) == "" {
		return errors.New("log message is required")
	}

	switch level {
	case LogInfo, LogWarning, LogError:
	default:
		return errors.New("invalid execution log level")
	}

	log := &ExecutionLog{
		ID:          uuid.NewString(),
		ExecutionID: executionID,
		LogLevel:    level,
		Message:     message,
		CreatedAt:   time.Now().UTC(),
	}

	return s.repository.CreateLog(ctx, log)
}

func (s *Service) GetByID(
	ctx context.Context,
	executionID string,
) (*Execution, error) {
	return s.repository.FindByID(ctx, executionID)
}

func (s *Service) GetByIDAndUser(
	ctx context.Context,
	executionID string,
	userID string,
) (*Execution, error) {
	if strings.TrimSpace(executionID) == "" {
		return nil, errors.New("execution id is required")
	}

	if strings.TrimSpace(userID) == "" {
		return nil, errors.New("user id is required")
	}

	return s.repository.FindByIDAndUser(ctx, executionID, userID)
}

func (s *Service) ListLogs(
	ctx context.Context,
	executionID string,
) ([]ExecutionLog, error) {
	return s.repository.FindLogs(ctx, executionID)
}

func (s *Service) ListLogsByUser(
	ctx context.Context,
	executionID string,
	userID string,
) ([]ExecutionLog, error) {
	if strings.TrimSpace(executionID) == "" {
		return nil, errors.New("execution id is required")
	}

	if strings.TrimSpace(userID) == "" {
		return nil, errors.New("user id is required")
	}

	return s.repository.FindLogsByUser(ctx, executionID, userID)
}

func durationMs(start, end time.Time) *int64 {
	duration := end.Sub(start).Milliseconds()
	return &duration
}

func (s *Service) Heartbeat(
	ctx context.Context,
	executionID string,
) error {
	executionID = strings.TrimSpace(executionID)

	if executionID == "" {
		return errors.New("execution id is required")
	}

	now := time.Now().UTC()

	return s.repository.Heartbeat(
		ctx,
		executionID,
		now,
	)
}

func (s *Service) RecoverStale(
	ctx context.Context,
) ([]string, error) {
	before := time.Now().UTC().Add(-staleExecutionTimeout)

	return s.repository.RecoveryStaleRunning(
		ctx,
		before,
	)
}

func (s *Service) FindExisting(
	ctx context.Context,
	triggerID string,
	scheduledAt time.Time,
) (*Execution, error) {
	return s.repository.FindByTriggerAndScheduledAt(
		ctx,
		triggerID,
		scheduledAt,
	)
}

func (s *Service) Reclaim(
	ctx context.Context,
	execution *Execution,
) (bool, error) {
	if execution == nil {
		return false, fmt.Errorf("execution is required")
	}

	if execution.Status != StatusRunning {
		return false, nil
	}

	if execution.HeartbeatAt == nil {
		return false, nil
	}

	now := time.Now()

	if now.Sub(*execution.HeartbeatAt) < staleExecutionTimeout {
		return false, nil
	}

	reclaimed, err := s.repository.ReclaimExecution(
		ctx,
		execution.ID,
		now,
	)
	if err != nil {
		return false, err
	}

	return reclaimed, nil
}
