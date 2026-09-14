package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BrunoTeixeira1996/gbackup/internal/config"
	"github.com/BrunoTeixeira1996/gbackup/internal/utils"
)

const sampleToml = `
[nas]
name = "nas1"
ip = "192.168.1.50"
mac = "aa:bb:cc:dd:ee:ff"

[pushgateway]
url = "http://pushgateway.lan:9091"

[external]
external_path = "/mnt/external"

[[external.rsync_commands]]
name = "externalToNAS"
command = "rsync -a /mnt/external/ nas1:/backup/{current_time}/"

[[targets]]
name = "target1"
ip = "192.168.1.10"
mac = "11:22:33:44:55:66"
external_path = "/mnt/external/target1"

[[targets.rsync_commands]]
name = "cmd1"
command = "rsync -a source/ dest/"

[[proxmox.objects]]
id = "ct-101"
name = "jellyfin"

[[proxmox.objects]]
id = "vm-111"
name = "WS2022-DC"
`

func writeTempToml(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("could not write temp toml file: %v", err)
	}
	return path
}

func TestReadTomlFile_Valid(t *testing.T) {
	path := writeTempToml(t, sampleToml)

	cfg, err := config.ReadTomlFile(path)
	if err != nil {
		t.Fatalf("ReadTomlFile() returned unexpected error: %v", err)
	}

	if cfg.NAS.Name != "nas1" {
		t.Errorf("NAS.Name = %q, want %q", cfg.NAS.Name, "nas1")
	}
	if cfg.NAS.IP != "192.168.1.50" {
		t.Errorf("NAS.IP = %q, want %q", cfg.NAS.IP, "192.168.1.50")
	}
	if cfg.Pushgateway.Url != "http://pushgateway.lan:9091" {
		t.Errorf("Pushgateway.Url = %q, want %q", cfg.Pushgateway.Url, "http://pushgateway.lan:9091")
	}
	if len(cfg.Targets) != 1 {
		t.Fatalf("len(Targets) = %d, want 1", len(cfg.Targets))
	}
	if cfg.Targets[0].Name != "target1" {
		t.Errorf("Targets[0].Name = %q, want %q", cfg.Targets[0].Name, "target1")
	}
	if len(cfg.Targets[0].RsyncCommands) != 1 {
		t.Fatalf("len(Targets[0].RsyncCommands) = %d, want 1", len(cfg.Targets[0].RsyncCommands))
	}
	if len(cfg.Proxmox.Objects) != 2 {
		t.Fatalf("len(Proxmox.Objects) = %d, want 2", len(cfg.Proxmox.Objects))
	}
	if cfg.Proxmox.Objects[0].ID != "ct-101" || cfg.Proxmox.Objects[0].Name != "jellyfin" {
		t.Errorf("Proxmox.Objects[0] = %+v, want {ID: ct-101, Name: jellyfin}", cfg.Proxmox.Objects[0])
	}
	if cfg.Proxmox.Objects[1].ID != "vm-111" || cfg.Proxmox.Objects[1].Name != "WS2022-DC" {
		t.Errorf("Proxmox.Objects[1] = %+v, want {ID: vm-111, Name: WS2022-DC}", cfg.Proxmox.Objects[1])
	}
}

// {current_time} in external rsync commands gets replaced with today's date
func TestReadTomlFile_ReplacesCurrentTime(t *testing.T) {
	path := writeTempToml(t, sampleToml)

	cfg, err := config.ReadTomlFile(path)
	if err != nil {
		t.Fatalf("ReadTomlFile() returned unexpected error: %v", err)
	}

	if len(cfg.External.RsyncCommands) != 1 {
		t.Fatalf("len(External.RsyncCommands) = %d, want 1", len(cfg.External.RsyncCommands))
	}

	wantCommand := "rsync -a /mnt/external/ nas1:/backup/" + utils.CurrentTime() + "/"
	got := cfg.External.RsyncCommands[0].Command
	if got != wantCommand {
		t.Errorf("External.RsyncCommands[0].Command = %q, want %q", got, wantCommand)
	}
}

func TestReadTomlFile_FileDoesNotExist(t *testing.T) {
	_, err := config.ReadTomlFile("/path/does/not/exist/config.toml")
	if err == nil {
		t.Fatal("expected an error when the config file does not exist, got nil")
	}
}

func TestReadTomlFile_MalformedToml(t *testing.T) {
	path := writeTempToml(t, `this is not valid = toml = [[[`)

	_, err := config.ReadTomlFile(path)
	if err == nil {
		t.Fatal("expected an error when the config file is malformed toml, got nil")
	}
}

// parses the real config.toml this project actually runs with (not a
// synthetic fixture), so a syntax mistake or a broken section gets caught
// by `go test` instead of only being noticed at deploy time. Assertions are
// deliberately loose (no exact counts) since targets/proxmox objects are
// meant to be added/removed freely.
func TestRealConfigToml_Parses(t *testing.T) {
	cfg, err := config.ReadTomlFile("../../config.toml")
	if err != nil {
		t.Fatalf("the real config.toml failed to parse: %v", err)
	}

	if cfg.NAS.Name == "" {
		t.Error("expected NAS.Name to be set in the real config.toml")
	}
	if cfg.External.ExternalPath == "" {
		t.Error("expected External.ExternalPath to be set in the real config.toml")
	}
	if len(cfg.Targets) == 0 {
		t.Error("expected at least one target in the real config.toml")
	}
	if len(cfg.Proxmox.Objects) == 0 {
		t.Error("expected at least one proxmox object in the real config.toml")
	}
	for _, o := range cfg.Proxmox.Objects {
		if o.ID == "" {
			t.Errorf("proxmox object has an empty id: %+v", o)
		}
		if o.Name == "" {
			t.Errorf("proxmox object has an empty name: %+v", o)
		}
	}
}

func TestReadTomlFile_EmptyFile(t *testing.T) {
	path := writeTempToml(t, "")

	cfg, err := config.ReadTomlFile(path)
	if err != nil {
		t.Fatalf("ReadTomlFile() returned unexpected error for empty file: %v", err)
	}

	if cfg.NAS.Name != "" || len(cfg.Targets) != 0 {
		t.Errorf("expected zero-value Config for empty file, got %+v", cfg)
	}
}
