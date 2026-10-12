package dashboard

import (
	"math"
	"testing"
)

func TestCalculateSuccessRate(t *testing.T) {
	tests := []struct {
		name     string
		success  int64
		failed   int64
		timedOut int64
		want     float64
	}{
		{
			name: "no completed executions",
		},
		{
			name:    "running and cancelled are excluded by denominator inputs",
			success: 2,
			failed:  1,
			want:    200.0 / 3.0,
		},
		{
			name:     "timeouts count as unsuccessful completion",
			success:  1,
			timedOut: 1,
			want:     50,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := calculateSuccessRate(test.success, test.failed, test.timedOut)
			if math.Abs(got-test.want) > 1e-9 {
				t.Fatalf("calculateSuccessRate() = %f, want %f", got, test.want)
			}
		})
	}
}
