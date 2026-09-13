package tasks

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Saif724/STAQ/backend/internal/queues"
	"github.com/google/uuid"
)

type Service struct {
	repository   *Repository
	queueService *queues.Service
}

func NewService(repository *Repository, queueService *queues.Service) *Service {
	return &Service{
		repository:   repository,
		queueService: queueService,
	}
}

func (s *Service) validateQueue(
	ctx context.Context,
	queueID string,
) error {
	if strings.TrimSpace(queueID) == "" {
		return errors.New("queue id is required")
	}

	queue, err := s.queueService.GetByID(ctx, queueID)
	if err != nil {
		if errors.Is(err, queues.ErrQueueNotFound) {
			return errors.New("queue not found")
		}

		return fmt.Errorf("failed to validate queue: %w", err)
	}

	if !queue.IsActive {
		return errors.New("queue is inactive")
	}

	return nil
}

type CreateTaskInput struct {
	UserID         string
	QueueID        string
	Name           string
	Description    *string
	TimeoutSeconds int
	MaxRetries     int
}

type UpdateTaskInput struct {
	TaskID         string
	UserID         string
	QueueID        string
	Name           string
	Description    *string
	Status         string
	TimeoutSeconds int
	MaxRetries     int
}

func (s *Service) Create(
	ctx context.Context,
	input CreateTaskInput,
) (*Task, error) {
	if strings.TrimSpace(input.UserID) == "" {
		return nil, errors.New("user id is required")
	}

	if err := s.validateQueue(ctx, input.QueueID); err != nil {
		return nil, err
	}

	name := strings.TrimSpace(input.Name)

	if name == "" {
		return nil, errors.New("task name is required")
	}

	if len(name) > 150 {
		return nil, errors.New("task name must not exceed 150 characters")
	}

	if input.TimeoutSeconds <= 0 {
		input.TimeoutSeconds = 300
	}

	if input.MaxRetries < 0 {
		return nil, errors.New("max retries cannot be negative")
	}

	now := time.Now().UTC()

	task := &Task{
		ID:             uuid.NewString(),
		UserID:         input.UserID,
		QueueID:        input.QueueID,
		Name:           name,
		Description:    input.Description,
		Status:         StatusActive,
		TimeoutSeconds: input.TimeoutSeconds,
		MaxRetries:     input.MaxRetries,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.repository.Create(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	return task, nil
}

func (s *Service) GetByID(
	ctx context.Context,
	taskID string,
	userID string,
) (*Task, error) {
	if strings.TrimSpace(taskID) == "" {
		return nil, errors.New("task id is required")
	}

	if strings.TrimSpace(userID) == "" {
		return nil, errors.New("user id is required")
	}

	task, err := s.repository.FindByIDAndUser(ctx, taskID, userID)

	if err != nil {
		return nil, err
	}

	return task, nil
}

func (s *Service) List(
	ctx context.Context,
	userID string,
) ([]Task, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, errors.New("user id is required")
	}

	tasks, err := s.repository.FindByUser(ctx, userID)

	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	return tasks, nil
}

func (s *Service) Update(
	ctx context.Context,
	input UpdateTaskInput,
) (*Task, error) {
	if strings.TrimSpace(input.TaskID) == "" {
		return nil, errors.New("task id is required")
	}

	if strings.TrimSpace(input.UserID) == "" {
		return nil, errors.New("user id is required")
	}

	name := strings.TrimSpace(input.Name)

	if name == "" {
		return nil, errors.New("task name is required")
	}

	if len(name) > 150 {
		return nil, errors.New("task name must not exceed 150 characters")
	}

	if !isValidStatus(input.Status) {
		return nil, errors.New("invalid task status")
	}

	if input.TimeoutSeconds <= 0 {
		return nil, errors.New("timeout seconds must be greater than zero")
	}

	if input.MaxRetries < 0 {
		return nil, errors.New("max retries cannot be negative")
	}

	if err := s.validateQueue(ctx, input.QueueID); err != nil {
		return nil, err
	}

	task := &Task{
		ID:             input.TaskID,
		UserID:         input.UserID,
		QueueID:        input.QueueID,
		Name:           name,
		Description:    input.Description,
		Status:         input.Status,
		TimeoutSeconds: input.TimeoutSeconds,
		MaxRetries:     input.MaxRetries,
		UpdatedAt:      time.Now().UTC(),
	}

	if err := s.repository.Update(ctx, task); err != nil {
		return nil, err
	}

	return s.GetByID(ctx, input.TaskID, input.UserID)
}

func (s *Service) Archive(
	ctx context.Context,
	taskID string,
	userID string,
) error {
	if strings.TrimSpace(taskID) == "" {
		return errors.New("task id is required")
	}

	if strings.TrimSpace(userID) == "" {
		return errors.New("user id is required")
	}

	return s.repository.Archive(ctx, taskID, userID)
}

func isValidStatus(status string) bool {
	switch status {
	case StatusActive, StatusPaused, StatusArchived:
		return true
	default:
		return false
	}
}
