package forward

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// the timeout test below overrides httpClient.Timeout itself, so this
// checks the actual shipped default is sane on its own
func TestHTTPClientHasATimeout(t *testing.T) {
	if httpClient.Timeout <= 0 {
		t.Errorf("httpClient.Timeout = %v, want a positive timeout", httpClient.Timeout)
	}
}

func withTelegramBotURL(t *testing.T, url string) {
	t.Helper()
	original := telegramBotURL
	telegramBotURL = url
	t.Cleanup(func() { telegramBotURL = original })
}

func TestForwardMessageToTelegram_Success(t *testing.T) {
	var gotBody map[string]string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("could not decode request body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	withTelegramBotURL(t, srv.URL)

	err := ForwardMessageToTelegram("STATUS", "hello", "")
	if err != nil {
		t.Fatalf("ForwardMessageToTelegram() returned unexpected error: %v", err)
	}

	if gotBody["source"] != "gbackup" {
		t.Errorf("request body source = %q, want %q", gotBody["source"], "gbackup")
	}
	if gotBody["message"] != "STATUS - hello" {
		t.Errorf("request body message = %q, want %q", gotBody["message"], "STATUS - hello")
	}
}

func TestForwardMessageToTelegram_NonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()
	withTelegramBotURL(t, srv.URL)

	err := ForwardMessageToTelegram("STATUS", "hello", "")
	if err == nil {
		t.Fatal("expected an error when the telegram bot returns a non-200 status, got nil")
	}
}

func TestForwardMessageToTelegram_TimesOutInsteadOfHanging(t *testing.T) {
	block := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-block // never respond until the test unblocks it at cleanup
	}))
	// close(block) needs to run before srv.Close(), otherwise Close() hangs
	defer srv.Close()
	defer close(block)
	withTelegramBotURL(t, srv.URL)

	originalTimeout := httpClient.Timeout
	httpClient.Timeout = 200 * time.Millisecond
	t.Cleanup(func() { httpClient.Timeout = originalTimeout })

	done := make(chan error, 1)
	go func() { done <- ForwardMessageToTelegram("STATUS", "hello", "") }()

	select {
	case err := <-done:
		if err == nil {
			t.Error("expected a timeout error from an unresponsive bot, got nil")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("ForwardMessageToTelegram() did not return within 2s; it should have timed out after 200ms")
	}
}

func TestForwardMessageToTelegram_UnreachableHost(t *testing.T) {
	withTelegramBotURL(t, "http://127.0.0.1:1") // nothing listens here

	if err := ForwardMessageToTelegram("STATUS", "hello", ""); err == nil {
		t.Fatal("expected an error when the telegram bot host is unreachable, got nil")
	}
}
