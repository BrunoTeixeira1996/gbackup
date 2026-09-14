package targets

import (
	"fmt"
	"testing"
	"time"

	"github.com/BrunoTeixeira1996/gbackup/internal/commands"
	"github.com/BrunoTeixeira1996/gbackup/internal/config"
)

func withMockedExec(t *testing.T, ping func(ip string) ([]byte, error), ipNeighbor func(mac string) ([]byte, error)) {
	t.Helper()
	origPing, origIPNeighbor, origDelay := execPing, execIPNeighbor, isAliveRetryDelay
	if ping != nil {
		execPing = ping
	}
	if ipNeighbor != nil {
		execIPNeighbor = ipNeighbor
	}
	isAliveRetryDelay = time.Millisecond // keep retries fast
	t.Cleanup(func() {
		execPing = origPing
		execIPNeighbor = origIPNeighbor
		isAliveRetryDelay = origDelay
	})
}

func TestGetAssociatedIPFromMAC_Success(t *testing.T) {
	withMockedExec(t, nil, func(mac string) ([]byte, error) {
		return []byte("192.168.1.50 dev eth0 lladdr " + mac + " REACHABLE\n"), nil
	})

	target := &Target{MAC: "aa:bb:cc:dd:ee:ff"}
	ip, err := target.getAssociatedIPFromMAC()
	if err != nil {
		t.Fatalf("getAssociatedIPFromMAC() returned unexpected error: %v", err)
	}
	if ip != "192.168.1.50" {
		t.Errorf("ip = %q, want %q", ip, "192.168.1.50")
	}
}

func TestGetAssociatedIPFromMAC_CommandFails(t *testing.T) {
	withMockedExec(t, nil, func(mac string) ([]byte, error) {
		return nil, fmt.Errorf("no match")
	})

	target := &Target{MAC: "aa:bb:cc:dd:ee:ff"}
	if _, err := target.getAssociatedIPFromMAC(); err == nil {
		t.Fatal("expected an error when the underlying command fails, got nil")
	}
}

func TestIsAlive_RespondsImmediately(t *testing.T) {
	withMockedExec(t, func(ip string) ([]byte, error) {
		return []byte("2 packets transmitted, 2 received"), nil
	}, nil)

	target := &Target{IP: "10.0.0.1"}
	alive, err := target.isAlive()
	if err != nil {
		t.Fatalf("isAlive() returned unexpected error: %v", err)
	}
	if !alive {
		t.Error("isAlive() = false, want true")
	}
}

func TestIsAlive_DestinationUnreachable(t *testing.T) {
	withMockedExec(t, func(ip string) ([]byte, error) {
		return []byte("From 10.0.0.1 icmp_seq=1 Destination Host Unreachable"), nil
	}, nil)

	target := &Target{IP: "10.0.0.1"}
	alive, err := target.isAlive()
	if err != nil {
		t.Fatalf("isAlive() returned unexpected error: %v", err)
	}
	if alive {
		t.Error("isAlive() = true, want false for an unreachable host")
	}
}

func TestIsAlive_AllRetriesFail(t *testing.T) {
	var calls int
	withMockedExec(t, func(ip string) ([]byte, error) {
		calls++
		return nil, fmt.Errorf("ping: unknown host")
	}, nil)

	target := &Target{IP: "10.0.0.1"}
	alive, err := target.isAlive()
	if err == nil {
		t.Fatal("expected an error when all ping attempts fail, got nil")
	}
	if alive {
		t.Error("isAlive() = true, want false")
	}
	if calls != 3 { // maxRetries=2 means 3 total attempts (i=0,1,2)
		t.Errorf("ping was called %d times, want 3", calls)
	}
}

func TestIsAlive_RetriesThenSucceeds(t *testing.T) {
	var calls int
	withMockedExec(t, func(ip string) ([]byte, error) {
		calls++
		if calls < 2 {
			return nil, fmt.Errorf("transient failure")
		}
		return []byte("2 packets transmitted, 2 received"), nil
	}, nil)

	target := &Target{IP: "10.0.0.1"}
	alive, err := target.isAlive()
	if err != nil {
		t.Fatalf("isAlive() returned unexpected error: %v", err)
	}
	if !alive {
		t.Error("isAlive() = false, want true after a successful retry")
	}
}

func TestIsAlive_ResolvesIPFromMACWhenIPEmpty(t *testing.T) {
	withMockedExec(t,
		func(ip string) ([]byte, error) {
			if ip != "10.0.0.99" {
				t.Errorf("ping called with ip = %q, want %q (resolved from MAC)", ip, "10.0.0.99")
			}
			return []byte("2 packets transmitted, 2 received"), nil
		},
		func(mac string) ([]byte, error) {
			return []byte("10.0.0.99 dev eth0 lladdr " + mac + " REACHABLE\n"), nil
		},
	)

	target := &Target{MAC: "aa:bb:cc:dd:ee:ff"}
	alive, err := target.isAlive()
	if err != nil {
		t.Fatalf("isAlive() returned unexpected error: %v", err)
	}
	if !alive {
		t.Error("isAlive() = false, want true")
	}
	if target.IP != "10.0.0.99" {
		t.Errorf("target.IP = %q, want it to be populated from the MAC lookup", target.IP)
	}
}

