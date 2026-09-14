package handle_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/BrunoTeixeira1996/gbackup/internal/handle"
	"github.com/BrunoTeixeira1996/gbackup/internal/run"
)

// only testing the fast-fail branches here, a valid POST would call
// run.Run() which talks to the real telegram bot

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
