package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

// InsertTraceData inserts trace telemetry data into the database
func InsertTraceData(data map[string]interface{}) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	resourceSpans, ok := data["resourceSpans"].([]interface{})
	if !ok {
		return fmt.Errorf("invalid trace data: missing resourceSpans")
	}

	for _, rs := range resourceSpans {
		resourceSpan, ok := rs.(map[string]interface{})
		if !ok {
			continue
		}

		// Get or create resource
		var resourceID int64
		if resource, ok := resourceSpan["resource"].(map[string]interface{}); ok {
			resourceID, err = GetOrCreateResource(tx, resource)
			if err != nil {
				return fmt.Errorf("failed to process resource: %w", err)
			}
		}

		// Process scope spans
		scopeSpans, ok := resourceSpan["scopeSpans"].([]interface{})
		if !ok {
			continue
		}

		for _, ss := range scopeSpans {
			scopeSpan, ok := ss.(map[string]interface{})
			if !ok {
				continue
			}

			// Get or create scope
			var scopeID int64
			if scope, ok := scopeSpan["scope"].(map[string]interface{}); ok {
				scopeID, err = GetOrCreateScope(tx, scope)
				if err != nil {
					return fmt.Errorf("failed to process scope: %w", err)
				}
			}

			// Process spans
			spans, ok := scopeSpan["spans"].([]interface{})
			if !ok {
				continue
			}

			for _, s := range spans {
				span, ok := s.(map[string]interface{})
				if !ok {
					continue
				}

				if err := InsertSpan(tx, span, resourceID, scopeID); err != nil {
					return fmt.Errorf("failed to insert span: %w", err)
				}
			}
		}
	}

	return tx.Commit()
}

// InsertSpan inserts a single span into the database
func InsertSpan(tx *sql.Tx, span map[string]interface{}, resourceID, scopeID int64) error {
	// Extract required fields
	traceID, ok := span["traceId"].(string)
	if !ok || traceID == "" {
		return fmt.Errorf("invalid span: traceId is required")
	}

	spanID, ok := span["spanId"].(string)
	if !ok || spanID == "" {
		return fmt.Errorf("invalid span: spanId is required")
	}

	// Extract optional fields
	traceState, _ := span["traceState"].(string)
	parentSpanID, _ := span["parentSpanId"].(string)
	name, _ := span["name"].(string)
	kind := int64(0)
	if k, ok := span["kind"].(float64); ok {
		kind = int64(k)
	}

	// Parse timestamps
	startTime := int64(0)
	if st, ok := span["startTimeUnixNano"].(string); ok {
		startTime = parseTimeNano(st)
	}
	endTime := int64(0)
	if et, ok := span["endTimeUnixNano"].(string); ok {
		endTime = parseTimeNano(et)
	}

	// Marshal complex fields to JSON
	attributes, _ := span["attributes"]
	attributesJSON, err := json.Marshal(attributes)
	if err != nil {
		return fmt.Errorf("failed to marshal span attributes: %w", err)
	}

	events, _ := span["events"]
	eventsJSON, err := json.Marshal(events)
	if err != nil {
		return fmt.Errorf("failed to marshal span events: %w", err)
	}

	links, _ := span["links"]
	linksJSON, err := json.Marshal(links)
	if err != nil {
		return fmt.Errorf("failed to marshal span links: %w", err)
	}

	// Extract status
	statusCode := int64(0)
	statusMessage := ""
	if status, ok := span["status"].(map[string]interface{}); ok {
		if code, ok := status["code"].(float64); ok {
			statusCode = int64(code)
		}
		statusMessage, _ = status["message"].(string)
	}

	// Insert span
	_, err = tx.Exec(`
		INSERT INTO spans (
			trace_id, span_id, trace_state, parent_span_id, name, kind,
			start_time_unix_nano, end_time_unix_nano, attributes, events, links,
			status_code, status_message, resource_id, scope_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		traceID, spanID, traceState, parentSpanID, name, kind,
		startTime, endTime, string(attributesJSON), string(eventsJSON), string(linksJSON),
		statusCode, statusMessage, resourceID, scopeID,
	)

	return err
}