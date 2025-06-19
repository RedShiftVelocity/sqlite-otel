package handlers

import (
	"encoding/json"
	"fmt"
)

func WriteTelemetryData(telemetryType string, body string) error {
	data := map[string]string{
		"type": telemetryType,
		"body": body,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal telemetry data: %w", err)
	}

	fmt.Println(string(jsonData))
	return nil
}