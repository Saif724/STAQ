package users

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Saif724/STAQ/backend/pkg/hash"
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

func (s *Service) Create(
	ctx context.Context,
	fullName string,
	email string,
	password string,
) (*User, error) {
	fullName = strings.TrimSpace(fullName)
	email = strings.ToLower(strings.TrimSpace(email))

	if fullName == "" {
		return nil, fmt.Errorf("full name is required")
	}

	if email == "" {
		return nil, fmt.Errorf("email is required")
	}

	if password == "" {
		return nil, fmt.Errorf("password is required")
	}

	passwordHash, err := hash.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	now := time.Now().UTC()

	user := &User{
		ID:            uuid.NewString(),
		FullName:      fullName,
		Email:         email,
		PasswordHash:  passwordHash,
		EmailVerified: false,
		IsActive:      true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.repository.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (s *Service) GetByID(
	ctx context.Context,
	id string,
) (*User, error) {
	id = strings.TrimSpace(id)

	if id == "" {
		return nil, fmt.Errorf("user id is required")
	}

	return s.repository.FindByID(ctx, id)
}

func (s *Service) GetByEmail(
	ctx context.Context,
	email string,
) (*User, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	if email == "" {
		return nil, fmt.Errorf("eamil is required")
	}

	return s.repository.FindByEmail(ctx, email)
}
