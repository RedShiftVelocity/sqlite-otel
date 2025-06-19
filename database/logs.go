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
			continue
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
			continue
		}

		for _, sl := range scopeLogs {
			scopeLog, ok := sl.(map[string]interface{})
			if !ok {
				continue
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
				continue
			}

			for _, lr := range logRecords {
				logRecord, ok := lr.(map[string]interface{})
				if !ok {
					continue
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
	if t, ok := logRecord["timeUnixNano"].(string); ok {
		timeUnix = parseTimeNano(t)
	}

	observedTime := int64(0)
	if ot, ok := logRecord["observedTimeUnixNano"].(string); ok {
		observedTime = parseTimeNano(ot)
	}

	// Extract severity
	severityNumber := int64(0)
	if sn, ok := logRecord["severityNumber"].(float64); ok {
		severityNumber = int64(sn)
	}
	severityText, _ := logRecord["severityText"].(string)

	// Extract body
	var bodyJSON []byte
	var err error
	if body, ok := logRecord["body"]; ok {
		bodyJSON, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal log body: %w", err)
		}
	}

	// Extract attributes
	attributes, _ := logRecord["attributes"]
	attributesJSON, err := json.Marshal(attributes)
	if err != nil {
		return fmt.Errorf("failed to marshal log attributes: %w", err)
	}

	// Extract trace context
	traceID, _ := logRecord["traceId"].(string)
	spanID, _ := logRecord["spanId"].(string)

	// Extract flags
	flags := int64(0)
	if f, ok := logRecord["flags"].(float64); ok {
		flags = int64(f)
	}

	// Insert log record
	_, err = tx.Exec(`
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