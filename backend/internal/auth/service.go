package auth

import (
	"context"
	"fmt"

	"github.com/Saif724/STAQ/backend/internal/auth/dto"
	"github.com/Saif724/STAQ/backend/internal/users"
)

type Service struct {
	usersService *users.Service
}

func NewService(usersService *users.Service) *Service {
	return &Service{
		usersService: usersService,
	}
}

func (s *Service) Register(
	ctx context.Context,
	req dto.RegisterRequest,
) (*dto.RegisterResponse, error) {
	if err := validateFullName(req.FullName); err != nil {
		return nil, err
	}

	if err := validateEmail(req.Email); err != nil {
		return nil, err
	}

	if err := validatePassword(req.Password); err != nil {
		return nil, err
	}

	user, err := s.usersService.Create(
		ctx,
		req.FullName,
		req.Email,
		req.Password,
	)

	if err != nil {
		return nil, fmt.Errorf("registration failed: %w", err)
	}

	return &dto.RegisterResponse{
		ID: user.ID,
		FullName: user.FullName,
		Email: user.Email,
		EmailVerified: user.EmailVerified,
	}, nil
}