package email

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
)

const googleUserInfoURL = "https://www.googleapis.com/oauth2/v3/userinfo"

type GoogleAccount struct {
	ID            string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
}

func FetchGoogleAccount(
	ctx context.Context,
	oauthConfig *oauth2.Config,
	token *oauth2.Token,
) (*GoogleAccount, error) {
	if oauthConfig == nil {
		return nil, errors.New("Google OAuth configuration is required")
	}

	if token == nil {
		return nil, errors.New("OAuth token is required")
	}

	client := oauthConfig.Client(ctx, token)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		googleUserInfoURL,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Google userinfo request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Google account: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("Google userinfo returned status %d", resp.StatusCode)
	}

	var account GoogleAccount

	if err := json.NewDecoder(req.Body).Decode(&account); err != nil {
		return nil, fmt.Errorf("failed to decode Google account: %w", err)
	}

	if account.ID == "" {
		return nil, errors.New("Google account ID is missing")
	}

	if account.Email == "" {
		return nil, errors.New("Google account email is missing")
	}

	if !account.EmailVerified {
		return nil, errors.New("Google account email is not verified")
	}

	return &account, nil
}
