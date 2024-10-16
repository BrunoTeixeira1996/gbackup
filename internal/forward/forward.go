package forward

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Message struct {
	Content interface{}
	Err     error
}

// Sends message from gbackup status to telegram bot
// then telegram bot will receive this communicate and it will
// display in the private chat
func ForwardMessageToTelegram(status string, message Message) error {

	m := fmt.Sprintf("%s - %v", status, message)

	requestBody := map[string]string{
		"type":    "gbackup",
		"message": m,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("[forward error] could not marshall JSON: %s\n", err)
	}

	// telegram bot IP
	resp, err := http.Post("http://192.168.30.21:8000/fwd", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("[forward error] could not make POST request: %s\n", err)
	}
	defer resp.Body.Close()

	return nil
}
