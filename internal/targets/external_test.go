package targets_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/BrunoTeixeira1996/gbackup/internal/commands"
	"github.com/BrunoTeixeira1996/gbackup/internal/config"
	"github.com/BrunoTeixeira1996/gbackup/internal/nas"
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

func TestVerifyExternalSize_Before_MissingFolderLogsAndLeavesZero(t *testing.T) {
	e := targets.External{ExternalPath: "/path/does/not/exist/hopefully"}
	ts := &utils.TargetSize{}

	e.VerifyExternalSize("before", ts)

	if ts.Before != 0 {
		t.Errorf("ts.Before = %v, want 0 when the folder doesn't exist", ts.Before)
	}
}

func TestVerifyExternalSize_After_MissingFolderLogsAndLeavesZero(t *testing.T) {
	e := targets.External{ExternalPath: "/path/does/not/exist/hopefully"}
	ts := &utils.TargetSize{Before: 42}

	e.VerifyExternalSize("after", ts)

	if ts.After != 0 {
		t.Errorf("ts.After = %v, want 0 when the folder doesn't exist", ts.After)
	}
	if ts.Before != 42 {
		t.Errorf("ts.Before = %v, want unchanged (42)", ts.Before)
	}
}

func TestExecuteExternalToNASBackup_Success(t *testing.T) {
	originalRsync := commands.RsyncCommand
	originalKeepLastTwo := nas.KeepLastTwo
	defer func() {
		commands.RsyncCommand = originalRsync
		nas.KeepLastTwo = originalKeepLastTwo
	}()

	var rsyncCalled, keepLastTwoCalled bool
	commands.RsyncCommand = func(cmd, to, target, pushgatewayURL string) error {
		rsyncCalled = true
		if to != "toNAS" {
			t.Errorf("RsyncCommand called with to = %q, want %q", to, "toNAS")
		}
		return nil
	}
	nas.KeepLastTwo = func() error {
		keepLastTwoCalled = true
		return nil
	}

	external := targets.External{
		RsyncCommands: []config.RsyncCommand{{Name: "cmd1", Command: "-av a/ b/"}},
	}

	if err := targets.ExecuteExternalToNASBackup(external, config.Config{}); err != nil {
		t.Fatalf("ExecuteExternalToNASBackup() returned unexpected error: %v", err)
	}
	if !rsyncCalled {
		t.Error("expected RsyncCommand to be called, but it wasn't")
	}
	if !keepLastTwoCalled {
		t.Error("expected nas.KeepLastTwo to be called, but it wasn't")
	}
}

func TestExecuteExternalToNASBackup_RsyncFails(t *testing.T) {
	originalRsync := commands.RsyncCommand
	originalKeepLastTwo := nas.KeepLastTwo
	defer func() {
		commands.RsyncCommand = originalRsync
		nas.KeepLastTwo = originalKeepLastTwo
	}()

	commands.RsyncCommand = func(cmd, to, target, pushgatewayURL string) error {
		return fmt.Errorf("rsync exploded")
	}
	var keepLastTwoCalled bool
	nas.KeepLastTwo = func() error {
		keepLastTwoCalled = true
		return nil
	}

	external := targets.External{
		RsyncCommands: []config.RsyncCommand{{Name: "cmd1", Command: "-av a/ b/"}},
	}

	if err := targets.ExecuteExternalToNASBackup(external, config.Config{}); err == nil {
		t.Fatal("expected an error when RsyncCommand fails, got nil")
	}
	if keepLastTwoCalled {
		t.Error("expected nas.KeepLastTwo NOT to be called when the rsync step fails, but it was")
	}
}

func TestExecuteExternalToNASBackup_KeepLastTwoFails(t *testing.T) {
	originalRsync := commands.RsyncCommand
	originalKeepLastTwo := nas.KeepLastTwo
	defer func() {
		commands.RsyncCommand = originalRsync
		nas.KeepLastTwo = originalKeepLastTwo
	}()

	commands.RsyncCommand = func(cmd, to, target, pushgatewayURL string) error { return nil }
	nas.KeepLastTwo = func() error { return fmt.Errorf("ssh exploded") }

	external := targets.External{
		RsyncCommands: []config.RsyncCommand{{Name: "cmd1", Command: "-av a/ b/"}},
	}

	if err := targets.ExecuteExternalToNASBackup(external, config.Config{}); err == nil {
		t.Fatal("expected an error when nas.KeepLastTwo fails, got nil")
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
