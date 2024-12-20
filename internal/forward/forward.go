package forward

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// Sends message from gbackup status to telegram bot
// then telegram bot will receive this communicate and it will
// display in the private chat
func ForwardMessageToTelegram(status string, messageContent string, messageErr string) error {

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

	// telegram bot IP
	resp, err := http.Post("http://192.168.30.21:8000/forward", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("[forward error] could not make POST request: %s\n", err)
	}
	defer resp.Body.Close()

	return nil
}
