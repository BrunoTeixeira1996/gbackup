package utils

import (
	"bytes"
	"io"
	"log"
	"os"
	"strings"
	"testing"
)

func captureScreenOutput(t *testing.T, fn func()) string {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("could not create pipe: %v", err)
	}

	origStdout := os.Stdout
	origLogOutput := log.Writer()
	os.Stdout = w
	log.SetOutput(w)
	t.Cleanup(func() {
		os.Stdout = origStdout
		log.SetOutput(origLogOutput)
	})

	done := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		done <- buf.String()
	}()

	fn()

	w.Close()
	return <-done
}

func TestHeader(t *testing.T) {
	out := captureScreenOutput(t, Header)
	if !strings.Contains(out, "Starting Gbackup") {
		t.Errorf("expected output to contain %q, got: %s", "Starting Gbackup", out)
	}
}

func TestBody(t *testing.T) {
	out := captureScreenOutput(t, func() { Body("[TEST] OK") })
	if !strings.Contains(out, "[TEST] OK") {
		t.Errorf("expected output to contain %q, got: %s", "[TEST] OK", out)
	}
}

func TestFooter(t *testing.T) {
	out := captureScreenOutput(t, Footer)
	if !strings.Contains(out, "Ending Gbackup") {
		t.Errorf("expected output to contain %q, got: %s", "Ending Gbackup", out)
	}
}
