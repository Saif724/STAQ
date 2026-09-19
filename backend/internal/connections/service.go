package connections

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrUnsupportedProvider = errors.New("unsupported connection provider")
)

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) GetByID(
	ctx context.Context,
	id string,
	userID string,
) (*Connection, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("connection id is required")
	}

	return s.repository.FindByIDAndUser(ctx, id, userID)
}

func (s *Service) List(
	ctx context.Context,
	userID string,
) ([]*Connection, error) {
	return s.repository.FindByUser(ctx, userID)
}

func (s *Service) Delete(
	ctx context.Context,
	id string,
	userID string,
) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("connection id is required")
	}

	return s.repository.Delete(ctx, id, userID)
}

func (s *Service) Create(
	ctx context.Context,
	userID string,
	provider string,
	providerAccountID string,
	accountEmail string,
	accessTokenEncrypted string,
	refreshTokenEncrypted *string,
	tokenExpiresAt *time.Time,
	scopes []string,
) (*Connection, error) {
	if err := validateRequired(userID, "user id"); err != nil {
		return nil, err
	}

	provider = strings.ToUpper(strings.TrimSpace(provider))

	if err := validateRequired(provider, "provider"); err != nil {
		return nil, err
	}

	if provider != ProviderGoogle {
		return nil, ErrUnsupportedProvider
	}

	if err := validateRequired(providerAccountID, "provider account id"); err != nil {
		return nil, err
	}

	if err := validateRequired(accountEmail, "account email"); err != nil {
		return nil, err
	}

	if err := validateRequired(accessTokenEncrypted, "access token"); err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	connection := &Connection{
		ID:                    uuid.NewString(),
		UserID:                userID,
		Provider:              provider,
		ProviderAccountID:     providerAccountID,
		AccountEmail:          accountEmail,
		AccessTokenEncrypted:  accessTokenEncrypted,
		RefreshTokenEncrypted: refreshTokenEncrypted,
		TokenExpiresAt:        tokenExpiresAt,
		Scopes:                scopes,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	if err := s.repository.Create(ctx, connection); err != nil {
		return nil, fmt.Errorf("failed to create connection: %w", err)
	}

	return connection, nil
}

func validateRequired(value string, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", fieldName)
	}

	return nil
}

func (s *Service) FindByIDAndUser(
	ctx context.Context,
	connectionID string,
	userID string,
) (*Connection, error) {
	if s == nil || s.repository == nil {
		return nil, errors.New("connection service is not configured")
	}

	return s.repository.FindByIDAndUser(ctx, connectionID, userID)
}

func (s *Service) UpdateToken(
	ctx context.Context,
	connectionID string,
	accessTokenEncrypted string,
	refreshTokenEncrypted *string,
	tokenExpiresAt *time.Time,
) error {
	if s == nil || s.repository == nil {
		return errors.New("connection service is not configured")
	}

	return s.repository.UpdateTokens(ctx, connectionID, accessTokenEncrypted, refreshTokenEncrypted, tokenExpiresAt)
}
