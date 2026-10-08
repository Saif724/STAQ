package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Saif724/STAQ/backend/internal/config"
	"github.com/Saif724/STAQ/backend/internal/users"
	"github.com/Saif724/STAQ/backend/pkg/jwt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	googleProvider         = "google"
	oauthStateLifetime     = 10 * time.Minute
	oauthLoginCodeLifetime = 60 * time.Second
)

var ErrInvalidOAuthLoginCode = errors.New("invalid or expired oauth login code")

type GoogleUser struct {
	ID            string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
}

type OAuthService struct {
	usersService     *users.Service
	oauthRepository  *OAuthAccountRepository
	jwtManager       *jwt.Manager
	refreshTokenRepo *RefreshTokenRepository
	googleConfig     *oauth2.Config
	redis            *redis.Client
}

func NewOAuthService(
	usersService *users.Service,
	oauthRepository *OAuthAccountRepository,
	jwtManager *jwt.Manager,
	refreshTokenRepo *RefreshTokenRepository,
	cfg config.GoogleConfig,
	redisClient *redis.Client,
) *OAuthService {
	return &OAuthService{
		usersService:     usersService,
		oauthRepository:  oauthRepository,
		jwtManager:       jwtManager,
		refreshTokenRepo: refreshTokenRepo,
		redis:            redisClient,
		googleConfig: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Scopes: []string{
				"openid",
				"profile",
				"email",
			},
			Endpoint: google.Endpoint,
		},
	}
}

func generateOAuthState() (string, error) {
	b := make([]byte, 32)

	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate oauth state: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (s *OAuthService) AuthorizationURL(
	ctx context.Context,
) (string, string, error) {
	state, err := generateOAuthState()
	if err != nil {
		return "", "", err
	}

	key := "oauth:state:" + state

	if err := s.redis.Set(ctx, key, "1", oauthStateLifetime).Err(); err != nil {
		return "", "", fmt.Errorf("failed to store oauth state: %w", err)
	}

	authURL := s.googleConfig.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
	)

	return authURL, state, nil
}

func (s *OAuthService) GoogleCallback(
	ctx context.Context,
	code string,
) (string, error) {
	if strings.TrimSpace(code) == "" {
		return "", errors.New("authorization code is required")
	}

	token, err := s.googleConfig.Exchange(ctx, code)
	if err != nil {
		return "", fmt.Errorf("failed to exchange google authorization code: %w", err)
	}

	client := s.googleConfig.Client(ctx, token)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://www.googleapis.com/oauth2/v3/userinfo",
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("failed to create google userinfo request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch google user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("google user info returned status %d", resp.StatusCode)
	}

	var googleUser GoogleUser
	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		return "", fmt.Errorf("failed to decode google user info: %w", err)
	}

	if googleUser.ID == "" || googleUser.Email == "" {
		return "", errors.New("google account information is incomplete")
	}

	if !googleUser.EmailVerified {
		return "", errors.New("google email is not verified")
	}

	return s.authenticateGoogleUser(ctx, googleUser)
}

func (s *OAuthService) authenticateGoogleUser(
	ctx context.Context,
	googleUser GoogleUser,
) (string, error) {
	account, err := s.oauthRepository.FindByProvider(
		ctx,
		googleProvider,
		googleUser.ID,
	)

	if err == nil {
		return account.UserID, nil
	}

	if !errors.Is(err, ErrOAuthAccountNotFound) {
		return "", err
	}

	existingUser, err := s.usersService.GetByEmail(
		ctx,
		googleUser.Email,
	)

	if err == nil {
		if !existingUser.IsActive {
			return "", ErrAccountInactive
		}

		account := &OAuthAccount{
			ID:             uuid.NewString(),
			UserID:         existingUser.ID,
			Provider:       googleProvider,
			ProviderUserID: googleUser.ID,
			CreatedAt:      time.Now().UTC(),
			UpdatedAt:      time.Now().UTC(),
		}

		if err := s.oauthRepository.Create(ctx, account); err != nil {
			return "", err
		}

		if !existingUser.EmailVerified {
			if err := s.usersService.MarkEmailVerified(
				ctx,
				existingUser.ID,
			); err != nil {
				return "", fmt.Errorf("failed to verify user email: %w", err)
			}
		}

		return existingUser.ID, nil
	}

	if !errors.Is(err, users.ErrUserNotFound) {
		return "", fmt.Errorf("failed to find existing user: %w", err)
	}

	randomPassword := uuid.NewString() + uuid.NewString()

	user, err := s.usersService.Create(
		ctx,
		googleUser.Name,
		googleUser.Email,
		randomPassword,
	)
	if err != nil {
		return "", fmt.Errorf("failed to create oauth user: %w", err)
	}

	if err := s.usersService.MarkEmailVerified(ctx, user.ID); err != nil {
		return "", fmt.Errorf("failed to verify oauth user email: %w", err)
	}

	newAccount := &OAuthAccount{
		ID:             uuid.NewString(),
		UserID:         user.ID,
		Provider:       googleProvider,
		ProviderUserID: googleUser.ID,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}

	if err := s.oauthRepository.Create(ctx, newAccount); err != nil {
		return "", fmt.Errorf("failed to create oauth account: %w", err)
	}

	return user.ID, nil
}

func (s *OAuthService) issueTokens(
	ctx context.Context,
	userID string,
) (string, string, error) {
	accessToken, err := s.jwtManager.GenerateAccessToken(userID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := generateRefreshToken()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	record := &RefreshToken{
		ID:        uuid.NewString(),
		UserID:    userID,
		TokenHash: hashRefreshToken(refreshToken),
		ExpiresAt: time.Now().UTC().Add(refreshTokenLifeTime),
		CreatedAt: time.Now().UTC(),
	}

	if err := s.refreshTokenRepo.Create(ctx, record); err != nil {
		return "", "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

func generateOAuthLoginCode() (string, error) {
	b := make([]byte, 32)

	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate oauth login code: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (s *OAuthService) CreateLoginCode(
	ctx context.Context,
	userID string,
) (string, error) {
	code, err := generateOAuthLoginCode()
	if err != nil {
		return "", err
	}

	key := "oauth:login-code:" + code

	if err := s.redis.Set(ctx, key, userID, oauthLoginCodeLifetime).Err(); err != nil {
		return "", fmt.Errorf("failed to store oauth login code: %w", err)
	}

	return code, nil
}

func (s *OAuthService) ExchangeLoginCode(
	ctx context.Context,
	code string,
) (string, string, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return "", "", ErrInvalidOAuthLoginCode
	}

	key := "oauth:login-code:" + code

	userID, err := s.redis.GetDel(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", "", ErrInvalidOAuthLoginCode
	}
	if err != nil {
		return "", "", fmt.Errorf("failed to consume oauth login code: %w", err)
	}
	if userID == "" {
		return "", "", ErrInvalidOAuthLoginCode
	}

	return s.issueTokens(ctx, userID)
}

func (s *OAuthService) ValidateOAuthState(
	ctx context.Context,
	state string,
) error {
	state = strings.TrimSpace(state)
	if state == "" {
		return errors.New("oauth state is required")
	}

	key := "oauth:state:" + state

	_, err := s.redis.GetDel(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return errors.New("invalid or expired oauth state")
	}
	if err != nil {
		return fmt.Errorf("failed to validate oauth state: %w", err)
	}

	return nil
}
