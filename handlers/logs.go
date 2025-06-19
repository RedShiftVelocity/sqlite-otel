package handlers

import (
	"net/http"

	"github.com/RedShiftVelocity/sqlite-otel/database"
)

func HandleLogs(w http.ResponseWriter, r *http.Request) {
	ProcessTelemetryRequest(w, r, "logs", database.InsertLogsData)
}