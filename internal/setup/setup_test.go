package setup

import (
	"os"
	"path/filepath"
	"testing"
)

func writeMountsFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "mounts")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("could not write fixture mounts file: %v", err)
	}
	return path
}

func TestIsMounted(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{
			name:    "mount point present",
			content: "/dev/sda1 / ext4 rw 0 0\n/dev/sdb1 /mnt/external ext4 rw 0 0\n",
			want:    true,
		},
		{
			name:    "mount point absent",
			content: "/dev/sda1 / ext4 rw 0 0\n",
			want:    false,
		},
		{
			name:    "empty file",
			content: "",
			want:    false,
		},
		{
			name:    "similarly-named but distinct mount point does not count",
			content: "/dev/sdb1 /mnt/external2 ext4 rw 0 0\n",
			want:    false,
		},
		{
			name:    "mount point present among several others",
			content: "/dev/sda1 / ext4 rw 0 0\n/dev/sdb1 /mnt/external2 ext4 rw 0 0\n/dev/sdc1 /mnt/external ext4 rw 0 0\n",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeMountsFile(t, tt.content)
			if got := isMounted(path, "/mnt/external"); got != tt.want {
				t.Errorf("isMounted(%q, /mnt/external) = %v, want %v", tt.content, got, tt.want)
			}
		})
	}
}

func TestIsMounted_MissingFile(t *testing.T) {
	if got := isMounted("/path/does/not/exist/mounts", "/mnt/external"); got != false {
		t.Errorf("isMounted() with missing file = %v, want false", got)
	}
}

func TestCheckEnvVars(t *testing.T) {
	allVars := []string{"PBS_SECRET", "PBS_TOKENID", "PVE_SECRET", "PVE_TOKENID"}

	t.Run("all set", func(t *testing.T) {
		for _, v := range allVars {
			t.Setenv(v, "value")
		}
		if !checkEnvVars() {
			t.Error("checkEnvVars() = false, want true when all env vars are set")
		}
	})

	for _, missing := range allVars {
		missing := missing
		t.Run("missing "+missing, func(t *testing.T) {
			for _, v := range allVars {
				if v == missing {
					t.Setenv(v, "")
				} else {
					t.Setenv(v, "value")
				}
			}
			if checkEnvVars() {
				t.Errorf("checkEnvVars() = true, want false when %s is empty", missing)
			}
		})
	}
}

func TestSetupToml(t *testing.T) {
	t.Run("valid file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.toml")
		content := "[nas]\nname = \"nas1\"\n"
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("could not write test config: %v", err)
		}

		cfg, err := setupToml(path)
		if err != nil {
			t.Fatalf("setupToml() returned unexpected error: %v", err)
		}
		if cfg.NAS.Name != "nas1" {
			t.Errorf("cfg.NAS.Name = %q, want %q", cfg.NAS.Name, "nas1")
		}
	})

	t.Run("missing file", func(t *testing.T) {
		if _, err := setupToml("/path/does/not/exist/config.toml"); err == nil {
			t.Error("setupToml() with missing file = nil error, want an error")
		}
	})
}

func TestIsEverythingConfigured_EmptyConfigPath(t *testing.T) {
	_, ok := IsEverythingConfigured("", false)
	if ok {
		t.Error("IsEverythingConfigured() with empty configPathFlag = true, want false")
	}
}

func TestIsEverythingConfigured_MissingEnvVars(t *testing.T) {
	for _, v := range []string{"PBS_SECRET", "PBS_TOKENID", "PVE_SECRET", "PVE_TOKENID"} {
		t.Setenv(v, "")
	}

	_, ok := IsEverythingConfigured("/some/config.toml", false)
	if ok {
		t.Error("IsEverythingConfigured() with missing env vars = true, want false")
	}
}
