package targets_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/BrunoTeixeira1996/gbackup/internal/targets"
	"github.com/BrunoTeixeira1996/gbackup/internal/utils"
)

func TestReturnFinalResultsFormatted(t *testing.T) {
	results := []targets.BackupResult{
		{
			TargetName:  "t1",
			ElapsedTime: utils.ElapsedTime{Target: "t1", Value: 12.5},
			TargetSize:  utils.TargetSize{Before: 100, After: 150},
			Err:         nil,
		},
		{
			TargetName:  "t2",
			ElapsedTime: utils.ElapsedTime{Target: "t2", Value: 3.25},
			TargetSize:  utils.TargetSize{Before: 10, After: 10},
			Err:         errors.New("boom"),
		},
	}

	got := targets.ReturnFinalResultsFormatted(results, 3661) // 1h 1m 1s

	if !strings.HasPrefix(got, "```\n") || !strings.HasSuffix(got, "```") {
		t.Errorf("expected output wrapped in code fences, got: %q", got)
	}
	if !strings.Contains(got, "TargetName: t1") || !strings.Contains(got, "TargetName: t2") {
		t.Errorf("expected both target names present, got: %q", got)
	}
	if !strings.Contains(got, "Error: boom") {
		t.Errorf("expected error message for t2 present, got: %q", got)
	}
	if !strings.Contains(got, "Total backup time: 01:01:01 (hh:mm:ss)") {
		t.Errorf("expected formatted total time, got: %q", got)
	}
}

func TestReturnFinalResultsFormatted_TimeFormatting(t *testing.T) {
	tests := []struct {
		name       string
		totalTime  float64
		wantSubstr string
	}{
		{"zero seconds", 0, "Total backup time: 00:00:00 (hh:mm:ss)"},
		{"under a minute", 45, "Total backup time: 00:00:45 (hh:mm:ss)"},
		{"exact hour", 3600, "Total backup time: 01:00:00 (hh:mm:ss)"},
		{"multi hour", 7325, "Total backup time: 02:02:05 (hh:mm:ss)"},
		{"59 minutes", 3540, "Total backup time: 00:59:00 (hh:mm:ss)"},
		{"59 minutes 59 seconds", 3599, "Total backup time: 00:59:59 (hh:mm:ss)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := targets.ReturnFinalResultsFormatted(nil, tt.totalTime)
			if !strings.Contains(got, tt.wantSubstr) {
				t.Errorf("ReturnFinalResultsFormatted(nil, %v) = %q, want substring %q", tt.totalTime, got, tt.wantSubstr)
			}
		})
	}
}

func TestDisplayFinalResults_NoPanic(t *testing.T) {
	// just logs, checking it doesn't panic
	targets.DisplayFinalResults(nil)
	targets.DisplayFinalResults([]targets.BackupResult{
		{TargetName: "t1", Err: errors.New("boom")},
	})
}

func TestValidateBackupResultErrors_NoPanic(t *testing.T) {
	// just logs, checking it doesn't panic
	targets.ValidateBackupResultErrors(nil)
	targets.ValidateBackupResultErrors([]targets.BackupResult{
		{TargetName: "t1", Err: nil},
		{TargetName: "t2", Err: errors.New("boom")},
	})
}
