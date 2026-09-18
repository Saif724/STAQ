package email

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Saif724/STAQ/backend/internal/connections"
	"github.com/Saif724/STAQ/backend/pkg/securetoken"
)

type Service struct {
	gmailOAuth  *GmailOAuthService
	connections *connections.Service
	encryptor   *securetoken.Encryptor
}

func NewService(
	gmailOAuth *GmailOAuthService,
	connections *connections.Service,
	encryptor *securetoken.Encryptor,
) *Service {
	return &Service{
		gmailOAuth:  gmailOAuth,
		connections: connections,
		encryptor:   encryptor,
	}
}

func (s *Service) BeginGmailAuthorization(
	ctx context.Context,
	userID string,
) (string, error) {
	if s.gmailOAuth == nil {
		return "", fmt.Errorf("Gmail OAuth service is not configured")
	}

	return s.gmailOAuth.AuthorizationURL(ctx, userID)
}

func (s *Service) CompleteGmailAuthorization(
	ctx context.Context,
	state string,
	code string,
) (string, *connections.Connection, error) {
	if s.gmailOAuth == nil {
		return "", nil, fmt.Errorf("Gmail OAuth service is not configured")
	}

	if s.connections == nil {
		return "", nil, fmt.Errorf("connection service is not configured")
	}

	if s.encryptor == nil {
		return "", nil, fmt.Errorf("token encryptor is not configured")
	}

	userID, err := s.gmailOAuth.ConsumeState(ctx, state)
	if err != nil {
		return "", nil, err
	}

	token, err := s.gmailOAuth.ExchangeCode(ctx, code)
	if err != nil {
		return "", nil, err
	}

	account, err := FetchGoogleAccount(
		ctx,
		s.gmailOAuth.config,
		token,
	)
	if err != nil {
		return "", nil, err
	}

	encryptedAccessToken, err := s.encryptor.Encrypt(token.AccessToken)
	if err != nil {
		return "", nil, fmt.Errorf("failed to encrypt access token: %w", err)
	}

	var encryptedRefreshToken *string

	if token.RefreshToken != "" {
		encrypted, err := s.encryptor.Encrypt(token.RefreshToken)
		if err != nil {
			return "", nil, fmt.Errorf("failed to encrypt refresh token: %w", err)
		}

		encryptedRefreshToken = &encrypted
	}

	var tokenExpiredAt *time.Time

	if !token.Expiry.IsZero() {
		expiry := token.Expiry
		tokenExpiredAt = &expiry
	}

	connection, err := s.connections.Create(
		ctx,
		userID,
		connections.ProviderGoogle,
		account.ID,
		account.Email,
		encryptedAccessToken,
		encryptedRefreshToken,
		tokenExpiredAt,
		[]string{
			"https://www.googleapis.com/auth/gmail.send",
		},
	)

	if err != nil {
		return "", nil, fmt.Errorf("failed to save Gmail connection: %w", err)
	}

	return strings.TrimSpace(userID), connection, nil
}
