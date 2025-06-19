package handlers

import (
	"encoding/json"
	"fmt"
	"sync"
)

// writeMutex protects concurrent writes to stdout
var writeMutex sync.Mutex

func WriteTelemetryData(telemetryType string, body string) error {
	data := map[string]string{
		"type": telemetryType,
		"body": body,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal telemetry data: %w", err)
	}

	// Synchronize writes to stdout to prevent race conditions
	writeMutex.Lock()
	fmt.Println(string(jsonData))
	writeMutex.Unlock()
	
	return nil
}