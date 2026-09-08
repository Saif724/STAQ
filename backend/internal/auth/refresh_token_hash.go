package auth

import (
	"crypto/sha256"
	"encoding/hex"
)

func hashRefreshToken(token string) string {
	hash := sha256.Sum256([]byte(token))

	return hex.EncodeToString(hash[:])
}
