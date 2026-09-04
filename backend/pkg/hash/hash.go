package hash

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	memory      uint32 = 64 * 1024
	iterations  uint32 = 3
	parallelism uint8  = 2
	keyLength   uint32 = 32
	saltLength         = 16
)

func HashPassword(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	salt := make([]byte, saltLength)

	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		iterations,
		memory,
		parallelism,
		keyLength,
	)

	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf(
		"$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		memory,
		iterations,
		parallelism,
		encodedSalt,
		encodedHash,
	), nil
}

func ComparePassword(password, encodedHash string) bool {
	parts := strings.Split(encodedHash, "$")

	if len(parts) != 6 || parts[0] != "" {
		return false
	}

	if parts[1] != "argon2id" {
		return false
	}

	if parts[2] != "v=19" {
		return false
	}

	params := strings.Split(parts[3], ",")

	if len(params) != 3 {
		return false
	}

	var (
		memoryCost uint32
		timeCost   uint32
		threads    uint8
	)

	for _, param := range params {
		parts := strings.SplitN(param, "=", 2)

		if len(parts) != 2 {
			return false
		}

		value, err := strconv.ParseUint(parts[1], 10, 32)
		if err != nil {
			return false
		}

		switch parts[0] {
		case "m":
			memoryCost = uint32(value)

		case "t":
			timeCost = uint32(value)

		case "p":
			if value > 255 {
				return false
			}

			threads = uint8(value)

		default:
			return false
		}
	}

	if memoryCost == 0 || timeCost == 0 || threads == 0 {
		return false
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false
	}

	actualHash := argon2.IDKey(
		[]byte(password),
		salt,
		iterations,
		memory,
		parallelism,
		uint32(len(expectedHash)),
	)

	return subtle.ConstantTimeCompare(actualHash, expectedHash) == 1
}
