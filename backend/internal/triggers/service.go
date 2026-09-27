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
	ErrInvalidStartAt     = errors.New("start_at is required")
	ErrStartAtInPast      = errors.New("start_at is in the past")
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

	if input.StartAt.IsZero() {
		return nil, ErrInvalidStartAt
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
		StartAt:        startAt,
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

	if input.StartAt.IsZero() {
		return nil, ErrInvalidStartAt
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
		StartAt:        startAt,
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

	userID = strings.TrimSpace(userID)

	if userID == "" {
		return errors.New("user id is required")
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

	scheduledAt := from.In(location)
	startAt := trigger.StartAt.In(location)

	switch trigger.TriggerType {
	case TypeOnce:
		return time.Time{}, nil

	case TypeDaily:
		return nextDailyFromAnchor(startAt, scheduledAt).UTC(), nil

	case TypeWeekly:
		return nextWeeklyFromAnchor(startAt, scheduledAt).UTC(), nil

	case TypeMonthly:
		return nextMonthlyFromAnchor(startAt, scheduledAt).UTC(), nil

	case TypeYearly:
		return nextYearlyFromAnchor(startAt, scheduledAt).UTC(), nil
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

		return schedule.Next(scheduledAt).UTC(), nil
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

		if expression == "" {
			return nil, ErrCronRequired
		}

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
	now := time.Now().In(location)

	switch triggerType {
	case TypeOnce:
		if !startAt.After(now) {
			return time.Time{}, ErrStartAtInPast
		}

		return startAt, nil

	case TypeDaily:
		return nextDaily(startAt, now), nil

	case TypeWeekly:
		return nextWeekly(startAt, now), nil

	case TypeMonthly:
		return nextMonthly(startAt, now), nil

	case TypeYearly:
		return nextYearly(startAt, now), nil

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

		base := startAt
		if !base.After(now) {
			base = now
		}

		return schedule.Next(base).In(location), nil
	}

	return time.Time{}, ErrInvalidTriggerType
}

func nextDailyFromAnchor(anchor, after time.Time) time.Time {
	if after.Before(anchor) {
		return anchor
	}

	days := int(after.Sub(anchor) / (24 * time.Hour))
	candidate := anchor.AddDate(0, 0, days)

	if !candidate.After(after) {
		candidate = candidate.AddDate(0, 0, 1)
	}

	return candidate
}

func nextWeeklyFromAnchor(anchor, after time.Time) time.Time {
	if after.Before(anchor) {
		return anchor
	}

	weeks := int(after.Sub(anchor) / (7 * 24 * time.Hour))
	candidate := anchor.AddDate(0, 0, weeks*7)

	if !candidate.After(after) {
		candidate = candidate.AddDate(0, 0, 7)
	}

	return candidate
}

func nextMonthlyFromAnchor(anchor, after time.Time) time.Time {
	if after.Before(anchor) {
		return anchor
	}

	months := (after.Year()-anchor.Year())*12 + int(after.Month()-anchor.Month())

	for {
		candidate := addMonthsPreservingAnchor(anchor, months)

		if candidate.After(after) {
			return candidate
		}

		months++
	}
}

func addMonthsPreservingAnchor(anchor time.Time, months int) time.Time {
	totalMonths := int(anchor.Month()) - 1 + months

	year := anchor.Year() + totalMonths/12
	month := time.Month(totalMonths%12 + 1)

	day := anchor.Day()
	lastDay := daysInMonth(year, month)

	if day > lastDay {
		day = lastDay
	}

	return time.Date(
		year,
		month,
		day,
		anchor.Hour(),
		anchor.Minute(),
		anchor.Second(),
		anchor.Nanosecond(),
		anchor.Location(),
	)
}

func nextYearlyFromAnchor(anchor, after time.Time) time.Time {
	if after.Before(anchor) {
		return anchor
	}

	years := after.Year() - anchor.Year()
	for {
		candidate := addYearsPreservingAnchor(anchor, years)

		if candidate.After(after) {
			return candidate
		}

		years++
	}
}

func addYearsPreservingAnchor(anchor time.Time, years int) time.Time {
	year := anchor.Year() + years

	day := anchor.Day()
	lastDay := daysInMonth(year, anchor.Month())

	if day > lastDay {
		day = lastDay
	}

	return time.Date(
		year,
		anchor.Month(),
		day,
		anchor.Hour(),
		anchor.Minute(),
		anchor.Second(),
		anchor.Nanosecond(),
		anchor.Location(),
	)
}
func nextDaily(startAt, now time.Time) time.Time {
	return nextDailyFromAnchor(startAt, now)
}

func nextWeekly(startAt, now time.Time) time.Time {
	return nextWeeklyFromAnchor(startAt, now)
}

func nextMonthly(startAt, now time.Time) time.Time {
	return nextMonthlyFromAnchor(startAt, now)
}

func nextYearly(startAt, now time.Time) time.Time {
	return nextYearlyFromAnchor(startAt, now)
}

func daysInMonth(year int, month time.Month) int {
	return time.Date(
		year,
		month+1,
		0,
		0,
		0,
		0,
		0,
		time.UTC,
	).Day()
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
