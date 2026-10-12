package email

import (
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestCompleteGmailAuthorizationRedirectsWithFixedResult(t *testing.T) {
	handler := &Handler{frontendURL: "https://app.example.test"}
	request := httptest.NewRequest("GET", "/connections/google/callback?error=access_denied", nil)
	response := httptest.NewRecorder()

	handler.CompleteGmailAuthorization(response, request)

	if response.Code != 303 {
		t.Fatalf("status = %d, want %d", response.Code, 303)
	}
	location, err := url.Parse(response.Header().Get("Location"))
	if err != nil {
		t.Fatalf("parse redirect location: %v", err)
	}
	if location.Scheme != "https" || location.Host != "app.example.test" {
		t.Fatalf("redirect origin = %s://%s, unexpected", location.Scheme, location.Host)
	}
	if location.Path != "/dashboard/connections" {
		t.Fatalf("redirect path = %q, want /dashboard/connections", location.Path)
	}
	if location.Query().Get("gmail") != "error" || location.Query().Get("reason") != "denied" {
		t.Fatalf("unexpected result query: %s", location.RawQuery)
	}
	if strings.Contains(location.RawQuery, "access_denied") || location.Query().Get("code") != "" {
		t.Fatalf("redirect leaked OAuth callback data: %s", location.RawQuery)
	}
}

func TestCompleteGmailAuthorizationStopsOnMissingParameters(t *testing.T) {
	handler := &Handler{frontendURL: "https://app.example.test"}
	request := httptest.NewRequest("GET", "/connections/google/callback", nil)
	response := httptest.NewRecorder()

	handler.CompleteGmailAuthorization(response, request)

	if response.Code != 303 {
		t.Fatalf("status = %d, want %d", response.Code, 303)
	}
	location, err := url.Parse(response.Header().Get("Location"))
	if err != nil {
		t.Fatalf("parse redirect location: %v", err)
	}
	if location.Query().Get("reason") != "invalid_callback" {
		t.Fatalf("reason = %q, want invalid_callback", location.Query().Get("reason"))
	}
}
