package response

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestErrorJSON(t *testing.T) {
	rec := httptest.NewRecorder()

	ErrorJSON(
		rec,
		http.StatusBadRequest,
		"INVALID_REQUEST",
		"Invalid request",
	)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}

	if contentType := rec.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("expected application/json, got %q", contentType)
	}

	expected := `{"error":{"code":"INVALID_REQUEST","message":"Invalid request"}}` + "\n"

	if rec.Body.String() != expected {
		t.Fatalf("unexpected response: %q", rec.Body.String())
	}
}