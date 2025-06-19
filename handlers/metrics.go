package handlers

import (
	"net/http"

	"github.com/RedShiftVelocity/sqlite-otel/database"
)

func HandleMetrics(w http.ResponseWriter, r *http.Request) {
	ProcessTelemetryRequest(w, r, "metrics", database.InsertMetricsData)
}