package tasksetups

import (
	"context"
	"errors"
	"testing"
)

func TestCreateRejectsInvalidActionCountBeforeDatabaseWork(t *testing.T) {
	service := &Service{}

	_, err := service.Create(context.Background(), "user-id", CreateInput{})
	if !errors.Is(err, ErrInvalidSetup) {
		t.Fatalf("Create() error = %v, want ErrInvalidSetup", err)
	}
}

func TestCreateRejectsTooManyActionsBeforeDatabaseWork(t *testing.T) {
	service := &Service{}
	input := CreateInput{Actions: make([]ActionInput, maxSetupActions+1)}

	_, err := service.Create(context.Background(), "user-id", input)
	if !errors.Is(err, ErrInvalidSetup) {
		t.Fatalf("Create() error = %v, want ErrInvalidSetup", err)
	}
}
