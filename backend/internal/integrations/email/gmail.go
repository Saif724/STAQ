package email

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/Saif724/STAQ/backend/internal/config"
	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	gmailProvider      = "GOOGLE"
	gmailStatePrefix   = "staq:gmail:oauth:state:"
	gmailStateLifetime = 10 * time.Minute
)

type GmailOAuthService struct {
	config *oauth2.Config
	redis  *redis.Client
}

func NewGmailOAuthService(
	cfg config.GoogleConfig,
	redisClient *redis.Client,
) *GmailOAuthService {
	return &GmailOAuthService{
		redis: redisClient,
		config: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.GmailRedirectURL,
			Scopes: []string{
				"openid",
				"profile",
				"email",
				"https://www.googleapis.com/auth/gmail.send",
			},
			Endpoint: google.Endpoint,
		},
	}
}

func (s *GmailOAuthService) AuthorizationURL(
	ctx context.Context,
	userID string,
) (string, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return "", fmt.Errorf("user id is required")
	}

	state, err := generateState()
	if err != nil {
		return "", err
	}

	key := gmailStatePrefix + state

	if err := s.redis.Set(
		ctx,
		key,
		userID,
		gmailStateLifetime,
	).Err(); err != nil {
		return "", fmt.Errorf("failed to store Gmail OAuth state: %w", err)
	}

	authURL := s.config.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "consent"),
	)

	return authURL, nil
}

func (s *GmailOAuthService) ConsumeState(
	ctx context.Context,
	state string,
) (string, error) {
	state = strings.TrimSpace(state)

	if state == "" {
		return "", fmt.Errorf("OAuth state is required")
	}

	key := gmailStatePrefix + state

	userID, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("invalid or expired OAuth state")
		}

		return "", fmt.Errorf("failed to retrieve OAuth state: %w", err)
	}

	if err := s.redis.Del(ctx, key).Err(); err != nil {
		return "", fmt.Errorf("failed to consume OAuth state: %w", err)
	}

	if strings.TrimSpace(userID) == "" {
		return "", fmt.Errorf("OAuth state contains an invalid user ID")
	}

	return userID, nil
}

func (s *GmailOAuthService) ExchangeCode(
	ctx context.Context,
	code string,
) (*oauth2.Token, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, fmt.Errorf("authorization code is required")
	}

	token, err := s.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange Gmail authorization code: %w", err)
	}

	return token, nil
}

func generateState() (string, error) {
	return randomURLSafeString(32)
}

func randomURLSafeString(length int) (string, error) {
	randomBytes := make([]byte, length)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate secure random state: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(randomBytes), nil
}

func (s *GmailOAuthService) Config() *oauth2.Config {
	return s.config
}
