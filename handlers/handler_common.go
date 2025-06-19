package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
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
		log.Printf("Unsupported Content-Type for %s: %s", telemetryType, contentType)
		http.Error(w, "Only application/json Content-Type is supported", http.StatusUnsupportedMediaType)
		return
	}

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading %s request body: %v", telemetryType, err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Write telemetry data to file (maintaining dual storage)
	if err := WriteTelemetryData(telemetryType, string(body)); err != nil {
		log.Printf("Error writing %s data to file: %v", telemetryType, err)
		http.Error(w, fmt.Sprintf("Failed to write %s data", telemetryType), http.StatusInternalServerError)
		return
	}

	// Parse JSON body
	var telemetryData map[string]interface{}
	if err := json.Unmarshal(body, &telemetryData); err != nil {
		log.Printf("Error parsing %s JSON: %v", telemetryType, err)
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Store telemetry data in database
	if err := insertFunc(telemetryData); err != nil {
		log.Printf("Error storing %s in database: %v", telemetryType, err)
		// Return 500 Internal Server Error as per OTLP/HTTP spec
		http.Error(w, fmt.Sprintf("Failed to process %s data", telemetryType), http.StatusInternalServerError)
		return
	}

	// Log request details
	log.Printf("Received %s request - Content-Type: %s, Content-Length: %d", 
		telemetryType, r.Header.Get("Content-Type"), len(body))

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{}`))
}