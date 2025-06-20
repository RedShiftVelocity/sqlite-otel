package handlers

import (
	"encoding/json"
	"fmt"
	"sync"
	
	"github.com/RedShiftVelocity/sqlite-otel/logging"
)

// stdoutMutex protects concurrent writes to stdout
var stdoutMutex sync.Mutex

// TelemetryOutput represents the JSON structure for stdout output
type TelemetryOutput struct {
	Type string          `json:"type"`
	Body json.RawMessage `json:"body"`
}

func WriteTelemetryData(telemetryType string, body []byte) error {
	// Log telemetry activity
	logging.Debug("Received %s telemetry data, size: %d bytes", telemetryType, len(body))
	
	// Create structured output using proper JSON marshaling
	output := TelemetryOutput{
		Type: telemetryType,
		Body: json.RawMessage(body),
	}

	jsonData, err := json.Marshal(output)
	if err != nil {
		logging.Error("Failed to marshal telemetry output: %v", err)
		return fmt.Errorf("failed to marshal telemetry output: %w", err)
	}

	// Synchronize writes to stdout to prevent race conditions
	stdoutMutex.Lock()
	defer stdoutMutex.Unlock()
	if _, err := fmt.Println(string(jsonData)); err != nil {
		logging.Error("Failed to write telemetry data to stdout: %v", err)
		return fmt.Errorf("failed to write telemetry data to stdout: %w", err)
	}
	
	logging.Debug("Successfully processed %s telemetry data", telemetryType)
	return nil
}