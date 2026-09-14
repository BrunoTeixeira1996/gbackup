package targets_test

import (
	"fmt"
	"testing"

	"github.com/BrunoTeixeira1996/gbackup/internal/commands"
	"github.com/BrunoTeixeira1996/gbackup/internal/config"
	"github.com/BrunoTeixeira1996/gbackup/internal/targets"
)

// leaving IP empty skips the ping check, which we can't rely on in tests

func TestExecuteTargetsBackups_AllSucceed(t *testing.T) {
	originalRsync := commands.RsyncCommand
	defer func() { commands.RsyncCommand = originalRsync }()

	var called []string
	commands.RsyncCommand = func(cmd, to, target, pushgatewayURL string) error {
		called = append(called, target)
		return nil
	}

	ts := []targets.Target{
		{
			Name:         "t1",
			ExternalPath: t.TempDir(),
			RsyncCommands: []config.RsyncCommand{
				{Name: "cmd1", Command: "-av a/ b/"},
			},
		},
		{
			Name:         "t2",
			ExternalPath: t.TempDir(),
			RsyncCommands: []config.RsyncCommand{
				{Name: "cmd2", Command: "-av c/ d/"},
			},
		},
	}

	results := targets.ExecuteTargetsBackups(ts, config.Config{})

	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
	for i, r := range results {
		if r.Err != nil {
			t.Errorf("results[%d].Err = %v, want nil", i, r.Err)
		}
	}
	if results[0].TargetName != "t1" || results[1].TargetName != "t2" {
		t.Errorf("results order/names = %q, %q, want t1, t2 (order preserved)", results[0].TargetName, results[1].TargetName)
	}
	if len(called) != 2 || called[0] != "cmd1" || called[1] != "cmd2" {
		t.Errorf("RsyncCommand called with targets %v, want [cmd1 cmd2]", called)
	}
}

func TestExecuteTargetsBackups_PartialFailure(t *testing.T) {
	originalRsync := commands.RsyncCommand
	defer func() { commands.RsyncCommand = originalRsync }()

	commands.RsyncCommand = func(cmd, to, target, pushgatewayURL string) error {
		if target == "failing" {
			return fmt.Errorf("simulated failure")
		}
		return nil
	}

	ts := []targets.Target{
		{
			Name:         "ok-target",
			ExternalPath: t.TempDir(),
			RsyncCommands: []config.RsyncCommand{
				{Name: "ok", Command: "-av a/ b/"},
			},
		},
		{
			Name:         "bad-target",
			ExternalPath: t.TempDir(),
			RsyncCommands: []config.RsyncCommand{
				{Name: "failing", Command: "-av a/ b/"},
			},
		},
	}

	results := targets.ExecuteTargetsBackups(ts, config.Config{})

	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
	if results[0].Err != nil {
		t.Errorf("results[0] (ok-target).Err = %v, want nil", results[0].Err)
	}
	if results[1].Err == nil {
		t.Error("results[1] (bad-target).Err = nil, want an error")
	}
}

func TestExecuteTargetsBackups_Empty(t *testing.T) {
	results := targets.ExecuteTargetsBackups(nil, config.Config{})
	if len(results) != 0 {
		t.Errorf("ExecuteTargetsBackups(nil, ...) = %+v, want empty slice", results)
	}
}
