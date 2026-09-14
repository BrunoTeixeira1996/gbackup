package handle

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/BrunoTeixeira1996/gbackup/internal/run"
)

// listenAndServe is unexported (white-box test, package handle not
// handle_test) so it can be overridden - the real ListenAndServe blocks
// forever on success and binds a real port, neither of which belongs in a
// unit test.
func TestStartWebHook_ListensOnPort8000(t *testing.T) {
	original := listenAndServe
	defer func() { listenAndServe = original }()

	var gotAddr string
	var gotHandler http.Handler
	listenAndServe = func(addr string, handler http.Handler) error {
		gotAddr = addr
		gotHandler = handler
		return nil
	}

	StartWebHook(run.Args{})

	if gotAddr != ":8000" {
		t.Errorf("listenAndServe called with addr = %q, want %q", gotAddr, ":8000")
	}
	if gotHandler == nil {
		t.Fatal("listenAndServe called with a nil handler")
	}
}

// confirms the mux StartWebHook builds actually routes POST /backup to
// BackupHandle, not just that *some* handler got passed
func TestStartWebHook_RoutesBackupPath(t *testing.T) {
	original := listenAndServe
	defer func() { listenAndServe = original }()

	var gotHandler http.Handler
	listenAndServe = func(addr string, handler http.Handler) error {
		gotHandler = handler
		return nil
	}

	StartWebHook(run.Args{})

	req := httptest.NewRequest(http.MethodGet, "/backup", nil)
	w := httptest.NewRecorder()
	gotHandler.ServeHTTP(w, req)

	// BackupHandle rejects non-POST with 400 - reaching that response at
	// all proves routing worked.
	if w.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("status code = %d, want %d (proves /backup routes to BackupHandle)", w.Result().StatusCode, http.StatusBadRequest)
	}
}

// calling StartWebHook more than once must not panic - it did when this
// used the shared http.DefaultServeMux via http.HandleFunc, since
// registering the same pattern twice panics
func TestStartWebHook_CallableMoreThanOnce(t *testing.T) {
	original := listenAndServe
	defer func() { listenAndServe = original }()
	listenAndServe = func(addr string, handler http.Handler) error { return nil }

	StartWebHook(run.Args{})
	StartWebHook(run.Args{}) // must not panic
}
