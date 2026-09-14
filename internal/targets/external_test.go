package targets_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BrunoTeixeira1996/gbackup/internal/targets"
	"github.com/BrunoTeixeira1996/gbackup/internal/utils"
)

func TestVerifyExternalSize_Before(t *testing.T) {
	tmpDir := t.TempDir()
	// needs to be big enough to round to a nonzero MB value
	if err := os.WriteFile(filepath.Join(tmpDir, "file.txt"), make([]byte, 50*1024), 0644); err != nil {
		t.Fatalf("could not create test file: %v", err)
	}

	e := targets.External{ExternalPath: tmpDir}
	ts := &utils.TargetSize{}

	e.VerifyExternalSize("before", ts)

	if ts.Before <= 0 {
		t.Errorf("ts.Before = %v, want > 0", ts.Before)
	}
	if ts.After != 0 {
		t.Errorf("ts.After = %v, want unchanged (0)", ts.After)
	}
}

func TestVerifyExternalSize_After(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "file.txt"), make([]byte, 50*1024), 0644); err != nil {
		t.Fatalf("could not create test file: %v", err)
	}

	e := targets.External{ExternalPath: tmpDir}
	ts := &utils.TargetSize{Before: 42}

	e.VerifyExternalSize("after", ts)

	if ts.After <= 0 {
		t.Errorf("ts.After = %v, want > 0", ts.After)
	}
	if ts.Before != 42 {
		t.Errorf("ts.Before = %v, want unchanged (42)", ts.Before)
	}
}

func TestVerifyExternalSize_UnknownOperation_NoPanic(t *testing.T) {
	e := targets.External{ExternalPath: t.TempDir()}
	ts := &utils.TargetSize{}

	// unknown operation just logs and leaves ts untouched
	e.VerifyExternalSize("sideways", ts)

	if ts.Before != 0 || ts.After != 0 {
		t.Errorf("ts = %+v, want zero-value for unknown operation", ts)
	}
}
