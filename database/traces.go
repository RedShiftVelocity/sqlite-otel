package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
)

// InsertResource inserts a resource and returns its ID
func InsertResource(tx *sql.Tx, resource map[string]interface{}) (int64, error) {
	attributes := getOrDefault(resource, "attributes", []interface{}{})
	attributesJSON, err := json.Marshal(attributes)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal resource attributes: %w", err)
	}

	schemaURL, _ := resource["schemaUrl"].(string)

	result, err := tx.Exec(
		"INSERT INTO resources (attributes, schema_url) VALUES (?, ?)",
		string(attributesJSON), schemaURL,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert resource: %w", err)
	}

	return result.LastInsertId()
}

// InsertScope inserts an instrumentation scope and returns its ID
func InsertScope(tx *sql.Tx, scope map[string]interface{}) (int64, error) {
	name, _ := scope["name"].(string)
	version, _ := scope["version"].(string)
	
	attributes := getOrDefault(scope, "attributes", []interface{}{})
	attributesJSON, err := json.Marshal(attributes)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal scope attributes: %w", err)
	}

	schemaURL, _ := scope["schemaUrl"].(string)

	result, err := tx.Exec(
		"INSERT INTO instrumentation_scopes (name, version, attributes, schema_url) VALUES (?, ?, ?, ?)",
		name, version, string(attributesJSON), schemaURL,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert scope: %w", err)
	}

	return result.LastInsertId()
}

// InsertTraceData processes and inserts trace data from OTLP JSON
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
		rsMap, ok := rs.(map[string]interface{})
		if !ok {
			continue
		}

		// Insert resource
		var resourceID int64
		if resource, ok := rsMap["resource"].(map[string]interface{}); ok {
			resourceID, err = InsertResource(tx, resource)
			if err != nil {
				return err
			}
		}

		// Process scope spans
		scopeSpans, ok := rsMap["scopeSpans"].([]interface{})
		if !ok {
			continue
		}

		for _, ss := range scopeSpans {
			ssMap, ok := ss.(map[string]interface{})
			if !ok {
				continue
			}

			// Insert scope
			var scopeID int64
			if scope, ok := ssMap["scope"].(map[string]interface{}); ok {
				scopeID, err = InsertScope(tx, scope)
				if err != nil {
					return err
				}
			}

			// Process spans
			spans, ok := ssMap["spans"].([]interface{})
			if !ok {
				continue
			}

			for _, span := range spans {
				spanMap, ok := span.(map[string]interface{})
				if !ok {
					continue
				}

				if err := InsertSpan(tx, spanMap, resourceID, scopeID); err != nil {
					return err
				}
			}
		}
	}

	return tx.Commit()
}

// InsertSpan inserts a single span with its events and links
func InsertSpan(tx *sql.Tx, span map[string]interface{}, resourceID, scopeID int64) error {
	// Extract span fields
	traceID, _ := span["traceId"].(string)
	spanID, _ := span["spanId"].(string)
	parentSpanID, _ := span["parentSpanId"].(string)
	traceState, _ := span["traceState"].(string)
	name, _ := span["name"].(string)
	
	kind := int64(0)
	if k, ok := span["kind"].(float64); ok {
		kind = int64(k)
	}

	startTime := parseTimeNano(span["startTimeUnixNano"])
	endTime := parseTimeNano(span["endTimeUnixNano"])

	attributes := getOrDefault(span, "attributes", []interface{}{})
	attributesJSON, _ := json.Marshal(attributes)
	
	droppedAttrsCount := int64(0)
	if d, ok := span["droppedAttributesCount"].(float64); ok {
		droppedAttrsCount = int64(d)
	}

	statusCode := int64(0)
	statusMessage := ""
	if status, ok := span["status"].(map[string]interface{}); ok {
		if code, ok := status["code"].(float64); ok {
			statusCode = int64(code)
		}
		statusMessage, _ = status["message"].(string)
	}

	flags := int64(0)
	if f, ok := span["flags"].(float64); ok {
		flags = int64(f)
	}

	// Insert span
	result, err := tx.Exec(`
		INSERT INTO spans (
			trace_id, span_id, parent_span_id, trace_state,
			name, kind, start_time_unix_nano, end_time_unix_nano,
			attributes, dropped_attributes_count,
			status_code, status_message, flags,
			resource_id, scope_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		traceID, spanID, parentSpanID, traceState,
		name, kind, startTime, endTime,
		string(attributesJSON), droppedAttrsCount,
		statusCode, statusMessage, flags,
		resourceID, scopeID,
	)
	if err != nil {
		return fmt.Errorf("failed to insert span: %w", err)
	}

	spanRowID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	// Insert events
	if events, ok := span["events"].([]interface{}); ok {
		for _, event := range events {
			if eventMap, ok := event.(map[string]interface{}); ok {
				if err := InsertSpanEvent(tx, eventMap, spanRowID); err != nil {
					return err
				}
			}
		}
	}

	// Insert links
	if links, ok := span["links"].([]interface{}); ok {
		for _, link := range links {
			if linkMap, ok := link.(map[string]interface{}); ok {
				if err := InsertSpanLink(tx, linkMap, spanRowID); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// InsertSpanEvent inserts a span event
func InsertSpanEvent(tx *sql.Tx, event map[string]interface{}, spanID int64) error {
	timeNano := parseTimeNano(event["timeUnixNano"])
	name, _ := event["name"].(string)
	
	attributes := getOrDefault(event, "attributes", []interface{}{})
	attributesJSON, _ := json.Marshal(attributes)
	
	droppedCount := int64(0)
	if d, ok := event["droppedAttributesCount"].(float64); ok {
		droppedCount = int64(d)
	}

	_, err := tx.Exec(`
		INSERT INTO span_events (span_id, time_unix_nano, name, attributes, dropped_attributes_count)
		VALUES (?, ?, ?, ?, ?)`,
		spanID, timeNano, name, string(attributesJSON), droppedCount,
	)
	return err
}

// InsertSpanLink inserts a span link
func InsertSpanLink(tx *sql.Tx, link map[string]interface{}, spanID int64) error {
	traceID, _ := link["traceId"].(string)
	linkedSpanID, _ := link["spanId"].(string)
	traceState, _ := link["traceState"].(string)
	
	attributes := getOrDefault(link, "attributes", []interface{}{})
	attributesJSON, _ := json.Marshal(attributes)
	
	droppedCount := int64(0)
	if d, ok := link["droppedAttributesCount"].(float64); ok {
		droppedCount = int64(d)
	}

	flags := int64(0)
	if f, ok := link["flags"].(float64); ok {
		flags = int64(f)
	}

	_, err := tx.Exec(`
		INSERT INTO span_links (span_id, trace_id, span_id_linked, trace_state, attributes, dropped_attributes_count, flags)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		spanID, traceID, linkedSpanID, traceState, string(attributesJSON), droppedCount, flags,
	)
	return err
}

// Helper function to parse time from various formats
func parseTimeNano(timeValue interface{}) int64 {
	switch v := timeValue.(type) {
	case string:
		if t, err := strconv.ParseInt(v, 10, 64); err == nil {
			return t
		}
	case float64:
		return int64(v)
	}
	return 0
}

// Helper to safely get values from map
func getOrDefault(m map[string]interface{}, key string, defaultValue interface{}) interface{} {
	if val, ok := m[key]; ok {
		return val
	}
	return defaultValue
}