package proxmox

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/BrunoTeixeira1996/gbackup/internal/config"
)

// white-box (package proxmox, not proxmox_test) so pbsAPIURL/pveAPIURL can
// be overridden - these back CheckPBSBackupStatus's own orchestration body,
// which every other test in this package bypasses by replacing the whole
// CheckPBSBackupStatus var with a mock.

func withPBSAndPVEURLs(t *testing.T, pbsURL, pveURL string) {
	t.Helper()
	origPBS, origPVE := pbsAPIURL, pveAPIURL
	pbsAPIURL, pveAPIURL = pbsURL, pveURL
	t.Cleanup(func() { pbsAPIURL, pveAPIURL = origPBS, origPVE })
}

func TestCheckPBSBackupStatus_Body_Success(t *testing.T) {
	objects := []config.ProxmoxObject{{ID: "ct-101", Name: "jellyfin"}}

	pbsSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := QueryBackup{
			DataBackup: []Backup{
				{Upid: "UPID:1", WorkerType: "backup", WordID: "backupProxmox:ct-101", Status: "OK"},
				{Upid: "UPID:2", WorkerType: "prune", WordID: "backupProxmox:ct-101", Status: "OK"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer pbsSrv.Close()

	pveSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"data":[]}`)
	}))
	defer pveSrv.Close()

	withPBSAndPVEURLs(t, pbsSrv.URL, pveSrv.URL)

	if err := CheckPBSBackupStatus(objects); err != nil {
		t.Fatalf("CheckPBSBackupStatus() returned unexpected error: %v", err)
	}
}

func TestCheckPBSBackupStatus_Body_PVEFails(t *testing.T) {
	pbsSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(QueryBackup{})
	}))
	defer pbsSrv.Close()

	pveSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer pveSrv.Close()

	withPBSAndPVEURLs(t, pbsSrv.URL, pveSrv.URL)

	if err := CheckPBSBackupStatus(nil); err == nil {
		t.Fatal("expected an error when the PVE call fails, got nil")
	}
}

func TestCheckPBSBackupStatus_Body_PBSFails(t *testing.T) {
	pbsSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer pbsSrv.Close()

	pveSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"data":[]}`)
	}))
	defer pveSrv.Close()

	withPBSAndPVEURLs(t, pbsSrv.URL, pveSrv.URL)

	if err := CheckPBSBackupStatus(nil); err == nil {
		t.Fatal("expected an error when the PBS call fails, got nil")
	}
}

// the 30-minute timeout, sped up: setting the max wait equal to the poll
// interval makes it trip on the very first iteration, no real sleep needed
func TestCheckBackupStatus_TimesOutWhenNeverComplete(t *testing.T) {
	origInterval, origMaxWait := pbsPollIntervalSeconds, pbsMaxWaitSeconds
	pbsPollIntervalSeconds = 1
	pbsMaxWaitSeconds = 1
	t.Cleanup(func() {
		pbsPollIntervalSeconds = origInterval
		pbsMaxWaitSeconds = origMaxWait
	})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// never returns enough jobs to complete
		json.NewEncoder(w).Encode(QueryBackup{})
	}))
	defer srv.Close()

	pbs := PBS{API: Boilerplate{Url: srv.URL, Node: "localhost"}}

	done := make(chan error, 1)
	go func() { done <- pbs.CheckBackupStatus([]config.ProxmoxObject{{ID: "ct-101"}}) }()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected a timeout error, got nil")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("CheckBackupStatus() did not time out within 5s")
	}
}
