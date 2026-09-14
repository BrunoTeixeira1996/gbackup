package monitoring

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/BrunoTeixeira1996/gbackup/internal/targets"
	"github.com/BrunoTeixeira1996/gbackup/internal/utils"
)

func withPushgatewayURL(t *testing.T, url string) {
	t.Helper()
	original := pushgatewayURL
	pushgatewayURL = url
	t.Cleanup(func() { pushgatewayURL = original })
}

func captureMonitoringLog(t *testing.T, fn func()) string {
	t.Helper()
	var buf bytes.Buffer
	original := log.Writer()
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(original) })
	fn()
	return buf.String()
}

func TestSendFinalResultsToMonitoring_Success(t *testing.T) {
	var (
		gotMethod string
		gotPath   string
		gotBody   string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	withPushgatewayURL(t, srv.URL)

	results := []targets.BackupResult{
		{
			TargetName:  "t1",
			ElapsedTime: utils.ElapsedTime{Value: 12.5},
			TargetSize:  utils.TargetSize{Before: 100, After: 150},
		},
	}

	logs := captureMonitoringLog(t, func() { SendFinalResultsToMonitoring(results) })

	if gotMethod != http.MethodPut {
		t.Errorf("request method = %q, want %q", gotMethod, http.MethodPut)
	}
	if !strings.Contains(gotPath, "gbackup_metrics") {
		t.Errorf("request path = %q, want it to contain %q", gotPath, "gbackup_metrics")
	}
	if !strings.Contains(gotBody, "backup_size_before_mb") || !strings.Contains(gotBody, "t1") {
		t.Errorf("request body = %q, want it to contain the target's metrics", gotBody)
	}
	if !strings.Contains(logs, "pushed successfully") {
		t.Errorf("expected logs to report success, got: %s", logs)
	}
}

func TestSendFinalResultsToMonitoring_PushFails(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()
	withPushgatewayURL(t, srv.URL)

	logs := captureMonitoringLog(t, func() {
		SendFinalResultsToMonitoring([]targets.BackupResult{{TargetName: "t1"}})
	})

	if !strings.Contains(logs, "could not push metrics") {
		t.Errorf("expected logs to report a failure, got: %s", logs)
	}
}

func TestSendFinalResultsToMonitoring_EmptyResults(t *testing.T) {
	var requestReceived bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestReceived = true
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	withPushgatewayURL(t, srv.URL)

	logs := captureMonitoringLog(t, func() { SendFinalResultsToMonitoring(nil) })

	if !requestReceived {
		t.Error("expected a push request even with no results")
	}
	if !strings.Contains(logs, "pushed successfully") {
		t.Errorf("expected logs to report success, got: %s", logs)
	}
}
