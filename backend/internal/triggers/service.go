package triggers

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Saif724/STAQ/backend/internal/tasks"
	"github.com/google/uuid"
	"github.com/robfig/cron"
)

var (
	ErrInvalidTriggerType = errors.New("invalid trigger type")
	ErrInvalidTimezone    = errors.New("invalid timezone")
	ErrCronRequired       = errors.New("cron expression is required")
	ErrCronInvalid        = errors.New("invalid cron expression")
)

type Service struct {
	repository  *Repository
	taskService *tasks.Service
}

func NewService(repository *Repository, taskService *tasks.Service) *Service {
	return &Service{
		repository:  repository,
		taskService: taskService,
	}
}

type CreateTriggerInput struct {
	TaskID         string
	TriggerType    string
	CronExpression *string
	Timezone       string
	StartAt        time.Time
}

type UpdateTriggerInput struct {
	ID             string
	TriggerType    string
	CronExpression *string
	Timezone       string
	StartAt        time.Time
	IsActive       bool
}

func (s *Service) Create(
	ctx context.Context,
	userID string,
	input CreateTriggerInput,
) (*Trigger, error) {
	taskID := strings.TrimSpace(input.TaskID)
	userID = strings.TrimSpace(userID)

	if taskID == "" {
		return nil, errors.New("task id is required")
	}

	if userID == "" {
		return nil, errors.New("user id is required")
	}

	if _, err := s.taskService.GetByID(ctx, taskID, userID); err != nil {
		return nil, err
	}

	triggerType := strings.ToUpper(strings.TrimSpace(input.TriggerType))

	if !isValidTriggerType(triggerType) {
		return nil, ErrInvalidTriggerType
	}

	timezone := strings.TrimSpace(input.Timezone)

	location, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, ErrInvalidTimezone
	}

	startAt := input.StartAt.In(location)

	cronExpression, err := validateSchedule(
		triggerType,
		input.CronExpression,
	)
	if err != nil {
		return nil, err
	}

	nextRunAt, err := calculateFirstRun(
		triggerType,
		cronExpression,
		startAt,
		location,
	)

	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	trigger := &Trigger{
		ID:             uuid.NewString(),
		TaskID:         taskID,
		TriggerType:    triggerType,
		CronExpression: cronExpression,
		TimeZone:       timezone,
		NextRunAt:      nextRunAt.UTC(),
		LastRunAt:      nil,
		IsActive:       true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.repository.Create(ctx, trigger); err != nil {
		return nil, err
	}

	return trigger, err
}

func (s *Service) GetByID(
	ctx context.Context,
	triggerID string,
	userID string,
) (*Trigger, error) {
	if strings.TrimSpace(triggerID) == "" {
		return nil, errors.New("trigger id is required")
	}

	userID = strings.TrimSpace(userID)

	if userID == "" {
		return nil, errors.New("user id is required")
	}

	trigger, err := s.repository.FindByID(ctx, triggerID)
	if err != nil {
		return nil, err
	}

	if _, err := s.taskService.GetByID(ctx, trigger.TaskID, userID); err != nil {
		return nil, err
	}

	return trigger, nil
}

func (s *Service) ListByTaskID(
	ctx context.Context,
	taskID string,
	userID string,
) ([]Trigger, error) {
	taskID = strings.TrimSpace(taskID)
	userID = strings.TrimSpace(userID)

	if taskID == "" {
		return nil, errors.New("task id is required")
	}

	if userID == "" {
		return nil, errors.New("user id is required")
	}

	if _, err := s.taskService.GetByID(ctx, taskID, userID); err != nil {
		return nil, err
	}
	return s.repository.FindByTaskID(ctx, taskID)
}

func (s *Service) Update(
	ctx context.Context,
	input UpdateTriggerInput,
	userID string,
) (*Trigger, error) {
	if strings.TrimSpace(input.ID) == "" {
		return nil, errors.New("trigger id is required")
	}

	userID = strings.TrimSpace(userID)

	if userID == "" {
		return nil, errors.New("user id is required")
	}

	existing, err := s.repository.FindByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	if _, err := s.taskService.GetByID(ctx, existing.TaskID, userID); err != nil {
		return nil, err
	}

	triggerType := strings.ToUpper(strings.TrimSpace(input.TriggerType))

	if !isValidTriggerType(triggerType) {
		return nil, ErrInvalidTriggerType
	}

	timezone := strings.TrimSpace(input.Timezone)

	location, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, ErrInvalidTimezone
	}

	startAt := input.StartAt.In(location)

	cronExpression, err := validateSchedule(
		triggerType,
		input.CronExpression,
	)
	if err != nil {
		return nil, err
	}

	nextRunAt, err := calculateFirstRun(
		triggerType,
		cronExpression,
		startAt,
		location,
	)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	trigger := &Trigger{
		ID:             existing.ID,
		TaskID:         existing.TaskID,
		TriggerType:    triggerType,
		CronExpression: cronExpression,
		TimeZone:       timezone,
		NextRunAt:      nextRunAt.UTC(),
		LastRunAt:      existing.LastRunAt,
		IsActive:       input.IsActive,
		CreatedAt:      existing.CreatedAt,
		UpdatedAt:      now,
	}

	if err := s.repository.Update(ctx, trigger); err != nil {
		return nil, err
	}

	return s.repository.FindByID(ctx, trigger.ID)
}

