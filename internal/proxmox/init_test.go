package proxmox_test

import (
	"testing"

	"github.com/BrunoTeixeira1996/gbackup/internal/proxmox"
)

func TestPBSInit_SetsFieldsFromEnv(t *testing.T) {
	t.Setenv("PBS_TOKENID", "test-token")
	t.Setenv("PBS_SECRET", "test-secret")

	var pbs proxmox.PBS
	if err := pbs.Init(); err != nil {
		t.Fatalf("PBS.Init() returned unexpected error: %v", err)
	}

	if pbs.API.TokenID != "test-token" {
		t.Errorf("API.TokenID = %q, want %q", pbs.API.TokenID, "test-token")
	}
	if pbs.API.Secret != "test-secret" {
		t.Errorf("API.Secret = %q, want %q", pbs.API.Secret, "test-secret")
	}
	if pbs.API.Url == "" {
		t.Error("API.Url is empty")
	}
	if pbs.API.Authorization == "" {
		t.Error("API.Authorization is empty")
	}
}

func TestPVEInit_SetsFieldsFromEnv(t *testing.T) {
	t.Setenv("PVE_TOKENID", "test-token")
	t.Setenv("PVE_SECRET", "test-secret")

	var pve proxmox.PVE
	if err := pve.Init(); err != nil {
		t.Fatalf("PVE.Init() returned unexpected error: %v", err)
	}

	if pve.API.TokenID != "test-token" {
		t.Errorf("API.TokenID = %q, want %q", pve.API.TokenID, "test-token")
	}
	if pve.API.Secret != "test-secret" {
		t.Errorf("API.Secret = %q, want %q", pve.API.Secret, "test-secret")
	}
	if pve.API.Url == "" {
		t.Error("API.Url is empty")
	}
	if pve.API.Authorization == "" {
		t.Error("API.Authorization is empty")
	}
}
