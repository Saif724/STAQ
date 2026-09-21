package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Saif724/STAQ/backend/internal/actions"
)

const maxResponseBodySize = 64 * 1024

type Executor struct {
	client *http.Client
}

func NewExecutor(client *http.Client) *Executor {
	if client == nil {
		client = &http.Client{}
	}
	return &Executor{
		client: client,
	}
}

type Configuration struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

func (e *Executor) Execute(
	ctx context.Context,
	executionContext actions.ExecutionContext,
	configuration []byte,
) (*actions.ExecutionResult, error) {
	var config Configuration

	if err := json.Unmarshal(configuration, &config); err != nil {
		return nil, errors.New("invalid http configuraion")
	}

	method := strings.ToUpper(strings.TrimSpace(config.Method))

	if method == "" {
		method = http.MethodGet
	}

	switch method {
	case http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodHead:
	default:
		return nil, errors.New("unsupported http method")
	}

	url := strings.TrimSpace(config.URL)
	if url == "" {
		return nil, errors.New("http url is required")
	}

	var body io.Reader

	if config.Body != "" {
		body = strings.NewReader(config.Body)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		method,
		url,
		body,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	for key, value := range config.Headers {
		req.Header.Set(key, value)
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBodySize+1))
	if err != nil {
		return nil, fmt.Errorf("failed to read http response: %w", err)
	}

	truncated := len(responseBody) > maxResponseBodySize
	if truncated {
		responseBody = responseBody[:maxResponseBodySize]
	}

	resultData := map[string]any{
		"status_code": resp.StatusCode,
		"body":        string(responseBody),
		"truncated":   truncated,
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return &actions.ExecutionResult{
			Message: "http action failed",
			Data:    resultData,
		}, fmt.Errorf("http request returned status %d", resp.StatusCode)
	}

	return &actions.ExecutionResult{
		Message: "http action executed",
		Data:    resultData,
	}, nil
}
