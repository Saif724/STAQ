package connections

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestConnectionResponseDoesNotExposeCredentials(t *testing.T) {
	secretAccessToken := "encrypted-access-token"
	secretRefreshToken := "encrypted-refresh-token"
	connection := &Connection{
		ID:                    "connection-id",
		UserID:                "user-id",
		Provider:              ProviderGoogle,
		ProviderAccountID:     "provider-account-id",
		AccountEmail:          "person@example.test",
		AccessTokenEncrypted:  secretAccessToken,
		RefreshTokenEncrypted: &secretRefreshToken,
		TokenExpiresAt:        ptrTime(time.Date(2026, 10, 13, 0, 0, 0, 0, time.UTC)),
		Scopes:                []string{"gmail.send"},
		CreatedAt:             time.Now().UTC(),
		UpdatedAt:             time.Now().UTC(),
	}

	payload, err := json.Marshal(toConnectionResponse(connection))
	if err != nil {
		t.Fatalf("marshal connection response: %v", err)
	}
	encoded := string(payload)
	if strings.Contains(encoded, secretAccessToken) || strings.Contains(encoded, secretRefreshToken) {
		t.Fatalf("connection response contains credential fields: %s", encoded)
	}
	if strings.Contains(encoded, "provider_account_id") || strings.Contains(encoded, "user_id") {
		t.Fatalf("connection response contains unnecessary identifiers: %s", encoded)
	}
}

func ptrTime(value time.Time) *time.Time {
	return &value
}
