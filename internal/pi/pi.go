package pi

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Sensor struct {
	Data struct {
		Result []struct {
			Value []any `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

// Function that gets Pi temperature
func GetPiTemp() string {
	req, err := http.NewRequest("GET", "http://192.168.30.77:9090/api/v1/query?query=node_hwmon_temp_celsius{sensor='temp0'}", nil)
	if err != nil {
		return ""
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer response.Body.Close()

	var s Sensor
	if err := json.NewDecoder(response.Body).Decode(&s); err != nil {
		return ""
	}

	return fmt.Sprintf("%s", s.Data.Result[0].Value[1])
}
