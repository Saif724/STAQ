package executions

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

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
	execution.RetryCount++

	return s.repository.Update(
		ctx,
		execution,
	)
}

func (s *Service) CompleteSuccess(
	ctx context.Context,
	execution *Execution,
) error {
	now := time.Now().UTC()

	execution.Status = StatusSuccess
	execution.CompletedAt = &now
	execution.DurationMs = durationMs(
		execution.StartedAt,
		now,
	)
	execution.ErrorMessage = nil
	return s.repository.Update(ctx, execution)
}

func (s *Service) CompleteFailure(
	ctx context.Context,
	execution *Execution,
	err error,
) error {
	now := time.Now().UTC()

	execution.Status = StatusFailed
	execution.CompletedAt = &now
	execution.DurationMs = durationMs(
		execution.StartedAt,
		now,
	)

	if err != nil {
		message := err.Error()
		execution.ErrorMessage = &message
	}

	return s.repository.Update(ctx, execution)

}

func (s *Service) CompleteTimeout(
	ctx context.Context,
	execution *Execution,
	err error,
) error {
	now := time.Now().UTC()

	execution.Status = StatusTimedOut
	execution.CompletedAt = &now
	execution.DurationMs = durationMs(execution.StartedAt, now)

	if err != nil {
		message := err.Error()
		execution.ErrorMessage = &message
	}

	return s.repository.Update(ctx, execution)
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
