package handlers

import (
	"io"
	"log"
	"net/http"
)

func HandleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check Content-Type header
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		log.Printf("Unsupported Content-Type for metrics: %s", contentType)
		http.Error(w, "Only application/json Content-Type is supported", http.StatusUnsupportedMediaType)
		return
	}

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading metrics request body: %v", err)
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Write telemetry data to file
	if err := WriteTelemetryData("metrics", string(body)); err != nil {
		log.Printf("Error writing metrics data to file: %v", err)
	}

	// Log request details
	log.Printf("Received metrics request - Content-Type: %s, Content-Length: %d", 
		r.Header.Get("Content-Type"), len(body))

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{}`))
}