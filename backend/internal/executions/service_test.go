package executions

import (
	"errors"
	"testing"
)

func TestListByUserRejectsInvalidInputs(t *testing.T) {
	tests := []struct {
		name      string
		userID    string
		status    string
		page      int64
		pageSize  int64
		wantError error
	}{
		{
			name:      "invalid status",
			userID:    "user-id",
			status:    "DONE",
			page:      1,
			pageSize:  DefaultExecutionPageSize,
			wantError: ErrInvalidExecutionFilter,
		},
		{
			name:      "page starts at one",
			userID:    "user-id",
			status:    StatusSuccess,
			page:      0,
			pageSize:  DefaultExecutionPageSize,
			wantError: ErrInvalidPagination,
		},
		{
			name:      "page size maximum",
			userID:    "user-id",
			status:    StatusSuccess,
			page:      1,
			pageSize:  MaxExecutionPageSize + 1,
			wantError: ErrInvalidPagination,
		},
		{
			name:      "page offset overflow",
			userID:    "user-id",
			status:    StatusSuccess,
			page:      1<<63 - 1,
			pageSize:  MaxExecutionPageSize,
			wantError: ErrInvalidPagination,
		},
	}

	service := &Service{}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := service.ListByUser(
				t.Context(),
				test.userID,
				"",
				test.status,
				test.page,
				test.pageSize,
			)
			if err == nil {
				t.Fatal("ListByUser() returned no error")
			}
			if !errors.Is(err, test.wantError) {
				t.Fatalf("ListByUser() error = %v, want %v", err, test.wantError)
			}
		})
	}
}