func TestIsAlive_MACLookupFails(t *testing.T) {
	withMockedExec(t, nil, func(mac string) ([]byte, error) {
		return nil, fmt.Errorf("no match")
	})

	target := &Target{MAC: "aa:bb:cc:dd:ee:ff"}
	if _, err := target.isAlive(); err == nil {
		t.Fatal("expected an error when the MAC lookup fails, got nil")
	}
}

// the tests above always replace execPing/execIPNeighbor with a mock, so
// their real bodies (the actual exec.Command calls) never run - these
// exercise the real implementations directly, using safe, always-available
// commands
func TestExecPing_RealCommand(t *testing.T) {
	out, err := execPing("127.0.0.1")
	if err != nil {
		t.Fatalf("execPing(127.0.0.1) returned unexpected error: %v", err)
	}
	if len(out) == 0 {
		t.Error("expected some ping output, got none")
	}
}

func TestExecIPNeighbor_RealCommand(t *testing.T) {
	// "ip neighbor | grep nomatch..." legitimately finds nothing, so grep
	// exits 1 and the command returns an error - this just confirms the
	// real body runs without panicking, not any particular output.
	_, err := execIPNeighbor("no-such-mac-should-match-anything")
	_ = err
}

// ExecuteTargetsBackups only calls isAlive when a target has an IP set -
// these exercise that path directly, now that ping is mockable.

func TestExecuteTargetsBackups_AliveTargetGetsBackedUp(t *testing.T) {
	withMockedExec(t, func(ip string) ([]byte, error) {
		return []byte("2 packets transmitted, 2 received"), nil
	}, nil)

	originalRsync := commands.RsyncCommand
	defer func() { commands.RsyncCommand = originalRsync }()
	var rsyncCalled bool
	commands.RsyncCommand = func(cmd, to, target, pushgatewayURL string) error {
		rsyncCalled = true
		return nil
	}

	ts := []Target{
		{
			Name:         "alive-target",
			IP:           "10.0.0.1",
			ExternalPath: t.TempDir(),
			RsyncCommands: []config.RsyncCommand{
				{Name: "cmd1", Command: "-av a/ b/"},
			},
		},
	}

	results := ExecuteTargetsBackups(ts, config.Config{})
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	if results[0].Err != nil {
		t.Errorf("results[0].Err = %v, want nil", results[0].Err)
	}
	if !rsyncCalled {
		t.Error("expected RsyncCommand to be called for an alive target, but it wasn't")
	}
}

// a target whose isAlive() check genuinely errors (not just a clean
// "unreachable" detection) must still log the error and skip the backup
func TestExecuteTargetsBackups_IsAliveErrorsAndSkipsBackup(t *testing.T) {
	withMockedExec(t, func(ip string) ([]byte, error) {
		return nil, fmt.Errorf("ping: unknown host")
	}, nil)

	originalRsync := commands.RsyncCommand
	defer func() { commands.RsyncCommand = originalRsync }()
	var rsyncCalled bool
	commands.RsyncCommand = func(cmd, to, target, pushgatewayURL string) error {
		rsyncCalled = true
		return nil
	}

	ts := []Target{
		{
			Name:         "erroring-target",
			IP:           "10.0.0.1",
			ExternalPath: t.TempDir(),
			RsyncCommands: []config.RsyncCommand{
				{Name: "cmd1", Command: "-av a/ b/"},
			},
		},
	}

	results := ExecuteTargetsBackups(ts, config.Config{})
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	if results[0].Err == nil {
		t.Error("results[0].Err = nil, want the isAlive error to be recorded")
	}
	if rsyncCalled {
		t.Error("expected RsyncCommand NOT to be called when isAlive() errors, but it was")
	}
}

func TestExecuteTargetsBackups_DeadTargetSkipsBackup(t *testing.T) {
	withMockedExec(t, func(ip string) ([]byte, error) {
		return []byte("From 10.0.0.1 icmp_seq=1 Destination Host Unreachable"), nil
	}, nil)

	originalRsync := commands.RsyncCommand
	defer func() { commands.RsyncCommand = originalRsync }()
	var rsyncCalled bool
	commands.RsyncCommand = func(cmd, to, target, pushgatewayURL string) error {
		rsyncCalled = true
		return nil
	}

	ts := []Target{
		{
			Name:         "dead-target",
			IP:           "10.0.0.1",
			ExternalPath: t.TempDir(),
			RsyncCommands: []config.RsyncCommand{
				{Name: "cmd1", Command: "-av a/ b/"},
			},
		},
	}

	results := ExecuteTargetsBackups(ts, config.Config{})
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	if rsyncCalled {
		t.Error("expected RsyncCommand NOT to be called for a dead target, but it was")
	}
}
