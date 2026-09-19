package email

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/Saif724/STAQ/backend/internal/connections"
	"golang.org/x/oauth2"
)

const gmailSendURL = "https://gmail.googleapis.com/gmail/v1/users/me/messages/send"

type gmailSendRequest struct {
	Raw string `json:"raw"`
}

func (s *Service) SendEmail(
	ctx context.Context,
	userID string,
	connectionID string,
	to string,
	subject string,
	body string,
) error {
	userID = strings.TrimSpace(userID)
	connectionID = strings.TrimSpace(connectionID)
	to = strings.TrimSpace(to)
	subject = strings.TrimSpace(subject)

	if userID == "" {
		return fmt.Errorf("user id is required")
	}

	if connectionID == "" {
		return fmt.Errorf("connection id is required")
	}

	if to == "" {
		return fmt.Errorf("recipient email is required")
	}

	if subject == "" {
		return fmt.Errorf("email subject is required")
	}

	if strings.TrimSpace(body) == "" {
		return fmt.Errorf("email body is required")
	}

	if containsHeaderInjection(to) {
		return fmt.Errorf("recipient contains invalid characters")
	}

	if containsHeaderInjection(subject) {
		return fmt.Errorf("subject contains invalid characters")
	}

	if s == nil {
		return fmt.Errorf("email service is not configured")
	}

	if s.connections == nil {
		return fmt.Errorf("connection service is not configured")
	}

	if s.encryptor == nil {
		return fmt.Errorf("token encryptor is not configured")
	}

	if s.gmailOAuth == nil || s.gmailOAuth.config == nil {
		return fmt.Errorf("Gmail OAuth service is not configured")
	}

	connection, err := s.connections.FindByIDAndUser(
		ctx,
		connectionID,
		userID,
	)
	if err != nil {
		return fmt.Errorf("failed to find Gmail connection: %w", err)
	}

	if connection.Provider != connections.ProviderGoogle {
		return fmt.Errorf("unsupported connection provider: %s", connection.Provider)
	}

	accessToken, err := s.encryptor.Decrypt(connection.AccessTokenEncrypted)
	if err != nil {
		return fmt.Errorf("failed to decrypt access token: %w", err)
	}

	currentToken := &oauth2.Token{
		AccessToken: accessToken,
	}

	if connection.RefreshTokenEncrypted != nil &&
		strings.TrimSpace(*connection.RefreshTokenEncrypted) != "" {
		refreshToken, err := s.encryptor.Decrypt(*connection.RefreshTokenEncrypted)
		if err != nil {
			return fmt.Errorf("failed to decrypt refresh token: %w", err)
		}

		currentToken.RefreshToken = refreshToken
	}

	if connection.TokenExpiresAt != nil {
		currentToken.Expiry = *connection.TokenExpiresAt
	}

	token, err := s.getValidGmailToken(
		ctx,
		connection,
		currentToken,
	)
	if err != nil {
		return err
	}

	rawMessage, err := buildRawEmail(
		to,
		subject,
		body,
	)
	if err != nil {
		return fmt.Errorf("failed to build email: %w", err)
	}

	encodedMessage := base64.RawURLEncoding.EncodeToString(
		[]byte(rawMessage),
	)

	payload := gmailSendRequest{
		Raw: encodedMessage,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to encode Gmail request: %w", err)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		gmailSendURL,
		strings.NewReader(string(payloadBytes)),
	)
	if err != nil {
		return fmt.Errorf("failed to create Gmail request: %w", err)
	}

	request.Header.Set("Authorization", "Bearer "+token.AccessToken)
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("failed to send Gmail request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK ||
		response.StatusCode >= http.StatusMultipleChoices {

		errorBody, _ := io.ReadAll(io.LimitReader(response.Body, 4096))

		return fmt.Errorf("Gmail API returned status %d: %s", response.StatusCode, strings.TrimSpace(string(errorBody)))
	}

	return nil
}

func (s *Service) getValidGmailToken(
	ctx context.Context,
	connection *connections.Connection,
	currentToken *oauth2.Token,
) (*oauth2.Token, error) {
	if currentToken.Valid() {
		return currentToken, nil
	}

	if strings.TrimSpace(currentToken.RefreshToken) == "" {
		return nil, fmt.Errorf("Gmail access token has expired and no refresh token is available")
	}

	tokenSource := s.gmailOAuth.config.TokenSource(
		ctx,
		currentToken,
	)

	refreshedToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh Gmail access token: %w", err)
	}

	encryptedAccessToken, err := s.encryptor.Encrypt(refreshedToken.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt refreshed access token: %w", err)
	}

	var encryptedRefreshToken *string

	if refreshedToken.RefreshToken != "" {
		encrypted, err := s.encryptor.Encrypt(refreshedToken.RefreshToken)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt refreshed refresh token: %w", err)
		}

		encryptedRefreshToken = &encrypted
	}

	var expiresAt *time.Time

	if !refreshedToken.Expiry.IsZero() {
		expiry := refreshedToken.Expiry
		expiresAt = &expiry
	}

	err = s.connections.UpdateToken(
		ctx,
		connection.ID,
		encryptedAccessToken,
		encryptedRefreshToken,
		expiresAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to save refreshed Gmail tokens: %w", err)
	}

	return refreshedToken, nil
}

func buildRawEmail(
	to string,
	subject string,
	body string,
) (string, error) {
	encodedSubject := mime.QEncoding.Encode(
		"UTF-8",
		subject,
	)

	var builder strings.Builder

	builder.WriteString("To: ")
	builder.WriteString(to)
	builder.WriteString("\r\n")

	builder.WriteString("Subject: ")
	builder.WriteString(encodedSubject)
	builder.WriteString("\r\n")

	builder.WriteString("MIME-Version: 1.0\r\n")
	builder.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	builder.WriteString("Content-Transfer-Encoding: 8bit\r\n")
	builder.WriteString("\r\n")
	builder.WriteString(body)

	return builder.String(), nil
}

func containsHeaderInjection(value string) bool {
	return strings.ContainsAny(value, "\r\n")
}
