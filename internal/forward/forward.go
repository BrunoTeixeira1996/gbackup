package forward

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// telegramBotURL is a var (rather than an inline literal) so tests can
// point it at a local httptest server instead of the real bot.
var telegramBotURL = "http://bot.lan:8000/forward"

// httpClient has an explicit timeout so an unresponsive telegram bot
// (hung, or silently dropping packets) can't block the caller forever -
// the default http.Client used by http.Post has no timeout at all.
var httpClient = &http.Client{Timeout: 10 * time.Second}

// Sends message from gbackup status to telegram bot
// then telegram bot will receive this communicate and it will
// display in the private chat
//
// Used like this (a var, not a plain func) to be able to mock this in
// tests (see internal/commands.RsyncCommand for the same pattern).
var ForwardMessageToTelegram = func(status string, messageContent string, messageErr string) error {

	m := fmt.Sprintf("%s - %v", status, messageContent)

	requestBody := map[string]string{
		"source":  "gbackup",
		"message": m,
		"error":   messageErr,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("[forward error] could not marshall JSON: %s\n", err)
	}

	resp, err := httpClient.Post(telegramBotURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("[forward error] could not make POST request: %s\n", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("[forward error] telegram bot returned status code: %d\n", resp.StatusCode)
	}

	return nil
}
