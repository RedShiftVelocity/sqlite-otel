package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/RedShiftVelocity/sqlite-otel/database"
)

func HandleTraces(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check Content-Type header
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		log.Printf("Unsupported Content-Type for traces: %s", contentType)
		http.Error(w, "Only application/json Content-Type is supported", http.StatusUnsupportedMediaType)
		return
	}

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading traces request body: %v", err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Write telemetry data to file (maintaining dual storage)
	if err := WriteTelemetryData("trace", string(body)); err != nil {
		log.Printf("Error writing traces data to file: %v", err)
	}

	// Parse JSON body
	var tracesData map[string]interface{}
	if err := json.Unmarshal(body, &tracesData); err != nil {
		log.Printf("Error parsing traces JSON: %v", err)
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Store traces in database
	if err := database.InsertTraceData(tracesData); err != nil {
		log.Printf("Error storing traces in database: %v", err)
		// Return 500 Internal Server Error as per OTLP/HTTP spec
		http.Error(w, "Failed to process trace data", http.StatusInternalServerError)
		return
	}

	// Log request details
	log.Printf("Received traces request - Content-Type: %s, Content-Length: %d", 
		r.Header.Get("Content-Type"), len(body))

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{}`))
}