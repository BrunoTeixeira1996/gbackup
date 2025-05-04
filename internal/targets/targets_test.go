package targets_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/BrunoTeixeira1996/gbackup/internal/commands"
	"github.com/BrunoTeixeira1996/gbackup/internal/config"
	"github.com/BrunoTeixeira1996/gbackup/internal/targets"
	"github.com/BrunoTeixeira1996/gbackup/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestExecuteBackup_MultipleCases(t *testing.T) {
	type testCase struct {
		name          string
		rsyncCommands []config.RsyncCommand
		mockErrors    map[string]error
		expectedErrs  int
	}

	tests := []testCase{
		/*{
			name: "No errors",
			rsyncCommands: []config.RsyncCommand{
				{Command: "echo success1", Name: "cmd1"},
				{Command: "echo success2", Name: "cmd2"},
			},
			mockErrors:   map[string]error{},
			expectedErrs: 0,
		},*/
		{
			name: "One error",
			rsyncCommands: []config.RsyncCommand{
				{Command: "echo success", Name: "cmd1"},
				{Command: "echo fail", Name: "cmd2"},
			},
			mockErrors: map[string]error{
				"cmd2": fmt.Errorf("simulated rsync failure for cmd2"),
			},
			expectedErrs: 1,
		},
		{
			name: "Multiple errors",
			rsyncCommands: []config.RsyncCommand{
				{Command: "echo fail1", Name: "cmd1"},
				{Command: "echo fail2", Name: "cmd2"},
			},
			mockErrors: map[string]error{
				"cmd1": fmt.Errorf("error1"),
				"cmd2": fmt.Errorf("error2"),
			},
			expectedErrs: 2,
		},
	}

	originalRsync := commands.RsyncCommand
	defer func() { commands.RsyncCommand = originalRsync }()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock RsyncCommand
			var called []string
			commands.RsyncCommand = func(cmd, to, target, pushgatewayURL string) error {
				called = append(called, target)
				if err, exists := tt.mockErrors[target]; exists {
					return err
				}
				return nil
			}

			// Temp folder
			tmpDir, err := os.MkdirTemp("", "external")
			assert.NoError(t, err)
			defer os.RemoveAll(tmpDir)

			err = os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("test"), 0644)
			assert.NoError(t, err)

			target := targets.Target{
				Name:          "TestTarget",
				ExternalPath:  tmpDir,
				RsyncCommands: tt.rsyncCommands,
			}

			elapsed := &utils.ElapsedTime{}
			size := &utils.TargetSize{}

			err = target.ExecuteBackup(config.Config{Pushgateway: config.Pushgateway{Url: "http://localhost"}}, elapsed, size)

			if tt.expectedErrs == 0 {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				for _, expected := range tt.mockErrors {
					assert.Contains(t, err.Error(), expected.Error())
				}
			}

			assert.Equal(t, "TestTarget", elapsed.Target)
			assert.Greater(t, elapsed.Value, 0.0)
			assert.GreaterOrEqual(t, size.Before, 0.0)
			assert.GreaterOrEqual(t, size.After, size.Before)
			assert.Equal(t, len(tt.rsyncCommands), len(called))
		})
	}
}
