package handle_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/BrunoTeixeira1996/gbackup/internal/handle"
	"github.com/BrunoTeixeira1996/gbackup/internal/run"
)

func TestBackupHandle_RejectsNonPOST(t *testing.T) {
	d := &handle.Demand{Args: run.Args{}}

	req := httptest.NewRequest(http.MethodGet, "/backup", nil)
	w := httptest.NewRecorder()

	d.BackupHandle(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status code = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestBackupHandle_RejectsMalformedJSON(t *testing.T) {
	d := &handle.Demand{Args: run.Args{}}

	req := httptest.NewRequest(http.MethodPost, "/backup", strings.NewReader("not valid json"))
	w := httptest.NewRecorder()

	d.BackupHandle(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("status code = %d, want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestBackupHandle_Success(t *testing.T) {
	original := run.Run
	defer func() { run.Run = original }()

	var gotArgs run.Args
	run.Run = func(args run.Args) error {
		gotArgs = args
		return nil
	}

	wantArgs := run.Args{ConfigPathFlag: "/some/config.toml", DebugFlag: true}
	d := &handle.Demand{Args: wantArgs}

	req := httptest.NewRequest(http.MethodPost, "/backup", strings.NewReader(`{"operation":"full"}`))
	w := httptest.NewRecorder()

	d.BackupHandle(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status code = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Executed gbackup on demand") {
		t.Errorf("body = %q, want it to mention the backup ran", body)
	}
	if gotArgs.ConfigPathFlag != wantArgs.ConfigPathFlag || gotArgs.DebugFlag != wantArgs.DebugFlag {
		t.Errorf("run.Run called with %+v, want %+v (BackupHandle should pass through d.Args)", gotArgs, wantArgs)
	}
}

func TestBackupHandle_RunFails(t *testing.T) {
	original := run.Run
	defer func() { run.Run = original }()

	run.Run = func(args run.Args) error {
		return fmt.Errorf("backup exploded")
	}

	d := &handle.Demand{Args: run.Args{}}

	req := httptest.NewRequest(http.MethodPost, "/backup", strings.NewReader(`{"operation":"full"}`))
	w := httptest.NewRecorder()

	d.BackupHandle(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "backup exploded") {
		t.Errorf("body = %q, want it to contain the run.Run error", body)
	}
}
