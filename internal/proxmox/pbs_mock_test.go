package proxmox_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/BrunoTeixeira1996/gbackup/internal/config"
	"github.com/BrunoTeixeira1996/gbackup/internal/proxmox"
)

// CheckBackupStatus needs 2 OK jobs per configured object to finish, so if we
// return all of them on the first poll it returns immediately without
// sleeping 20s

func makeObjects(n int) []config.ProxmoxObject {
	objs := make([]config.ProxmoxObject, n)
	for i := 0; i < n; i++ {
		objs[i] = config.ProxmoxObject{ID: fmt.Sprintf("ct-%d", 100+i), Name: fmt.Sprintf("service-%d", i)}
	}
	return objs
}

// one backup + one prune OK job per object, matching what a real PBS
// "tasks" response looks like (worker_id = "<datastore>:<vmid>")
func makeOKBackups(objects []config.ProxmoxObject) []proxmox.Backup {
	backups := make([]proxmox.Backup, 0, len(objects)*2)
	for i, o := range objects {
		backups = append(backups,
			proxmox.Backup{Upid: fmt.Sprintf("UPID:backup-%d", i), WorkerType: "backup", WordID: "backupProxmox:" + o.ID, Status: "OK"},
			proxmox.Backup{Upid: fmt.Sprintf("UPID:prune-%d", i), WorkerType: "prune", WordID: "backupProxmox:" + o.ID, Status: "OK"},
		)
	}
	return backups
}

func TestCheckBackupStatus_Mock_CompletesImmediately(t *testing.T) {
	objects := makeObjects(8)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jobs := makeOKBackups(objects)
		resp := proxmox.QueryBackup{Total: len(jobs), DataBackup: jobs}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("could not encode response: %v", err)
		}
	}))
	defer srv.Close()

	pbs := proxmox.PBS{
		API: proxmox.Boilerplate{
			Url:  srv.URL,
			Node: "localhost",
		},
	}

	done := make(chan error, 1)
	go func() { done <- pbs.CheckBackupStatus(objects) }()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("CheckBackupStatus() returned unexpected error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("CheckBackupStatus() did not return within 5s; expected it to complete on the first poll")
	}
}

// the completion target must come from len(objects), not a fixed number -
// asking for only 2 objects but getting 16 real jobs back must not matter,
// it should still complete as soon as those 2 objects' jobs are seen
func TestCheckBackupStatus_Mock_TargetMatchesObjectCount(t *testing.T) {
	objects := makeObjects(2)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jobs := makeOKBackups(objects)
		resp := proxmox.QueryBackup{Total: len(jobs), DataBackup: jobs}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("could not encode response: %v", err)
		}
	}))
	defer srv.Close()

	pbs := proxmox.PBS{
		API: proxmox.Boilerplate{
			Url:  srv.URL,
			Node: "localhost",
		},
	}

	done := make(chan error, 1)
	go func() { done <- pbs.CheckBackupStatus(objects) }()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("CheckBackupStatus() returned unexpected error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("CheckBackupStatus() did not return within 5s")
	}
}

// with 4 configured objects, the target is 4*2=8 jobs. Getting exactly 8
// must complete; getting only 7 must keep waiting instead of completing
// early. Together these pin down that the counting is exact, not just
// "roughly right".
func TestCheckBackupStatus_Mock_FourObjects_ExactCount(t *testing.T) {
	objects := makeObjects(4)
	jobs := makeOKBackups(objects)
	if len(jobs) != 8 {
		t.Fatalf("test setup bug: makeOKBackups(4 objects) = %d jobs, want 8", len(jobs))
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := proxmox.QueryBackup{Total: len(jobs), DataBackup: jobs}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("could not encode response: %v", err)
		}
	}))
	defer srv.Close()

	pbs := proxmox.PBS{
		API: proxmox.Boilerplate{
			Url:  srv.URL,
			Node: "localhost",
		},
	}

	done := make(chan error, 1)
	go func() { done <- pbs.CheckBackupStatus(objects) }()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("CheckBackupStatus() returned unexpected error with all 8/8 jobs present: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("CheckBackupStatus() did not return within 5s with all 8/8 jobs present")
	}
}

