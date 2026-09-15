package commands_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/BrunoTeixeira1996/gbackup/internal/commands"
)

// fake pushgateway that just accepts anything
func newPushgatewayStub(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestExecCmdToProm_Success(t *testing.T) {
	pg := newPushgatewayStub(t)

	err := commands.ExecCmdToProm("echo", []string{"hello"}, "toExternal", "testInstance", pg.URL)
	if err != nil {
		t.Fatalf("ExecCmdToProm() returned unexpected error: %v", err)
	}
}

func TestExecCmdToProm_CommandNotFound(t *testing.T) {
	pg := newPushgatewayStub(t)

	err := commands.ExecCmdToProm("this-binary-does-not-exist-hopefully", nil, "toExternal", "testInstance", pg.URL)
	if err == nil {
		t.Fatal("expected an error when the underlying command does not exist, got nil")
	}
}

func TestExecCmdToProm_NonZeroExitReturnsError(t *testing.T) {
	pg := newPushgatewayStub(t)

	err := commands.ExecCmdToProm("false", nil, "toExternal", "testInstance", pg.URL)
	if err == nil {
		t.Fatal("expected an error for a command that exits non-zero, got nil")
	}
}
