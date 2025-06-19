package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

// InsertMetricsData inserts metrics telemetry data into the database
func InsertMetricsData(data map[string]interface{}) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	resourceMetrics, ok := data["resourceMetrics"].([]interface{})
	if !ok {
		return fmt.Errorf("invalid metrics data: missing resourceMetrics")
	}

	for _, rm := range resourceMetrics {
		resourceMetric, ok := rm.(map[string]interface{})
		if !ok {
			continue
		}

		// Get or create resource
		var resourceID int64
		if resource, ok := resourceMetric["resource"].(map[string]interface{}); ok {
			resourceID, err = GetOrCreateResource(tx, resource)
			if err != nil {
				return fmt.Errorf("failed to process resource: %w", err)
			}
		}

		// Process scope metrics
		scopeMetrics, ok := resourceMetric["scopeMetrics"].([]interface{})
		if !ok {
			continue
		}

		for _, sm := range scopeMetrics {
			scopeMetric, ok := sm.(map[string]interface{})
			if !ok {
				continue
			}

			// Get or create scope
			var scopeID int64
			if scope, ok := scopeMetric["scope"].(map[string]interface{}); ok {
				scopeID, err = GetOrCreateScope(tx, scope)
				if err != nil {
					return fmt.Errorf("failed to process scope: %w", err)
				}
			}

			// Process metrics
			metrics, ok := scopeMetric["metrics"].([]interface{})
			if !ok {
				continue
			}

			for _, m := range metrics {
				metric, ok := m.(map[string]interface{})
				if !ok {
					continue
				}

				if err := InsertMetric(tx, metric, resourceID, scopeID); err != nil {
					return fmt.Errorf("failed to insert metric: %w", err)
				}
			}
		}
	}

	return tx.Commit()
}

// InsertMetric inserts a single metric and its data points
func InsertMetric(tx *sql.Tx, metric map[string]interface{}, resourceID, scopeID int64) error {
	// Extract metric metadata
	name, ok := metric["name"].(string)
	if !ok || name == "" {
		return fmt.Errorf("invalid metric: name is required")
	}

	description, _ := metric["description"].(string)
	unit, _ := metric["unit"].(string)

	// Determine metric type
	var metricType, dataField string
	for _, mt := range []string{"gauge", "sum", "histogram", "exponentialHistogram", "summary"} {
		if _, ok := metric[mt]; ok {
			metricType = mt
			dataField = mt
			break
		}
	}
	if metricType == "" {
		return fmt.Errorf("unknown metric type for metric: %s", name)
	}

	// Insert metric
	result, err := tx.Exec(
		"INSERT INTO metrics (name, description, unit, type, resource_id, scope_id) VALUES (?, ?, ?, ?, ?, ?)",
		name, description, unit, metricType, resourceID, scopeID,
	)
	if err != nil {
		return fmt.Errorf("failed to insert metric: %w", err)
	}

	metricID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get metric ID: %w", err)
	}

	// Insert data points
	if data, ok := metric[dataField].(map[string]interface{}); ok {
		if dataPoints, ok := data["dataPoints"].([]interface{}); ok {
			for _, dp := range dataPoints {
				if dataPoint, ok := dp.(map[string]interface{}); ok {
					if err := InsertMetricDataPoint(tx, dataPoint, metricID); err != nil {
						return fmt.Errorf("failed to insert data point: %w", err)
					}
				}
			}
		}
	}

	return nil
}

// InsertMetricDataPoint inserts a single metric data point
func InsertMetricDataPoint(tx *sql.Tx, dp map[string]interface{}, metricID int64) error {
	// Extract common fields
	attributes, _ := dp["attributes"]
	attributesJSON, err := json.Marshal(attributes)
	if err != nil {
		return fmt.Errorf("failed to marshal attributes: %w", err)
	}

	startTime := int64(0)
	if st, ok := dp["startTimeUnixNano"].(string); ok {
		startTime = parseTimeNano(st)
	}

	timeUnix := int64(0)
	if t, ok := dp["timeUnixNano"].(string); ok {
		timeUnix = parseTimeNano(t)
	}

	exemplars, _ := dp["exemplars"]
	exemplarsJSON, err := json.Marshal(exemplars)
	if err != nil {
		return fmt.Errorf("failed to marshal exemplars: %w", err)
	}

	flags := int64(0)
	if f, ok := dp["flags"].(float64); ok {
		flags = int64(f)
	}

	// Extract value based on type
	var valueDouble sql.NullFloat64
	var valueInt sql.NullInt64

	if v, ok := dp["asDouble"].(float64); ok {
		valueDouble = sql.NullFloat64{Float64: v, Valid: true}
	} else if v, ok := dp["asInt"].(string); ok {
		// Parse int from string
		var intVal int64
		if _, err := fmt.Sscanf(v, "%d", &intVal); err == nil {
			valueInt = sql.NullInt64{Int64: intVal, Valid: true}
		}
	}

	// Insert data point
	_, err = tx.Exec(`
		INSERT INTO metric_data_points (
			metric_id, attributes, start_time_unix_nano, time_unix_nano,
			value_double, value_int, exemplars, flags
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		metricID, string(attributesJSON), startTime, timeUnix,
		valueDouble, valueInt, string(exemplarsJSON), flags,
	)

	return err
}