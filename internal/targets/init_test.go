package targets_test

import (
	"testing"

	"github.com/BrunoTeixeira1996/gbackup/internal/config"
	"github.com/BrunoTeixeira1996/gbackup/internal/targets"
)

func TestInitTargets(t *testing.T) {
	cfg := config.Config{
		Targets: []config.Target{
			{
				Name:         "t1",
				IP:           "10.0.0.1",
				MAC:          "aa:bb:cc:dd:ee:ff",
				ExternalPath: "/mnt/external/t1",
				RsyncCommands: []config.RsyncCommand{
					{Name: "cmd1", Command: "-av src/ dst/"},
				},
			},
			{
				Name: "t2",
			},
		},
	}

	got := targets.InitTargets(cfg)

	if len(got) != 2 {
		t.Fatalf("len(InitTargets()) = %d, want 2", len(got))
	}

	if got[0].Name != "t1" || got[0].IP != "10.0.0.1" ||
		got[0].MAC != "aa:bb:cc:dd:ee:ff" || got[0].ExternalPath != "/mnt/external/t1" {
		t.Errorf("InitTargets()[0] = %+v, did not match expected fields from config.Target", got[0])
	}
	if len(got[0].RsyncCommands) != 1 || got[0].RsyncCommands[0].Name != "cmd1" {
		t.Errorf("InitTargets()[0].RsyncCommands = %+v, want 1 entry named cmd1", got[0].RsyncCommands)
	}

	if got[1].Name != "t2" {
		t.Errorf("InitTargets()[1].Name = %q, want %q", got[1].Name, "t2")
	}
}

func TestInitTargets_Empty(t *testing.T) {
	got := targets.InitTargets(config.Config{})
	if len(got) != 0 {
		t.Errorf("InitTargets() on empty config = %+v, want empty slice", got)
	}
}

func TestInitExternal(t *testing.T) {
	cfg := config.Config{
		External: config.External{
			ExternalPath: "/mnt/external",
			RsyncCommands: []config.RsyncCommand{
				{Name: "cmd1", Command: "-av /mnt/external/ nas1:/backup/"},
			},
		},
	}

	got := targets.InitExternal(cfg)

	if got.ExternalPath != "/mnt/external" {
		t.Errorf("InitExternal().ExternalPath = %q, want %q", got.ExternalPath, "/mnt/external")
	}
	if len(got.RsyncCommands) != 1 || got.RsyncCommands[0].Name != "cmd1" {
		t.Errorf("InitExternal().RsyncCommands = %+v, want 1 entry named cmd1", got.RsyncCommands)
	}
}