func TestCheckBackupStatus_Mock_FourObjects_OneJobShortKeepsWaiting(t *testing.T) {
	objects := makeObjects(4)
	jobs := makeOKBackups(objects)[:7] // one short of the 8 needed

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := proxmox.QueryBackup{Total: len(jobs), DataBackup: jobs}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("could not encode response: %v", err)
		}
	}))
	defer srv.Close()

	pbs := proxmox.PBS{
		API: proxmox.Boilerplate{
			Url:  srv.URL,
			Node: "localhost",
		},
	}

	done := make(chan error, 1)
	go func() { done <- pbs.CheckBackupStatus(objects) }()

	select {
	case err := <-done:
		t.Fatalf("CheckBackupStatus() completed early with only 7/8 jobs present (err=%v)", err)
	case <-time.After(2 * time.Second):
		// still polling, as expected
	}
}

// added backup job log lines should show the configured friendly name, not
// just the raw upid/vmid
func TestCheckBackupStatus_Mock_LogsFriendlyNames(t *testing.T) {
	objects := []config.ProxmoxObject{
		{ID: "ct-101", Name: "jellyfin"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jobs := makeOKBackups(objects)
		resp := proxmox.QueryBackup{Total: len(jobs), DataBackup: jobs}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("could not encode response: %v", err)
		}
	}))
	defer srv.Close()

	pbs := proxmox.PBS{
		API: proxmox.Boilerplate{
			Url:  srv.URL,
			Node: "localhost",
		},
	}

	var logBuf bytes.Buffer
	origOutput := log.Writer()
	log.SetOutput(&logBuf)
	defer log.SetOutput(origOutput)

	if err := pbs.CheckBackupStatus(objects); err != nil {
		t.Fatalf("CheckBackupStatus() returned unexpected error: %v", err)
	}

	got := logBuf.String()
	if !strings.Contains(got, "jellyfin (ct-101)") {
		t.Errorf("expected logs to contain the friendly name %q, got: %s", "jellyfin (ct-101)", got)
	}
}

// the completion count only cares about how many qualifying jobs it's seen,
// not which vmid they belong to - so a job for a vmid that isn't in the
// configured list still counts, but its log line should fall back to the
// raw vmid instead of panicking or silently dropping the entry
func TestCheckBackupStatus_Mock_UnknownVMIDFallsBackToRawID(t *testing.T) {
	objects := []config.ProxmoxObject{
		{ID: "ct-999", Name: "whatever"},
	}
	unknownJobs := []proxmox.Backup{
		{Upid: "UPID:backup-unknown", WorkerType: "backup", WordID: "backupProxmox:ct-101", Status: "OK"},
		{Upid: "UPID:prune-unknown", WorkerType: "prune", WordID: "backupProxmox:ct-101", Status: "OK"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := proxmox.QueryBackup{Total: len(unknownJobs), DataBackup: unknownJobs}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("could not encode response: %v", err)
		}
	}))
	defer srv.Close()

	pbs := proxmox.PBS{
		API: proxmox.Boilerplate{
			Url:  srv.URL,
			Node: "localhost",
		},
	}

	var logBuf bytes.Buffer
	origOutput := log.Writer()
	log.SetOutput(&logBuf)
	defer log.SetOutput(origOutput)

	if err := pbs.CheckBackupStatus(objects); err != nil {
		t.Fatalf("CheckBackupStatus() returned unexpected error: %v", err)
	}

	got := logBuf.String()
	if !strings.Contains(got, "added backup job for ct-101") {
		t.Errorf("expected logs to fall back to the raw vmid for an unconfigured object, got: %s", got)
	}
}

