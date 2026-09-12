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
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const googleProvider = "google"

type GoogleUser struct {
	ID            string `json:"id"`
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
}

func NewOAuthService(
	usersService *users.Service,
	oauthRepository *OAuthAccountRepository,
	jwtManager *jwt.Manager,
	refreshTokenRepo *RefreshTokenRepository,
	cfg config.GoogleConfig,
) *OAuthService {

	return &OAuthService{
		usersService:     usersService,
		oauthRepository:  oauthRepository,
		jwtManager:       jwtManager,
		refreshTokenRepo: refreshTokenRepo,
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

func (s *OAuthService) AuthorizationURL() (string, string, error) {
	state, err := generateOAuthState()
	if err != nil {
		return "", "", err
	}

	url := s.googleConfig.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
	)

	return url, state, nil
}

func (s *OAuthService) GoogleCallback(
	ctx context.Context,
	code string,
) (accessToken string, refreshToken string, err error) {
	if strings.TrimSpace(code) == "" {
		return "", "", errors.New("authorization code is required")
	}

	token, err := s.googleConfig.Exchange(ctx, code)
	if err != nil {
		return "", "", fmt.Errorf("failed to exchange google authorization code: %w", err)
	}

	client := s.googleConfig.Client(ctx, token)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://www.googleapis.com/oauth2/v3/userinfo",
		nil,
	)
	if err != nil {
		return "", "", fmt.Errorf("failed to create google userinfo request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch google user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", "", fmt.Errorf("google user info returned status %d", resp.StatusCode)
	}

	var googleUser GoogleUser
	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		return "", "", fmt.Errorf("failed to decode google user info: %w", err)
	}

	if googleUser.ID == "" || googleUser.Email == "" {
		return "", "", errors.New("google account information is incomplete")
	}

	if !googleUser.EmailVerified {
		return "", "", errors.New("google email is not verified")
	}

	return s.authenticateGoogleUser(ctx, googleUser)
}

func (s *OAuthService) authenticateGoogleUser(
	ctx context.Context,
	googleUser GoogleUser,
) (string, string, error) {
	account, err := s.oauthRepository.FindByProvider(
		ctx,
		googleProvider,
		googleUser.ID,
	)

	if err == nil {
		return s.issueTokens(ctx, account.UserID)
	}

	if !errors.Is(err, ErrOAuthAccountNotFound) {
		return "", "", err
	}

	existingUser, err := s.usersService.GetByEmail(
		ctx,
		googleUser.Email,
	)

	if err == nil {
		if !existingUser.IsActive {
			return "", "", ErrAccountInactive
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
			return "", "", err
		}

		if !existingUser.EmailVerified {
			if err := s.usersService.MarkEmailVerified(
				ctx,
				existingUser.ID,
			); err != nil {
				return "", "", fmt.Errorf(
					"failed to verify user email: %w",
					err,
				)
			}
		}

		return s.issueTokens(ctx, existingUser.ID)
	}

	if !errors.Is(err, users.ErrUserNotFound) {
		return "", "", fmt.Errorf(
			"failed to find existing user: %w",
			err,
		)
	}

	randomPassword := uuid.NewString() + uuid.NewString()

	user, err := s.usersService.Create(
		ctx,
		googleUser.Name,
		googleUser.Email,
		randomPassword,
	)

	if err != nil {
		return "", "", fmt.Errorf(
			"failed to create oauth user: %w",
			err,
		)
	}

	if err := s.usersService.MarkEmailVerified(
		ctx,
		user.ID,
	); err != nil {
		return "", "", fmt.Errorf(
			"failed to verify oauth user email: %w",
			err,
		)
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
		return "", "", fmt.Errorf(
			"failed to create oauth account: %w",
			err,
		)
	}

	return s.issueTokens(ctx, user.ID)
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
		return "", "", fmt.Errorf(
			"failed to generate refresh token: %w",
			err,
		)
	}

	refreshTokenHash := hashRefreshToken(refreshToken)

	record := &RefreshToken{
		ID:        uuid.NewString(),
		UserID:    userID,
		TokenHash: refreshTokenHash,
		ExpiresAt: time.Now().UTC().Add(refreshTokenLifeTime),
		CreatedAt: time.Now().UTC(),
	}

	if err := s.refreshTokenRepo.Create(ctx, record); err != nil {
		return "", "", fmt.Errorf(
			"failed to store refresh token: %w",
			err,
		)
	}

	return accessToken, refreshToken, nil
}
