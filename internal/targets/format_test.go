package targets

import "testing"

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name    string
		seconds float64
		want    string
	}{
		{"sub-second", 0.9, "0.9s"},
		{"a few seconds", 22.128, "22.1s"},
		{"just under a minute", 59.9, "59.9s"},
		{"exactly a minute", 60, "1m 0.0s"},
		{"minutes and seconds", 125.5, "2m 5.5s"},
		{"many minutes", 3661, "61m 1.0s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatDuration(tt.seconds); got != tt.want {
				t.Errorf("formatDuration(%v) = %q, want %q", tt.seconds, got, tt.want)
			}
		})
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		name string
		mb   float64
		want string
	}{
		{"small", 28.5, "28.50 MB"},
		{"just under 1GB", 1023.99, "1023.99 MB"},
		{"exactly 1GB", 1024, "1.00 GB"},
		{"several GB", 15189.16, "14.83 GB"},
		{"zero", 0, "0.00 MB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatSize(tt.mb); got != tt.want {
				t.Errorf("formatSize(%v) = %q, want %q", tt.mb, got, tt.want)
			}
		})
	}
}