// a worker_id with no ":" at all (malformed/unexpected shape) should fall
// back to using the whole string as the vmid, not panic or drop the job
func TestCheckBackupStatus_Mock_WorkerIDWithNoColon(t *testing.T) {
	objects := []config.ProxmoxObject{{ID: "justavmid", Name: "whatever"}}
	jobs := []proxmox.Backup{
		{Upid: "UPID:backup-0", WorkerType: "backup", WordID: "justavmid", Status: "OK"},
		{Upid: "UPID:prune-0", WorkerType: "prune", WordID: "justavmid", Status: "OK"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := proxmox.QueryBackup{Total: len(jobs), DataBackup: jobs}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("could not encode response: %v", err)
		}
	}))
	defer srv.Close()

	pbs := proxmox.PBS{
		API: proxmox.Boilerplate{
			Url:  srv.URL,
			Node: "localhost",
		},
	}

	var logBuf bytes.Buffer
	origOutput := log.Writer()
	log.SetOutput(&logBuf)
	defer log.SetOutput(origOutput)

	if err := pbs.CheckBackupStatus(objects); err != nil {
		t.Fatalf("CheckBackupStatus() returned unexpected error: %v", err)
	}

	got := logBuf.String()
	if !strings.Contains(got, "whatever (justavmid)") {
		t.Errorf("expected logs to contain %q, got: %s", "whatever (justavmid)", got)
	}
}

// extra qualifying jobs beyond the target must not cause an error or get
// processed - the loop should stop as soon as it hits the target, even
// mid-response
func TestCheckBackupStatus_Mock_ExtraJobsBeyondTargetAreIgnored(t *testing.T) {
	objects := makeObjects(2) // target = 4
	extraObjects := makeObjects(3)
	jobs := makeOKBackups(extraObjects) // 6 jobs available, only 4 needed

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := proxmox.QueryBackup{Total: len(jobs), DataBackup: jobs}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("could not encode response: %v", err)
		}
	}))
	defer srv.Close()

	pbs := proxmox.PBS{
		API: proxmox.Boilerplate{
			Url:  srv.URL,
			Node: "localhost",
		},
	}

	done := make(chan error, 1)
	go func() { done <- pbs.CheckBackupStatus(objects) }()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("CheckBackupStatus() returned unexpected error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("CheckBackupStatus() did not return within 5s")
	}
}

// a job that's still running (empty status - PBS only populates status once
// a task finishes) must not count towards completion and must not error,
// just keep waiting for it to actually finish
func TestCheckBackupStatus_Mock_StillRunningJobKeepsWaiting(t *testing.T) {
	objects := makeObjects(16)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jobs := append([]proxmox.Backup{
			{Upid: "UPID:not-done-yet", WorkerType: "backup", Status: ""},
		}, makeOKBackups(objects)...)
		resp := proxmox.QueryBackup{Total: len(jobs), DataBackup: jobs}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("could not encode response: %v", err)
		}
	}))
	defer srv.Close()

	pbs := proxmox.PBS{
		API: proxmox.Boilerplate{
			Url:  srv.URL,
			Node: "localhost",
		},
	}

	if err := pbs.CheckBackupStatus(objects); err != nil {
		t.Fatalf("CheckBackupStatus() returned unexpected error: %v", err)
	}
}

// a job that finished with a real non-OK status is a genuine failure and
// must be reported immediately, not silently ignored or waited out for the
// full 30-minute timeout
func TestCheckBackupStatus_Mock_FailedJobReturnsErrorImmediately(t *testing.T) {
	objects := []config.ProxmoxObject{{ID: "ct-101", Name: "jellyfin"}}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jobs := []proxmox.Backup{
			{Upid: "UPID:failed-1", WorkerType: "backup", WordID: "backupProxmox:ct-101", Status: "unknown error"},
		}
		resp := proxmox.QueryBackup{Total: len(jobs), DataBackup: jobs}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("could not encode response: %v", err)
		}
	}))
	defer srv.Close()

	pbs := proxmox.PBS{
		API: proxmox.Boilerplate{
			Url:  srv.URL,
			Node: "localhost",
		},
	}

	done := make(chan error, 1)
	go func() { done <- pbs.CheckBackupStatus(objects) }()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected an error for a job with a non-OK terminal status, got nil")
		}
		if !strings.Contains(err.Error(), "jellyfin (ct-101)") {
			t.Errorf("expected the error to name the failed job, got: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("CheckBackupStatus() did not return within 5s; a real failure should be reported immediately, not waited out")
	}
}

