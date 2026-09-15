package commands_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/BrunoTeixeira1996/gbackup/internal/commands"
)

func TestRsyncCommand_ActuallyCopiesFiles(t *testing.T) {
	if _, err := exec.LookPath("rsync"); err != nil {
		t.Skip("rsync binary not found on PATH, skipping")
	}

	pg := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer pg.Close()

	srcDir := t.TempDir()
	dstDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(srcDir, "file.txt"), []byte("hello world"), 0644); err != nil {
		t.Fatalf("could not create source file: %v", err)
	}

	// RsyncCommand already execs "rsync" itself, so cmd is just the args
	cmd := "-a " + srcDir + "/ " + dstDir + "/"
	if err := commands.RsyncCommand(cmd, "toExternal", "testtarget", pg.URL); err != nil {
		t.Fatalf("RsyncCommand() returned unexpected error: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dstDir, "file.txt"))
	if err != nil {
		t.Fatalf("expected file.txt to be copied to destination: %v", err)
	}
	if string(got) != "hello world" {
		t.Errorf("copied file content = %q, want %q", got, "hello world")
	}
}

func TestRsyncCommand_PropagatesError(t *testing.T) {
	if _, err := exec.LookPath("rsync"); err != nil {
		t.Skip("rsync binary not found on PATH, skipping")
	}

	pg := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer pg.Close()

	// source dir doesn't exist, rsync exits non-zero, and that must surface
	// as a real error
	cmd := "-a /path/does/not/exist/hopefully/ " + t.TempDir() + "/"
	err := commands.RsyncCommand(cmd, "toExternal", "testtarget", pg.URL)
	if err == nil {
		t.Fatal("expected an error when the source directory doesn't exist, got nil")
	}
}
