package queues

import (
	"context"
	"errors"
	"fmt"
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

type CreateQueueInput struct {
	Name        string
	Description *string
}

type UpdateQueueInput struct {
	ID          string
	Name        string
	Description *string
	IsActive    bool
}

func (s *Service) Create(
	ctx context.Context,
	input CreateQueueInput,
) (*Queue, error) {
	name := strings.TrimSpace(input.Name)

	if name == "" {
		return nil, errors.New("queue name is required")
	}

	if len(name) > 100 {
		return nil, errors.New("queue name must not exceed 100 characters")
	}

	now := time.Now().UTC()

	queue := &Queue{
		ID:          uuid.NewString(),
		Name:        name,
		Description: input.Description,
		IsActive:    true,
		CreatedAt:   now,
	}

	if err := s.repository.Create(ctx, queue); err != nil {
		return nil, fmt.Errorf("failed to create queue: %w", err)
	}

	return queue, nil
}

func (s *Service) GetByID(
	ctx context.Context,
	queueID string,
) (*Queue, error) {
	if strings.TrimSpace(queueID) == "" {
		return nil, errors.New("queue id is required")
	}

	return s.repository.FindByID(ctx, queueID)
}

func (s *Service) List(
	ctx context.Context,
) ([]Queue, error) {
	return s.repository.FindAll(ctx)
}

func (s *Service) Update(
	ctx context.Context,
	input UpdateQueueInput,
) (*Queue, error) {
	if strings.TrimSpace(input.ID) == "" {
		return nil, errors.New("queue id is required")
	}

	name := strings.TrimSpace(input.Name)

	if name == "" {
		return nil, errors.New("queue name is required")
	}

	if len(name) > 100 {
		return nil, errors.New("queue name must not exceed 100 characters")
	}

	queue := &Queue{
		ID:          input.ID,
		Name:        name,
		Description: input.Description,
		IsActive:    input.IsActive,
	}

	if err := s.repository.Update(ctx, queue); err != nil {
		return nil, err
	}

	return s.repository.FindByID(ctx, queue.ID)
}

func (s *Service) Delete(
	ctx context.Context,
	queueID string,
) error {
	if strings.TrimSpace(queueID) == "" {
		return errors.New("queue id is required")
	}

	return s.repository.Delete(ctx, queueID)
}
