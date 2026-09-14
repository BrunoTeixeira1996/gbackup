package forward

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// var so tests can point it at a local server
var telegramBotURL = "http://bot.lan:8000/forward"

// timeout so a hung bot doesn't block forever
var httpClient = &http.Client{Timeout: 10 * time.Second}

// var so tests can force a marshal failure
var marshalJSON = json.Marshal

// Sends message from gbackup status to telegram bot
// then telegram bot will receive this communicate and it will
// display in the private chat
var ForwardMessageToTelegram = func(status string, messageContent string, messageErr string) error {

	m := fmt.Sprintf("%s - %v", status, messageContent)

	requestBody := map[string]string{
		"source":  "gbackup",
		"message": m,
		"error":   messageErr,
	}

	jsonData, err := marshalJSON(requestBody)
	if err != nil {
		return fmt.Errorf("[forward error] could not marshall JSON: %s\n", err)
	}

	resp, err := httpClient.Post(telegramBotURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("[forward error] could not make POST request: %s\n", err)
		return fmt.Errorf("[forward error] could not make POST request: %s\n", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[forward error] telegram bot returned status code: %d\n", resp.StatusCode)
		return fmt.Errorf("[forward error] telegram bot returned status code: %d\n", resp.StatusCode)
	}

	return nil
}
