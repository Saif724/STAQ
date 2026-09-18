package connections

import "time"

const (
	ProviderGoogle = "GOOGLE"
)

type Connection struct {
	ID                    string
	UserID                string
	Provider              string
	ProviderAccountID     string
	AccountEmail          string
	AccessTokenEncrypted  string
	RefreshTokenEncrypted *string
	TokenExpiresAt        *time.Time
	Scopes                []string
	CreatedAt             time.Time
	UpdatedAt             time.Time
}
