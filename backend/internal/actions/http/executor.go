package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/Saif724/STAQ/backend/internal/actions"
)

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

	if strings.TrimSpace(config.URL) == "" {
		return nil, errors.New("http url is required")
	}

	var body io.Reader

	if config.Body != "" {
		body = strings.NewReader(config.Body)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		method,
		config.URL,
		body,
	)

	if err != nil {
		return nil, errors.New("failed to create http request")
	}

	for key, value := range config.Headers {
		req.Header.Set(key, value)
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer req.Body.Close()

	return &actions.ExecutionResult{
		Message: "http action executed",
		Data: map[string]any{
			"status_code": resp.StatusCode,
		},
	}, nil
}
