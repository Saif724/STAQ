package email

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const gmailSendURL = "https://gmail.googleapis.com/gmail/v1/users/me/messages/send"

type GmailSender struct {
	config *oauth2.Config
	token  *oauth2.Token
	client *http.Client
}

func NewGmailSender(
	clientID string,
	clientSecret string,
	refreshToken string,
) *GmailSender {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		Scopes: []string{
			"https://www.googleapis.com/auth/gmail.send",
		},
	}

	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}

	return &GmailSender{
		config: config,
		token:  token,
		client: &http.Client{},
	}
}

func (s *GmailSender) Send(
	to string,
	subject string,
	body string,
) error {
	to = strings.TrimSpace(to)
	subject = strings.TrimSpace(subject)

	if to == "" {
		return fmt.Errorf("recipient email is required")
	}

	if subject == "" {
		return fmt.Errorf("email subject is required")
	}

	if body == "" {
		return fmt.Errorf("email body is required")
	}

	if strings.ContainsAny(to, "\r\n") ||
		strings.ContainsAny(subject, "\r\n") {
		return fmt.Errorf("invalid email headers")
	}

	tokenSource := s.config.TokenSource(
		context.Background(),
		s.token,
	)

	token, err := tokenSource.Token()
	if err != nil {
		return fmt.Errorf("failed to refresh Gmail access token: %w", err)
	}

	rawMessage := buildRawEmail(to, subject, body)

	payload := struct {
		Raw string `json:"raw"`
	}{
		Raw: base64.RawURLEncoding.EncodeToString([]byte(rawMessage)),
	}

	requestBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to encode Gmail request: %w", err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		gmailSendURL,
		bytes.NewReader(requestBody),
	)
	if err != nil {
		return fmt.Errorf("failed to create Gmail request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "STAQ/1.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Gmail request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Gmail returned status %d", resp.StatusCode)
	}

	return nil
}

func buildRawEmail(to, subject, body string) string {
	return strings.Join([]string{
		"To: " + to,
		"Subject: " + subject,
		"Content-Type: text/plain; charset=UTF-8",
		"MIME-Version: 1.0",
		"",
		body,
	}, "\r\n")
}

var _ Sender = (*GmailSender)(nil)
