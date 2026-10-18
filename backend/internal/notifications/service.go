package notifications

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	defaultListLimit = 30
	maxListLimit     = 100
)

var ErrNotFound = errors.New("notification not found")

type ListResult struct {
	Items       []Notification `json:"items"`
	UnreadCount int            `json:"unread_count"`
}

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

// CreateReminder records a reminder for the task owner. Repeating the same
// (executionID, actionID) pair is a no-op, so a retried run cannot remind twice.
func (s *Service) CreateReminder(
	ctx context.Context,
	userID, taskID, executionID, actionID, message string,
) error {
	message = strings.TrimSpace(message)

	if strings.TrimSpace(userID) == "" {
		return errors.New("user id is required")
	}

	if message == "" {
		return errors.New("reminder message is required")
	}

	return s.repository.Create(ctx, &Notification{
		ID:          uuid.NewString(),
		UserID:      userID,
		TaskID:      &taskID,
		ExecutionID: &executionID,
		ActionID:    &actionID,
		Type:        TypeReminder,
		Message:     message,
		CreatedAt:   time.Now().UTC(),
	})
}

func (s *Service) List(
	ctx context.Context,
	userID string,
	limit int,
) (*ListResult, error) {
	if limit <= 0 {
		limit = defaultListLimit
	}
	if limit > maxListLimit {
		limit = maxListLimit
	}

	items, err := s.repository.ListByUser(ctx, userID, limit)
	if err != nil {
		return nil, err
	}

	unread, err := s.repository.CountUnread(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &ListResult{Items: items, UnreadCount: unread}, nil
}

func (s *Service) MarkRead(ctx context.Context, id, userID string) error {
	if _, err := uuid.Parse(id); err != nil {
		return ErrNotFound
	}

	found, err := s.repository.MarkRead(ctx, id, userID)
	if err != nil {
		return err
	}

	if !found {
		return ErrNotFound
	}

	return nil
}

func (s *Service) MarkAllRead(ctx context.Context, userID string) error {
	return s.repository.MarkAllRead(ctx, userID)
}
