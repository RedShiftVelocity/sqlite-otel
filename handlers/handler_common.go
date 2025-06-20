package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	
	"github.com/RedShiftVelocity/sqlite-otel/logging"
)

// ProcessTelemetryRequest handles common logic for all telemetry endpoints
func ProcessTelemetryRequest(w http.ResponseWriter, r *http.Request, telemetryType string, insertFunc func(data map[string]interface{}) error) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check Content-Type header (support prefix matching for charset)
	contentType := r.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		logging.Debug("Unsupported Content-Type for %s: %s", telemetryType, contentType)
		http.Error(w, "Only application/json Content-Type is supported", http.StatusUnsupportedMediaType)
		return
	}

	// Enforce request body size limit to prevent DoS attacks
	const maxBodySize = 10 * 1024 * 1024 // 10 MB limit
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	defer r.Body.Close()

	// Parse JSON body directly from stream to avoid memory allocation
	var telemetryData map[string]interface{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&telemetryData); err != nil {
		if err == io.EOF {
			logging.Error("Empty request body for %s", telemetryType)
			http.Error(w, "Request body cannot be empty", http.StatusBadRequest)
			return
		}
		logging.Error("Error parsing %s JSON: %v", telemetryType, err)
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Store telemetry data in database (SQLite only storage)
	if err := insertFunc(telemetryData); err != nil {
		logging.Error("Error storing %s in database: %v", telemetryType, err)
		// Return 500 Internal Server Error as per OTLP/HTTP spec
		http.Error(w, fmt.Sprintf("Failed to process %s data", telemetryType), http.StatusInternalServerError)
		return
	}

	// Log request details (execution logging only, no telemetry data)
	// Note: Content-Length may be -1 if not specified by client
	contentLength := r.ContentLength
	if contentLength > 0 {
		logging.Debug("Received %s telemetry data, size: %d bytes", telemetryType, contentLength)
	} else {
		logging.Debug("Received %s telemetry data", telemetryType)
	}
	logging.Info("Stored %s data in SQLite - Content-Type: %s", 
		telemetryType, r.Header.Get("Content-Type"))

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{}`))
}