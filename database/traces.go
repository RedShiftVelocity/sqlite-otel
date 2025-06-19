package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

// GetOrCreateResource finds an existing resource or creates a new one
func GetOrCreateResource(tx *sql.Tx, resource map[string]interface{}) (int64, error) {
	attributes := getOrDefault(resource, "attributes", []interface{}{})
	attributesJSON, err := json.Marshal(attributes)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal resource attributes: %w", err)
	}

	schemaURL, _ := resource["schemaUrl"].(string)

	// Check if resource already exists
	var id int64
	err = tx.QueryRow(
		"SELECT id FROM resources WHERE attributes = ? AND (schema_url = ? OR (schema_url IS NULL AND ? IS NULL))",
		string(attributesJSON), schemaURL, schemaURL,
	).Scan(&id)
	
	if err == nil {
		return id, nil // Found existing resource
	}
	
	if err != sql.ErrNoRows {
		return 0, fmt.Errorf("failed to query for existing resource: %w", err)
	}

	// Resource doesn't exist, insert it
	result, err := tx.Exec(
		"INSERT INTO resources (attributes, schema_url) VALUES (?, ?)",
		string(attributesJSON), schemaURL,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert resource: %w", err)
	}

	return result.LastInsertId()
}

// GetOrCreateScope finds an existing scope or creates a new one
func GetOrCreateScope(tx *sql.Tx, scope map[string]interface{}) (int64, error) {
	name, ok := scope["name"].(string)
	if !ok {
		return 0, fmt.Errorf("invalid scope data: name is missing or not a string")
	}
	
	version, _ := scope["version"].(string)
	
	attributes := getOrDefault(scope, "attributes", []interface{}{})
	attributesJSON, err := json.Marshal(attributes)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal scope attributes: %w", err)
	}

	schemaURL, _ := scope["schemaUrl"].(string)

	// Check if scope already exists
	var id int64
	err = tx.QueryRow(
		`SELECT id FROM instrumentation_scopes 
		WHERE name = ? AND version = ? AND attributes = ? 
		AND (schema_url = ? OR (schema_url IS NULL AND ? IS NULL))`,
		name, version, string(attributesJSON), schemaURL, schemaURL,
	).Scan(&id)
	
	if err == nil {
		return id, nil // Found existing scope
	}
	
	if err != sql.ErrNoRows {
		return 0, fmt.Errorf("failed to query for existing scope: %w", err)
	}

	// Scope doesn't exist, insert it
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

	// Prepare statements for better performance
	spanStmt, err := tx.Prepare(`
		INSERT INTO spans (
			trace_id, span_id, parent_span_id, trace_state,
			name, kind, start_time_unix_nano, end_time_unix_nano,
			attributes, dropped_attributes_count,
			status_code, status_message, flags,
			resource_id, scope_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("failed to prepare span insert statement: %w", err)
	}
	defer spanStmt.Close()

	eventStmt, err := tx.Prepare(`
		INSERT INTO span_events (span_id, time_unix_nano, name, attributes, dropped_attributes_count)
		VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("failed to prepare event insert statement: %w", err)
	}
	defer eventStmt.Close()

	linkStmt, err := tx.Prepare(`
		INSERT INTO span_links (span_id, trace_id, span_id_linked, trace_state, attributes, dropped_attributes_count, flags)
		VALUES (?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("failed to prepare link insert statement: %w", err)
	}
	defer linkStmt.Close()

	resourceSpans, ok := data["resourceSpans"].([]interface{})
	if !ok {
		return fmt.Errorf("invalid trace data: missing resourceSpans")
	}

	for _, rs := range resourceSpans {
		rsMap, ok := rs.(map[string]interface{})
		if !ok {
			continue
		}

		// Get or create resource
		var resourceID int64
		if resource, ok := rsMap["resource"].(map[string]interface{}); ok {
			resourceID, err = GetOrCreateResource(tx, resource)
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

			// Get or create scope
			var scopeID int64
			if scope, ok := ssMap["scope"].(map[string]interface{}); ok {
				scopeID, err = GetOrCreateScope(tx, scope)
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

				if err := InsertSpan(spanStmt, eventStmt, linkStmt, spanMap, resourceID, scopeID); err != nil {
					return err
				}
			}
		}
	}

	return tx.Commit()
}

// InsertSpan inserts a single span with its events and links
func InsertSpan(spanStmt, eventStmt, linkStmt *sql.Stmt, span map[string]interface{}, resourceID, scopeID int64) error {
	// Extract span fields with error checking
	traceID, ok := span["traceId"].(string)
	if !ok || traceID == "" {
		return fmt.Errorf("invalid span data: traceId is missing or not a string")
	}
	
	spanID, ok := span["spanId"].(string)
	if !ok || spanID == "" {
		return fmt.Errorf("invalid span data: spanId is missing or not a string")
	}
	
	parentSpanID, _ := span["parentSpanId"].(string)
	traceState, _ := span["traceState"].(string)
	
	name, ok := span["name"].(string)
	if !ok {
		return fmt.Errorf("invalid span data: name is missing or not a string")
	}
	
	kind := int64(0)
	if k, ok := span["kind"].(float64); ok {
		kind = int64(k)
	}

	startTime := parseTimeNano(span["startTimeUnixNano"])
	endTime := parseTimeNano(span["endTimeUnixNano"])

	attributes := getOrDefault(span, "attributes", []interface{}{})
	attributesJSON, err := json.Marshal(attributes)
	if err != nil {
		return fmt.Errorf("failed to marshal span attributes: %w", err)
	}
	
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

	// Insert span using prepared statement
	result, err := spanStmt.Exec(
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
				if err := InsertSpanEvent(eventStmt, eventMap, spanRowID); err != nil {
					return err
				}
			}
		}
	}

	// Insert links
	if links, ok := span["links"].([]interface{}); ok {
		for _, link := range links {
			if linkMap, ok := link.(map[string]interface{}); ok {
				if err := InsertSpanLink(linkStmt, linkMap, spanRowID); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// InsertSpanEvent inserts a span event
func InsertSpanEvent(eventStmt *sql.Stmt, event map[string]interface{}, spanID int64) error {
	timeNano := parseTimeNano(event["timeUnixNano"])
	
	name, ok := event["name"].(string)
	if !ok {
		return fmt.Errorf("invalid event data: name is missing or not a string")
	}
	
	attributes := getOrDefault(event, "attributes", []interface{}{})
	attributesJSON, err := json.Marshal(attributes)
	if err != nil {
		return fmt.Errorf("failed to marshal event attributes: %w", err)
	}
	
	droppedCount := int64(0)
	if d, ok := event["droppedAttributesCount"].(float64); ok {
		droppedCount = int64(d)
	}

	_, err = eventStmt.Exec(spanID, timeNano, name, string(attributesJSON), droppedCount)
	return err
}

// InsertSpanLink inserts a span link
func InsertSpanLink(linkStmt *sql.Stmt, link map[string]interface{}, spanID int64) error {
	traceID, ok := link["traceId"].(string)
	if !ok || traceID == "" {
		return fmt.Errorf("invalid link data: traceId is missing or not a string")
	}
	
	linkedSpanID, ok := link["spanId"].(string)
	if !ok || linkedSpanID == "" {
		return fmt.Errorf("invalid link data: spanId is missing or not a string")
	}
	
	traceState, _ := link["traceState"].(string)
	
	attributes := getOrDefault(link, "attributes", []interface{}{})
	attributesJSON, err := json.Marshal(attributes)
	if err != nil {
		return fmt.Errorf("failed to marshal link attributes: %w", err)
	}
	
	droppedCount := int64(0)
	if d, ok := link["droppedAttributesCount"].(float64); ok {
		droppedCount = int64(d)
	}

	flags := int64(0)
	if f, ok := link["flags"].(float64); ok {
		flags = int64(f)
	}

	_, err = linkStmt.Exec(spanID, traceID, linkedSpanID, traceState, string(attributesJSON), droppedCount, flags)
	return err
}

