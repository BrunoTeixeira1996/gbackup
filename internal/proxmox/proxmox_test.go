package proxmox_test

import (
	"os"
	"testing"

	"github.com/BrunoTeixeira1996/gbackup/internal/proxmox"
)

func TestInit(t *testing.T) {
	tokenID := os.Getenv("PVE_TOKENID")
	secret := os.Getenv("PVE_SECRET")

	if tokenID == "" || secret == "" {
		t.Fatalf("PVE_TOKENID or PVE_SECRET not defined.")
	}

	var pve proxmox.PVE
	if err := pve.Init(); err != nil {
		t.Fatalf("Error in pve.Init() PVE: %s", err)
	}

	if pve.API.Url == "" {
		t.Error("API URL not used")
	}

	t.Logf("URL: %s", pve.API.Url)
}

func TestGetAllObjects(t *testing.T) {
	tokenID := os.Getenv("PVE_TOKENID")
	secret := os.Getenv("PVE_SECRET")

	if tokenID == "" || secret == "" {
		t.Fatalf("PVE_TOKENID or PVE_SECRET not defined.")
	}

	var pve proxmox.PVE
	if err := pve.Init(); err != nil {
		t.Fatalf("Couldn't init PVE: %s", err)
	}

	if err := pve.GetAllObjects(); err != nil {
		t.Fatalf("Couldn't obtain all objects: %s", err)
	}

	// Verifica se temos LXCs e VMs
	if len(pve.LXCs) == 0 {
		t.Error("No LXC were found")
	} else {
		t.Logf("Found %d LXC", len(pve.LXCs))
	}

	if len(pve.VMs) == 0 {
		t.Error("No VMs were found")
	} else {
		t.Logf("Found %d VMs", len(pve.VMs))
	}
}