func (s *Service) Delete(
	ctx context.Context,
	triggerID string,
	userID string,
) error {
	triggerID = strings.TrimSpace(triggerID)

	if triggerID == "" {
		return errors.New("trigger id is required")
	}

	trigger, err := s.repository.FindByID(ctx, triggerID)
	if err != nil {
		return err
	}

	if _, err := s.taskService.GetByID(ctx, trigger.TaskID, userID); err != nil {
		return err
	}

	return s.repository.Delete(ctx, triggerID)
}

func (s *Service) CalculateNextRun(
	trigger *Trigger,
	from time.Time,
) (time.Time, error) {
	location, err := time.LoadLocation(trigger.TimeZone)
	if err != nil {
		return time.Time{}, ErrInvalidTimezone
	}

	localFrom := from.In(location)

	switch trigger.TriggerType {
	case TypeOnce:
		return time.Time{}, nil

	case TypeDaily:
		return localFrom.AddDate(0, 0, 1).UTC(), nil

	case TypeWeekly:
		return localFrom.AddDate(0, 0, 7).UTC(), nil

	case TypeMonthly:
		return localFrom.AddDate(0, 1, 0).UTC(), nil

	case TypeYearly:
		return localFrom.AddDate(1, 0, 0).UTC(), nil

	case TypeCron:
		if trigger.CronExpression == nil ||
			strings.TrimSpace(*trigger.CronExpression) == "" {
			return time.Time{}, ErrCronRequired
		}

		schedule, err := cron.ParseStandard(
			strings.TrimSpace(*trigger.CronExpression),
		)
		if err != nil {
			return time.Time{}, ErrCronInvalid
		}
		return schedule.Next(localFrom).UTC(), nil
	}

	return time.Time{}, ErrInvalidTriggerType
}

func isValidTriggerType(triggerType string) bool {
	switch triggerType {
	case TypeOnce,
		TypeDaily,
		TypeWeekly,
		TypeMonthly,
		TypeYearly,
		TypeCron:
		return true
	default:
		return false
	}
}

func validateSchedule(
	triggerType string,
	cronExpression *string,
) (*string, error) {
	switch triggerType {
	case TypeOnce,
		TypeDaily,
		TypeWeekly,
		TypeMonthly,
		TypeYearly:

		if cronExpression != nil &&
			strings.TrimSpace(*cronExpression) != "" {
			return nil, errors.New("cron expression is only allowed for CRON triggers")
		}

		return nil, nil

	case TypeCron:
		if cronExpression == nil {
			return nil, ErrCronRequired
		}

		expression := strings.TrimSpace(*cronExpression)

		if _, err := cron.ParseStandard(expression); err != nil {
			return nil, ErrCronInvalid
		}

		return &expression, nil
	}

	return nil, ErrInvalidTriggerType
}

func calculateFirstRun(
	triggerType string,
	cronExpression *string,
	startAt time.Time,
	location *time.Location,
) (time.Time, error) {
	switch triggerType {
	case TypeOnce:
		return startAt, nil

	case TypeDaily:
		return startAt, nil

	case TypeWeekly:
		return startAt, nil

	case TypeMonthly:
		return startAt, nil

	case TypeYearly:
		return startAt, nil

	case TypeCron:
		if cronExpression == nil {
			return time.Time{}, ErrCronRequired
		}

		schedule, err := cron.ParseStandard(
			strings.TrimSpace(*cronExpression),
		)
		if err != nil {
			return time.Time{}, fmt.Errorf("%w: %v", ErrCronInvalid, err)
		}

		return schedule.Next(
			startAt.Add(-time.Nanosecond),
		).In(location), nil
	}

	return time.Time{}, ErrInvalidTriggerType
}

func (s *Service) GetByIDForScheduler(
	ctx context.Context,
	triggerID string,
) (*Trigger, error) {
	if strings.TrimSpace(triggerID) == "" {
		return nil, errors.New("trigger id is required")
	}

	return s.repository.FindByID(ctx, triggerID)
}
