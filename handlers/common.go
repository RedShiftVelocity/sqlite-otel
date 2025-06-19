package handlers

import (
	"encoding/json"
	"fmt"
	"sync"
)

// stdoutMutex protects concurrent writes to stdout
var stdoutMutex sync.Mutex

// TelemetryOutput represents the JSON structure for stdout output
type TelemetryOutput struct {
	Type string          `json:"type"`
	Body json.RawMessage `json:"body"`
}

func WriteTelemetryData(telemetryType string, body []byte) error {
	// Create structured output using proper JSON marshaling
	output := TelemetryOutput{
		Type: telemetryType,
		Body: json.RawMessage(body),
	}

	jsonData, err := json.Marshal(output)
	if err != nil {
		return fmt.Errorf("failed to marshal telemetry output: %w", err)
	}

	// Synchronize writes to stdout to prevent race conditions
	stdoutMutex.Lock()
	defer stdoutMutex.Unlock()
	if _, err := fmt.Println(string(jsonData)); err != nil {
		return fmt.Errorf("failed to write telemetry data to stdout: %w", err)
	}
	
	return nil
}