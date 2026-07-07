package config

import (
	"fmt"
)

func (c *Config) Validate() error {

	if c.Database.URL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}

	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}

	if c.App.Port == "" {
		return fmt.Errorf("PORT is required")
	}

	if c.App.Env == "" {
		return fmt.Errorf("APP_ENV is required")
	}

	return nil
}
