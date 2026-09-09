package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type ResendSender struct {
	apiKey string
	from   string
	client *http.Client
}

func NewResendSender(apiKey, from string) *ResendSender {
	return &ResendSender{
		apiKey: apiKey,
		from:   from,
		client: &http.Client{},
	}
}

type resendResponse struct {
	ID string `json:"id"`
}

func (s *ResendSender) Send(
	to string,
	subject string,
	body string,
) error {
	payload := map[string]interface{}{
		"from":    s.from,
		"to":      []string{to},
		"subject": subject,
		"text":    body,
	}

	requestBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to encode email request: %w", err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		"https://api.resend.com/emails",
		bytes.NewReader(requestBody),
	)
	if err != nil {
		return fmt.Errorf("failed to create email request: %w", err)
	}

	req.Header.Set("Authorizatin", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "STAQ/1.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send email request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("resend returned status %d", resp.StatusCode)
	}

	var result resendResponse

	if err := json.NewDecoder(req.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode resend response: %w", err)
	}

	return nil
}
