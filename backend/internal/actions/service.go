package actions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Saif724/STAQ/backend/internal/tasks"
	"github.com/google/uuid"
)

var (
	ErrInvalidActionType     = errors.New("invalid action type")
	ErrInvalidExecutionOrder = errors.New("execution order must be greater than zero")
	ErrInvalidConfiguration  = errors.New("invalid action configuration")
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

type CreateActionInput struct {
	TaskID            string
	ActionType        string
	ExecutionOrder    int
	Configuration     json.RawMessage
	ContinueOnFailure bool
}

type UpdateActionInput struct {
	ID                string
	ActionType        string
	ExecutionOrder    int
	Configuration     json.RawMessage
	ContinueOnFailure bool
}

func (s *Service) Create(
	ctx context.Context,
	userID string,
	input CreateActionInput,
) (*Action, error) {
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

	actionType := strings.ToUpper(
		strings.TrimSpace(input.ActionType),
	)

	if !isValidActionType(actionType) {
		return nil, ErrInvalidActionType
	}

	if input.ExecutionOrder < 1 {
		return nil, ErrInvalidExecutionOrder
	}

	if !json.Valid(input.Configuration) {
		return nil, ErrInvalidConfiguration
	}

	now := time.Now().UTC()

	action := &Action{
		ID:                uuid.NewString(),
		TaskID:            taskID,
		ActionType:        actionType,
		ExecutionOrder:    input.ExecutionOrder,
		Configuration:     input.Configuration,
		ContinueOnFailure: input.ContinueOnFailure,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if err := s.repository.Create(ctx, action); err != nil {
		return nil, fmt.Errorf("failed to create action: %w", err)
	}

	return action, nil
}

func (s *Service) GetByID(
	ctx context.Context,
	actionID string,
	userID string,
) (*Action, error) {
	actionID = strings.TrimSpace(actionID)
	userID = strings.TrimSpace(userID)

	if actionID == "" {
		return nil, errors.New("action id is required")
	}

	if userID == "" {
		return nil, errors.New("user id is required")
	}

	action, err := s.repository.FindByID(ctx, actionID)
	if err != nil {
		return nil, err
	}

	if _, err := s.taskService.GetByID(ctx, action.TaskID, userID); err != nil {
		return nil, err
	}

	return action, nil
}

func (s *Service) ListByTaskID(
	ctx context.Context,
	taskID string,
	userID string,
) ([]Action, error) {
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
	userID string,
	input UpdateActionInput,
) (*Action, error) {
	input.ID = strings.TrimSpace(input.ID)
	userID = strings.TrimSpace(userID)

	if input.ID == "" {
		return nil, errors.New("action id is required")
	}

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

	actionType := strings.ToUpper(strings.TrimSpace(input.ActionType))

	if !isValidActionType(actionType) {
		return nil, ErrInvalidActionType
	}

	if input.ExecutionOrder < 1 {
		return nil, ErrInvalidExecutionOrder
	}

	if !json.Valid(input.Configuration) {
		return nil, ErrInvalidConfiguration
	}

	action := &Action{
		ID:                existing.ID,
		TaskID:            existing.TaskID,
		ActionType:        actionType,
		ExecutionOrder:    input.ExecutionOrder,
		Configuration:     input.Configuration,
		ContinueOnFailure: input.ContinueOnFailure,
		CreatedAt:         existing.CreatedAt,
		UpdatedAt:         time.Now().UTC(),
	}

	if err := s.repository.Update(ctx, action); err != nil {
		return nil, err
	}

	return s.repository.FindByID(ctx, action.ID)
}

func (s *Service) Delete(
	ctx context.Context,
	actionID string,
	userID string,
) error {
	actionID = strings.TrimSpace(actionID)
	userID = strings.TrimSpace(userID)

	if actionID == "" {
		return errors.New("action id is required")
	}

	if userID == "" {
		return errors.New("user id is required")
	}

	action, err := s.repository.FindByID(ctx, actionID)
	if err != nil {
		return err
	}

	if _, err := s.taskService.GetByID(ctx, action.TaskID, userID); err != nil {
		return err
	}

	return s.repository.Delete(ctx, actionID)
}

func isValidActionType(actionType string) bool {
	switch actionType {
	case TypeReminder,
		TypeEmail,
		TypeHTTP,
		TypeShell:
		return true
	default:
		return false
	}
}
