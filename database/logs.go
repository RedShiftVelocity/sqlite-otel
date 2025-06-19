package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

// InsertLogsData processes and inserts logs data from OTLP JSON
func InsertLogsData(data map[string]interface{}) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Prepare log record insert statement
	logStmt, err := tx.Prepare(`
		INSERT INTO log_records (
			time_unix_nano, observed_time_unix_nano,
			severity_number, severity_text,
			body, attributes,
			trace_id, span_id, trace_flags,
			resource_id, scope_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("failed to prepare log insert statement: %w", err)
	}
	defer logStmt.Close()

	resourceLogs, ok := data["resourceLogs"].([]interface{})
	if !ok {
		return fmt.Errorf("invalid logs data: missing resourceLogs")
	}

	for _, rl := range resourceLogs {
		rlMap, ok := rl.(map[string]interface{})
		if !ok {
			continue
		}

		// Get or create resource
		var resourceID int64
		if resource, ok := rlMap["resource"].(map[string]interface{}); ok {
			resourceID, err = GetOrCreateResource(tx, resource)
			if err != nil {
				return err
			}
		}

		// Process scope logs
		scopeLogs, ok := rlMap["scopeLogs"].([]interface{})
		if !ok {
			continue
		}

		for _, sl := range scopeLogs {
			slMap, ok := sl.(map[string]interface{})
			if !ok {
				continue
			}

			// Get or create scope
			var scopeID int64
			if scope, ok := slMap["scope"].(map[string]interface{}); ok {
				scopeID, err = GetOrCreateScope(tx, scope)
				if err != nil {
					return err
				}
			}

			// Process log records
			logRecords, ok := slMap["logRecords"].([]interface{})
			if !ok {
				continue
			}

			for _, logRecord := range logRecords {
				logMap, ok := logRecord.(map[string]interface{})
				if !ok {
					continue
				}

				if err := InsertLogRecord(logStmt, logMap, resourceID, scopeID); err != nil {
					return err
				}
			}
		}
	}

	return tx.Commit()
}

// InsertLogRecord inserts a single log record
func InsertLogRecord(logStmt *sql.Stmt, logRecord map[string]interface{}, resourceID, scopeID int64) error {
	// Parse timestamps
	var timeUnixNano sql.NullInt64
	if t := parseTimeNano(logRecord["timeUnixNano"]); t > 0 {
		timeUnixNano = sql.NullInt64{Int64: t, Valid: true}
	}

	observedTimeUnixNano := parseTimeNano(logRecord["observedTimeUnixNano"])
	if observedTimeUnixNano == 0 {
		return fmt.Errorf("invalid log record: observedTimeUnixNano is required")
	}

	// Parse severity
	var severityNumber sql.NullInt64
	if sn, ok := logRecord["severityNumber"].(float64); ok {
		severityNumber = sql.NullInt64{Int64: int64(sn), Valid: true}
	}
	severityText, _ := logRecord["severityText"].(string)

	// Parse body (can be various types)
	bodyJSON, err := json.Marshal(logRecord["body"])
	if err != nil {
		return fmt.Errorf("failed to marshal log body: %w", err)
	}

	// Parse attributes
	attributes := getOrDefault(logRecord, "attributes", []interface{}{})
	attributesJSON, err := json.Marshal(attributes)
	if err != nil {
		return fmt.Errorf("failed to marshal log attributes: %w", err)
	}

	// Parse trace correlation
	traceID, _ := logRecord["traceId"].(string)
	spanID, _ := logRecord["spanId"].(string)
	
	traceFlags := int64(0)
	if tf, ok := logRecord["flags"].(float64); ok {
		traceFlags = int64(tf)
	}

	_, err = logStmt.Exec(
		timeUnixNano, observedTimeUnixNano,
		severityNumber, severityText,
		string(bodyJSON), string(attributesJSON),
		traceID, spanID, traceFlags,
		resourceID, scopeID,
	)
	
	return err
}

