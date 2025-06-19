package handlers

import (
	"net/http"

	"github.com/RedShiftVelocity/sqlite-otel/database"
)

func HandleTraces(w http.ResponseWriter, r *http.Request) {
	ProcessTelemetryRequest(w, r, "trace", database.InsertTraceData)
}