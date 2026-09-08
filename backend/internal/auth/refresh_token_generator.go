package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

func generateRefreshToken() (string, error) {
	bytes := make([]byte, 32)

	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
