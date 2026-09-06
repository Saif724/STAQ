package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/Saif724/STAQ/backend/internal/auth/dto"
	"github.com/Saif724/STAQ/backend/internal/users"
	"github.com/Saif724/STAQ/backend/pkg/hash"
	"github.com/Saif724/STAQ/backend/pkg/jwt"
)

type Service struct {
	usersService *users.Service
	jwtManager *jwt.Manager
}

func NewService(usersService *users.Service,jwtManager *jwt.Manager) *Service {
	return &Service{
		usersService: usersService,
		jwtManager: jwtManager,
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

func (s *Service) Login(
	ctx context.Context,
	req dto.LoginRequest,
) (*users.User, error) {
	if err := validateEmail(req.Email); err != nil {
		return nil, err
	}

	if err := validatePassword(req.Password); err != nil {
		return nil, err
	}

	user, err := s.usersService.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("login failed: %w", err)
	}

	if !user.IsActive {
		return nil, errors.New("account is inactive")
	}

	if !hash.ComparePassword(user.PasswordHash, req.Password) {
		return nil, errors.New("invalid email or password")
	}

	return user, nil
}