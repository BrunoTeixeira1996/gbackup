package utils

import (
	"testing"
	"time"
)

func TestCalculateTotalElapsedTime(t *testing.T) {
	tests := []struct {
		name  string
		times []ElapsedTime
		want  float64
	}{
		{
			name:  "empty slice",
			times: []ElapsedTime{},
			want:  0,
		},
		{
			name:  "single entry",
			times: []ElapsedTime{{Target: "a", Value: 1.5}},
			want:  1.5,
		},
		{
			name: "multiple entries",
			times: []ElapsedTime{
				{Target: "a", Value: 1.5},
				{Target: "b", Value: 2.25},
				{Target: "c", Value: 0.25},
			},
			want: 4.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CalculateTotalElapsedTime(tt.times); got != tt.want {
				t.Errorf("CalculateTotalElapsedTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCurrentTime(t *testing.T) {
	want := time.Now().Format("2006-01-02")
	if got := CurrentTime(); got != want {
		t.Errorf("CurrentTime() = %q, want %q", got, want)
	}
}

func TestEpoch(t *testing.T) {
	got := Epoch()

	gotTime := time.Unix(got, 0)
	now := time.Now()

	if gotTime.Year() != now.Year() || gotTime.Month() != now.Month() || gotTime.Day() != now.Day() {
		t.Errorf("Epoch() = %v, want same date as %v", gotTime, now)
	}

	if gotTime.Hour() != 0 || gotTime.Minute() != 0 || gotTime.Second() != 0 {
		t.Errorf("Epoch() = %v, want midnight (00:00:00)", gotTime)
	}
}

func TestNextFridayAt13(t *testing.T) {
	tests := []struct {
		name string
		now  time.Time
		want time.Time
	}{
		{
			name: "monday before 13h -> next friday same week",
			now:  time.Date(2024, time.March, 4, 9, 0, 0, 0, time.UTC), // Monday
			want: time.Date(2024, time.March, 8, 13, 0, 0, 0, time.UTC),
		},
		{
			name: "on friday before 13h -> today at 13h",
			now:  time.Date(2024, time.March, 8, 8, 0, 0, 0, time.UTC), // Friday
			want: time.Date(2024, time.March, 8, 13, 0, 0, 0, time.UTC),
		},
		{
			name: "on friday after 13h -> today's date at 13h again (caller handles rollover)",
			now:  time.Date(2024, time.March, 8, 15, 0, 0, 0, time.UTC), // Friday afternoon
			want: time.Date(2024, time.March, 8, 13, 0, 0, 0, time.UTC),
		},
		{
			name: "saturday -> next friday",
			now:  time.Date(2024, time.March, 9, 10, 0, 0, 0, time.UTC), // Saturday
			want: time.Date(2024, time.March, 15, 13, 0, 0, 0, time.UTC),
		},
		{
			name: "sunday -> next friday",
			now:  time.Date(2024, time.March, 10, 10, 0, 0, 0, time.UTC), // Sunday
			want: time.Date(2024, time.March, 15, 13, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NextFridayAt13(tt.now)
			if !got.Equal(tt.want) {
				t.Errorf("NextFridayAt13(%v) = %v, want %v", tt.now, got, tt.want)
			}
		})
	}
}
