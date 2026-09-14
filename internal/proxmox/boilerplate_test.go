package proxmox

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// http.NewRequest rejects a method containing control characters
func TestBoilerplateRequest_InvalidMethod(t *testing.T) {
	b := Boilerplate{Url: "http://127.0.0.1:1", Node: "localhost"}

	if _, err := b.request("GET\n", "path"); err == nil {
		t.Fatal("expected an error for an invalid HTTP method, got nil")
	}
}

// a response whose declared Content-Length is longer than what's actually
// sent causes io.ReadAll to fail with an unexpected EOF
func TestBoilerplateRequest_BodyReadError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "1000")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("short"))
	}))
	defer srv.Close()

	b := Boilerplate{Url: srv.URL, Node: "localhost"}

	if _, err := b.request("GET", "path"); err == nil {
		t.Fatal("expected an error when the response body is truncated, got nil")
	}
}
