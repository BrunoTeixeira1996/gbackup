package run

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/BrunoTeixeira1996/gbackup/internal/config"
	"github.com/BrunoTeixeira1996/gbackup/internal/forward"
	"github.com/BrunoTeixeira1996/gbackup/internal/proxmox"
)

// utils.Body prints via both fmt.Printf (stdout) and log.Printf (stderr
// by default), so capture both to see the full output
func captureOutput(t *testing.T, fn func()) string {
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
		var buf strings.Builder
		io.Copy(&buf, r)
		done <- buf.String()
	}()

	fn()

	w.Close()
	return <-done
}

func TestCheckPBSBackupAndReport_Success(t *testing.T) {
	originalCheck := proxmox.CheckPBSBackupStatus
	originalForward := forward.ForwardMessageToTelegram
	defer func() {
		proxmox.CheckPBSBackupStatus = originalCheck
		forward.ForwardMessageToTelegram = originalForward
	}()

	proxmox.CheckPBSBackupStatus = func(objects []config.ProxmoxObject) error { return nil }

	var forwardCalled bool
	forward.ForwardMessageToTelegram = func(status, messageContent, messageErr string) error {
		forwardCalled = true
		return nil
	}

	output := captureOutput(t, func() { checkPBSBackupAndReport(nil) })

	if !strings.Contains(output, "[PBS] Backup OK") {
		t.Errorf("expected stdout to contain the PBS OK message on success, got: %q", output)
	}
	if forwardCalled {
		t.Error("expected ForwardMessageToTelegram not to be called on success, but it was")
	}
}

func TestCheckPBSBackupAndReport_Failure(t *testing.T) {
	originalCheck := proxmox.CheckPBSBackupStatus
	originalForward := forward.ForwardMessageToTelegram
	defer func() {
		proxmox.CheckPBSBackupStatus = originalCheck
		forward.ForwardMessageToTelegram = originalForward
	}()

	proxmox.CheckPBSBackupStatus = func(objects []config.ProxmoxObject) error { return fmt.Errorf("pbs is on fire") }

	var (
		forwardCalled bool
		gotErrMsg     string
	)
	forward.ForwardMessageToTelegram = func(status, messageContent, messageErr string) error {
		forwardCalled = true
		gotErrMsg = messageErr
		return nil
	}

	output := captureOutput(t, func() { checkPBSBackupAndReport(nil) })

	if strings.Contains(output, "[PBS] Backup OK") {
		t.Errorf("expected stdout NOT to contain the PBS OK message after a failed check, got: %q", output)
	}
	if !forwardCalled {
		t.Fatal("expected ForwardMessageToTelegram to be called on failure, but it wasn't")
	}
	if !strings.Contains(gotErrMsg, "pbs is on fire") {
		t.Errorf("expected the forwarded error to mention the underlying failure, got: %q", gotErrMsg)
	}
}

func TestCheckPBSBackupAndReport_PassesObjectsThrough(t *testing.T) {
	originalCheck := proxmox.CheckPBSBackupStatus
	defer func() { proxmox.CheckPBSBackupStatus = originalCheck }()

	want := []config.ProxmoxObject{{ID: "ct-101", Name: "jellyfin"}}
	var got []config.ProxmoxObject
	proxmox.CheckPBSBackupStatus = func(objects []config.ProxmoxObject) error {
		got = objects
		return nil
	}

	captureOutput(t, func() { checkPBSBackupAndReport(want) })

	if len(got) != 1 || got[0] != want[0] {
		t.Errorf("checkPBSBackupAndReport did not pass its objects through to CheckPBSBackupStatus, got: %+v", got)
	}
}
