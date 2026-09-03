package auth

import (
	"fmt"
	"net/mail"
	"strings"
	"unicode/utf8"
)

const (
	minPasswordLength = 8
	maxPasswordLength = 120

	minNameLength = 2
	maxNameLength = 100
)

func validateFullName(name string) error {
	name = strings.TrimSpace(name)

	if name == "" {
		return fmt.Errorf("full name is required")
	}

	length := utf8.RuneCountInString(name)

	if length < minNameLength {
		return fmt.Errorf("full name must be at least %d characters", minNameLength)
	}

	if length > maxNameLength {
		return fmt.Errorf("full name must not exceed %d characters", maxNameLength)
	}

	return nil
}

func validateEmail(email string) error {
	email = strings.TrimSpace(email)

	if email == "" {
		return fmt.Errorf("email is required")
	}

	if len(email) > 255 {
		return fmt.Errorf("email must not exceed 255 characters")
	}

	address, err := mail.ParseAddress(email)

	if err != nil || address.Address != email {
		return fmt.Errorf("invalid email address")
	}

	return nil
}

func validatePassword(password string) error {
	length := utf8.RuneCountInString(password)

	if length < minPasswordLength {
		return fmt.Errorf("password must be at least %d characters", minPasswordLength)
	}

	if length > maxPasswordLength {
		return fmt.Errorf("password must not exceed %d characters", maxPasswordLength)
	}

	return nil
}