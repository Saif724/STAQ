package users

import (
	"errors"
	"testing"
)

func TestUpdateProfileValidatesFieldsBeforePersistence(t *testing.T) {
	tests := []struct {
		name     string
		userID   string
		fullName string
		timezone string
		wantErr  error
	}{
		{
			name:    "missing user",
			wantErr: errors.New("user id is required"),
		},
		{
			name:     "short name",
			userID:   "user-id",
			fullName: "A",
			timezone: "UTC",
			wantErr:  ErrInvalidProfile,
		},
		{
			name:     "invalid timezone",
			userID:   "user-id",
			fullName: "Ada Lovelace",
			timezone: "Not/A_Timezone",
			wantErr:  ErrInvalidProfile,
		},
	}

	service := &Service{}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := service.UpdateProfile(t.Context(), test.userID, test.fullName, test.timezone)
			if err == nil {
				t.Fatal("UpdateProfile() returned no error")
			}
			if test.wantErr == ErrInvalidProfile && !errors.Is(err, ErrInvalidProfile) {
				t.Fatalf("UpdateProfile() error = %v, want invalid profile", err)
			}
		})
	}
}