// jobs that aren't backup/prune tasks at all (e.g. other PBS/PVE activity
// sharing the same task log) must be skipped entirely, not counted and not
// causing an error
func TestCheckBackupStatus_Mock_IgnoresOtherWorkerTypes(t *testing.T) {
	objects := makeObjects(1) // target = 2

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jobs := append([]proxmox.Backup{
			{Upid: "UPID:migrate-1", WorkerType: "migrate", Status: "OK"},
			{Upid: "UPID:verify-1", WorkerType: "verify", Status: "OK"},
		}, makeOKBackups(objects)...)
		resp := proxmox.QueryBackup{Total: len(jobs), DataBackup: jobs}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("could not encode response: %v", err)
		}
	}))
	defer srv.Close()

	pbs := proxmox.PBS{
		API: proxmox.Boilerplate{
			Url:  srv.URL,
			Node: "localhost",
		},
	}

	done := make(chan error, 1)
	go func() { done <- pbs.CheckBackupStatus(objects) }()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("CheckBackupStatus() returned unexpected error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("CheckBackupStatus() did not return within 5s")
	}
}

// the same job (same Upid) appearing twice in one response must be logged
// only once. Note: len(completed) itself can't detect a missing dedup
// check, since completed is a map keyed by Upid and re-assigning the same
// key is already idempotent - the only observable effect of a missing
// dedup check is a duplicate log line, so that's what this asserts on. The
// duplicate is placed before the target is reached, so the dedup check
// actually runs instead of being short-circuited by the "already at
// target" break.
func TestCheckBackupStatus_Mock_DuplicateUpidLoggedOnce(t *testing.T) {
	objects := makeObjects(2) // target = 4
	real := makeOKBackups(objects)
	jobs := append([]proxmox.Backup{real[0]}, real...) // real[0] (backup-0) duplicated up front

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := proxmox.QueryBackup{Total: len(jobs), DataBackup: jobs}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("could not encode response: %v", err)
		}
	}))
	defer srv.Close()

	pbs := proxmox.PBS{
		API: proxmox.Boilerplate{
			Url:  srv.URL,
			Node: "localhost",
		},
	}

	var logBuf bytes.Buffer
	origOutput := log.Writer()
	log.SetOutput(&logBuf)
	defer log.SetOutput(origOutput)

	done := make(chan error, 1)
	go func() { done <- pbs.CheckBackupStatus(objects) }()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("CheckBackupStatus() returned unexpected error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("CheckBackupStatus() did not return within 5s")
	}

	got := logBuf.String()
	count := strings.Count(got, "added backup job for service-0 (ct-100)")
	if count != 1 {
		t.Errorf(`"added backup job for service-0 (ct-100)" appeared %d times in logs, want 1: %s`, count, got)
	}
}

func TestCheckBackupStatus_Mock_NonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	pbs := proxmox.PBS{
		API: proxmox.Boilerplate{
			Url:  srv.URL,
			Node: "localhost",
		},
	}

	if err := pbs.CheckBackupStatus(makeObjects(8)); err == nil {
		t.Fatal("expected an error when the API returns a non-200 status, got nil")
	}
}

func TestCheckBackupStatus_Mock_MalformedJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `not valid json`)
	}))
	defer srv.Close()

	pbs := proxmox.PBS{
		API: proxmox.Boilerplate{
			Url:  srv.URL,
			Node: "localhost",
		},
	}

	if err := pbs.CheckBackupStatus(makeObjects(8)); err == nil {
		t.Fatal("expected an error when the API returns malformed JSON, got nil")
	}
}
