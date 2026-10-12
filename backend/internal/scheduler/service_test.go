package scheduler

import (
	"testing"

	"github.com/Saif724/STAQ/backend/internal/tasks"
)

func TestShouldDispatch(t *testing.T) {
	tests := []struct {
		name       string
		taskStatus string
		stale      bool
		want       bool
	}{
		{
			name:       "active task with current occurrence",
			taskStatus: tasks.StatusActive,
			want:       true,
		},
		{
			name:       "paused task",
			taskStatus: tasks.StatusPaused,
			want:       false,
		},
		{
			name:       "archived task",
			taskStatus: tasks.StatusArchived,
			want:       false,
		},
		{
			name:       "active task with stale occurrence",
			taskStatus: tasks.StatusActive,
			stale:      true,
			want:       false,
		},
		{
			name:       "unknown task status",
			taskStatus: "unknown",
			want:       false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := shouldDispatch(test.taskStatus, test.stale); got != test.want {
				t.Fatalf("shouldDispatch(%q, %t) = %t, want %t", test.taskStatus, test.stale, got, test.want)
			}
		})
	}
}
