package proxmox_test

import (
	"os"
	"testing"
	"time"

	"github.com/BrunoTeixeira1996/gbackup/internal/proxmox"
)

// TestPBSInit checks if the PBS struct initializes correctly
func TestPBSInit(t *testing.T) {
	tokenID := os.Getenv("PBS_TOKENID")
	secret := os.Getenv("PBS_SECRET")

	if tokenID == "" || secret == "" {
		t.Skip("PBS_TOKENID or PBS_SECRET environment variables are not set. Skipping test.")
	}

	var pbs proxmox.PBS
	if err := pbs.Init(); err != nil {
		t.Fatalf("Failed to initialize PBS: %s", err)
	}

	if pbs.API.TokenID != tokenID {
		t.Errorf("PBS TokenID mismatch. Expected %s, got %s", tokenID, pbs.API.TokenID)
	}

	if pbs.API.Secret != secret {
		t.Errorf("PBS Secret mismatch. Expected %s, got %s", secret, pbs.API.Secret)
	}

	if pbs.API.Url == "" {
		t.Error("PBS API URL not initialized")
	}
	t.Log("URL:", pbs.API.Url)
}

// TestCheckBackupStatus checks the checkBackupStatus function.
// This is an integration test, so it will connect to the real PBS API.
func TestCheckBackupStatus(t *testing.T) {
	tokenID := os.Getenv("PBS_TOKENID")
	secret := os.Getenv("PBS_SECRET")

	if tokenID == "" || secret == "" {
		t.Skip("PBS_TOKENID or PBS_SECRET environment variables are not set. Skipping test.")
	}

	var pbs proxmox.PBS
	if err := pbs.Init(); err != nil {
		t.Fatalf("Failed to initialize PBS: %s", err)
	}

	// Set a small totalObjects count for testing
	totalObjects := 1

	// Set a timeout to prevent infinite loops if PBS API is unresponsive
	done := make(chan error, 1)
	go func() {
		done <- pbs.CheckBackupStatus(totalObjects)
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("checkBackupStatus failed: %s", err)
		}
	case <-time.After(1 * time.Minute):
		t.Fatal("checkBackupStatus timed out after 1 minute")
	}
}
