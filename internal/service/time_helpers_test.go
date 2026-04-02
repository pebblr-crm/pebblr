package service

import (
	"testing"
	"time"
)

func TestCalendarMonths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		from time.Time
		to   time.Time
		want int
	}{
		{
			name: "same month",
			from: time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC),
			to:   time.Date(2025, 3, 31, 0, 0, 0, 0, time.UTC),
			want: 1,
		},
		{
			name: "two months",
			from: time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC),
			to:   time.Date(2025, 4, 15, 0, 0, 0, 0, time.UTC),
			want: 2,
		},
		{
			name: "cross year",
			from: time.Date(2024, 11, 1, 0, 0, 0, 0, time.UTC),
			to:   time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC),
			want: 4,
		},
		{
			name: "reversed range returns 1",
			from: time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
			to:   time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC),
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := calendarMonths(tt.from, tt.to)
			if got != tt.want {
				t.Errorf("calendarMonths(%v, %v) = %d, want %d", tt.from, tt.to, got, tt.want)
			}
		})
	}
}
