package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/RedShiftVelocity/sqlite-otel/database"
)

func WriteTelemetryData(telemetryType string, body string) error {
	// Keep existing JSON output to stdout
	data := map[string]string{
		"type": telemetryType,
		"body": body,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal telemetry data: %w", err)
	}

	fmt.Println(string(jsonData))
	
	// Also write to SQLite database
	var parsedData map[string]interface{}
	if err := json.Unmarshal([]byte(body), &parsedData); err != nil {
		return fmt.Errorf("failed to parse telemetry data: %w", err)
	}

	switch telemetryType {
	case "trace":
		if err := database.InsertTraceData(parsedData); err != nil {
			return fmt.Errorf("failed to insert trace data: %w", err)
		}
	case "metrics":
		if err := database.InsertMetricsData(parsedData); err != nil {
			return fmt.Errorf("failed to insert metrics data: %w", err)
		}
	case "logs":
		if err := database.InsertLogsData(parsedData); err != nil {
			return fmt.Errorf("failed to insert logs data: %w", err)
		}
	}

	return nil
}