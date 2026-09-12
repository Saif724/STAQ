package auth

import "time"

type EmailVerification struct {
	ID         string
	UserID     string
	Token      string
	ExpiresAt  time.Time
	VerifiedAt *time.Time
	RevokedAt  *time.Time
	CreatedAt  time.Time
}
