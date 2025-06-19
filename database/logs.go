package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

// InsertLogsData inserts logs telemetry data into the database
func InsertLogsData(data map[string]interface{}) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	resourceLogs, ok := data["resourceLogs"].([]interface{})
	if !ok {
		return fmt.Errorf("invalid logs data: missing resourceLogs")
	}

	for _, rl := range resourceLogs {
		resourceLog, ok := rl.(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid resourceLog type: expected map[string]interface{}, got %T", rl)
		}

		// Get or create resource
		var resourceID int64
		if resource, ok := resourceLog["resource"].(map[string]interface{}); ok {
			resourceID, err = GetOrCreateResource(tx, resource)
			if err != nil {
				return fmt.Errorf("failed to process resource: %w", err)
			}
		}

		// Process scope logs
		scopeLogs, ok := resourceLog["scopeLogs"].([]interface{})
		if !ok {
			return fmt.Errorf("invalid scopeLogs type in resourceLog: expected []interface{}, got %T", resourceLog["scopeLogs"])
		}

		for _, sl := range scopeLogs {
			scopeLog, ok := sl.(map[string]interface{})
			if !ok {
				return fmt.Errorf("invalid scopeLog type: expected map[string]interface{}, got %T", sl)
			}

			// Get or create scope
			var scopeID int64
			if scope, ok := scopeLog["scope"].(map[string]interface{}); ok {
				scopeID, err = GetOrCreateScope(tx, scope)
				if err != nil {
					return fmt.Errorf("failed to process scope: %w", err)
				}
			}

			// Process log records
			logRecords, ok := scopeLog["logRecords"].([]interface{})
			if !ok {
				return fmt.Errorf("invalid logRecords type in scopeLog: expected []interface{}, got %T", scopeLog["logRecords"])
			}

			for _, lr := range logRecords {
				logRecord, ok := lr.(map[string]interface{})
				if !ok {
					return fmt.Errorf("invalid logRecord type: expected map[string]interface{}, got %T", lr)
				}

				if err := InsertLogRecord(tx, logRecord, resourceID, scopeID); err != nil {
					return fmt.Errorf("failed to insert log record: %w", err)
				}
			}
		}
	}

	return tx.Commit()
}

// InsertLogRecord inserts a single log record into the database
func InsertLogRecord(tx *sql.Tx, logRecord map[string]interface{}, resourceID, scopeID int64) error {
	// Parse timestamps
	timeUnix := int64(0)
	if t, ok := logRecord["timeUnixNano"].(string); ok && t != "" {
		var err error
		timeUnix, err = parseTimeNano(t)
		if err != nil {
			return fmt.Errorf("failed to parse timeUnixNano: %w", err)
		}
	}

	observedTime := int64(0)
	if ot, ok := logRecord["observedTimeUnixNano"].(string); ok && ot != "" {
		var err error
		observedTime, err = parseTimeNano(ot)
		if err != nil {
			return fmt.Errorf("failed to parse observedTimeUnixNano: %w", err)
		}
	}

	// Extract severity
	severityNumber := int64(0)
	if sn, ok := logRecord["severityNumber"].(float64); ok {
		severityNumber = int64(sn)
	}
	
	// Extract severity text with type checking
	var severityText string
	if st, ok := logRecord["severityText"]; ok && st != nil {
		var ok bool
		severityText, ok = st.(string)
		if !ok {
			return fmt.Errorf("invalid type for severityText: expected string, got %T", st)
		}
	}

	// Extract body (optional field)
	var bodyJSON []byte
	if body, ok := logRecord["body"]; ok && body != nil {
		var err error
		bodyJSON, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal log body: %w", err)
		}
	} else {
		// If body is not present or nil, use empty JSON object
		bodyJSON = []byte("{}")
	}

	// Extract attributes (optional field)
	var attributesJSON []byte
	if attributes, ok := logRecord["attributes"]; ok && attributes != nil {
		var err error
		attributesJSON, err = json.Marshal(attributes)
		if err != nil {
			return fmt.Errorf("failed to marshal log attributes: %w", err)
		}
	} else {
		// If attributes not present or nil, use empty JSON array
		attributesJSON = []byte("[]")
	}

	// Extract trace context with type checking
	var traceID string
	if tid, ok := logRecord["traceId"]; ok && tid != nil {
		var ok bool
		traceID, ok = tid.(string)
		if !ok {
			return fmt.Errorf("invalid type for traceId: expected string, got %T", tid)
		}
	}
	
	var spanID string
	if sid, ok := logRecord["spanId"]; ok && sid != nil {
		var ok bool
		spanID, ok = sid.(string)
		if !ok {
			return fmt.Errorf("invalid type for spanId: expected string, got %T", sid)
		}
	}

	// Extract flags
	flags := int64(0)
	if f, ok := logRecord["flags"].(float64); ok {
		flags = int64(f)
	}

	// Insert log record
	_, err := tx.Exec(`
		INSERT INTO log_records (
			time_unix_nano, observed_time_unix_nano, severity_number, severity_text,
			body, attributes, trace_id, span_id, flags, resource_id, scope_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		timeUnix, observedTime, severityNumber, severityText,
		string(bodyJSON), string(attributesJSON), traceID, spanID,
		flags, resourceID, scopeID,
	)

	return err
}