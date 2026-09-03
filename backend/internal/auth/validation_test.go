package auth

import "testing"

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name string
		password string
		valid bool
	}{
		{
			name: "valid password",
			password: "secure-password",
			valid: true,
		},{
			name: "too short",
			password: "1234567",
			valid: false,
		},{
			name: "exactly eight characters",
			password: "12345678",
			valid: true,
		},{
			name: "too long",
			password: string(make([]byte, 129)),
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePassword(tt.password)

			if (err == nil) != tt.valid {
				t.Fatalf("expected valid=%v, got error=%v", tt.valid, err)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name string
		email string
		valid bool
	}{
		{
			name: "valid email",
			email: "user@example.com",
			valid: true,
		},{
			name: "empty email",
			email: "",
			valid: false,
		},{
			name: "invalid email",
			email: "not-an-email",
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEmail(tt.email)

			if (err == nil) != tt.valid {
				t.Fatalf("expected valid=%v, got error=%v", tt.valid, err)
			}
		})
	}
}

func TestValidateFullName(t *testing.T) {
	tests := []struct {
		name string
		fullName string
		valid bool
	}{
		{
			name: "valid name",
			fullName: "Saif Ahmed",
			valid: true,
		},{
			name: "empty name",
			fullName: "",
			valid: false,
		},{
			name: "one character",
			fullName: "A",
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFullName(tt.fullName)

			if (err == nil) != tt.valid {
				t.Fatalf("expected valid=%v, got error=%v", tt.valid, err)
			}
		})
	}
}