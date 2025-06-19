package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/RedShiftVelocity/sqlite-otel/database"
)

func HandleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check Content-Type header
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		log.Printf("Unsupported Content-Type for logs: %s", contentType)
		http.Error(w, "Only application/json Content-Type is supported", http.StatusUnsupportedMediaType)
		return
	}

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading logs request body: %v", err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Write telemetry data to file (maintaining dual storage)
	if err := WriteTelemetryData("logs", string(body)); err != nil {
		log.Printf("Error writing logs data to file: %v", err)
	}

	// Parse JSON body
	var logsData map[string]interface{}
	if err := json.Unmarshal(body, &logsData); err != nil {
		log.Printf("Error parsing logs JSON: %v", err)
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Store logs in database
	if err := database.InsertLogsData(logsData); err != nil {
		log.Printf("Error storing logs in database: %v", err)
		// Return 500 Internal Server Error as per OTLP/HTTP spec
		http.Error(w, "Failed to process logs data", http.StatusInternalServerError)
		return
	}

	// Log request details
	log.Printf("Received logs request - Content-Type: %s, Content-Length: %d", 
		r.Header.Get("Content-Type"), len(body))

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{}`))
}