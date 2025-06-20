package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
)

// InsertMetricsData inserts metrics telemetry data into the database
func InsertMetricsData(data map[string]interface{}) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			// Transaction was already committed, which is fine
		}
	}()

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
		resource, ok := resourceMetric["resource"].(map[string]interface{})
		if !ok {
			// Resource field is required in ResourceMetrics
			return fmt.Errorf("invalid resourceMetric: missing resource field")
		}
		resourceID, err := GetOrCreateResource(tx, resource)
		if err != nil {
			return fmt.Errorf("failed to process resource: %w", err)
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
			scope, ok := scopeMetric["scope"].(map[string]interface{})
			if !ok {
				// Use default empty scope when not provided (per OTLP spec)
				scope = map[string]interface{}{
					"name":       "",
					"version":    "",
					"attributes": []interface{}{},
					"schemaUrl":  "",
				}
			}
			scopeID, err := GetOrCreateScope(tx, scope)
			if err != nil {
				return fmt.Errorf("failed to process scope: %w", err)
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

	// Get or create metric
	metricID, err := GetOrCreateMetric(tx, name, description, unit, metricType, resourceID, scopeID)
	if err != nil {
		return fmt.Errorf("failed to get or create metric: %w", err)
	}

	// Insert data points
	if data, ok := metric[dataField].(map[string]interface{}); ok {
		if dataPoints, ok := data["dataPoints"].([]interface{}); ok {
			for _, dp := range dataPoints {
				if dataPoint, ok := dp.(map[string]interface{}); ok {
					if err := InsertMetricDataPoint(tx, dataPoint, metricID, metricType); err != nil {
						return fmt.Errorf("failed to insert data point: %w", err)
					}
				}
			}
		}
	}

	return nil
}

// InsertMetricDataPoint inserts a single metric data point
func InsertMetricDataPoint(tx *sql.Tx, dp map[string]interface{}, metricID int64, metricType string) error {
	// Extract common fields
	attributes, _ := dp["attributes"]
	attributesJSON, err := json.Marshal(attributes)
	if err != nil {
		return fmt.Errorf("failed to marshal attributes: %w", err)
	}

	startTime := int64(0)
	if st, ok := dp["startTimeUnixNano"].(string); ok && st != "" {
		var err error
		startTime, err = parseTimeNano(st)
		if err != nil {
			return fmt.Errorf("failed to parse startTimeUnixNano: %w", err)
		}
	}

	timeUnix := int64(0)
	if t, ok := dp["timeUnixNano"].(string); ok && t != "" {
		var err error
		timeUnix, err = parseTimeNano(t)
		if err != nil {
			return fmt.Errorf("failed to parse timeUnixNano: %w", err)
		}
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

	// Handle simple values
	if v, ok := dp["asDouble"].(float64); ok {
		valueDouble = sql.NullFloat64{Float64: v, Valid: true}
	} else if v, ok := dp["asInt"].(string); ok {
		// Parse int from string
		if intVal, err := strconv.ParseInt(v, 10, 64); err == nil {
			valueInt = sql.NullInt64{Int64: intVal, Valid: true}
		} else {
			return fmt.Errorf("failed to parse asInt value '%s': %w", v, err)
		}
	}

	// Handle complex metric types
	// TODO: Future enhancement - add dedicated columns for histogram/summary data
	// For now, store complex metric data in attributes as JSON
	complexData := make(map[string]interface{})
	
	switch metricType {
	case "histogram":
		if count, ok := dp["count"].(string); ok {
			complexData["count"] = count
		}
		if sum, ok := dp["sum"].(float64); ok {
			complexData["sum"] = sum
		}
		if bucketCounts, ok := dp["bucketCounts"].([]interface{}); ok {
			complexData["bucketCounts"] = bucketCounts
		}
		if explicitBounds, ok := dp["explicitBounds"].([]interface{}); ok {
			complexData["explicitBounds"] = explicitBounds
		}
	case "exponentialHistogram":
		if count, ok := dp["count"].(string); ok {
			complexData["count"] = count
		}
		if sum, ok := dp["sum"].(float64); ok {
			complexData["sum"] = sum
		}
		if scale, ok := dp["scale"].(float64); ok {
			complexData["scale"] = scale
		}
		if zeroCount, ok := dp["zeroCount"].(string); ok {
			complexData["zeroCount"] = zeroCount
		}
		if positive, ok := dp["positive"].(map[string]interface{}); ok {
			complexData["positive"] = positive
		}
		if negative, ok := dp["negative"].(map[string]interface{}); ok {
			complexData["negative"] = negative
		}
	case "summary":
		if count, ok := dp["count"].(string); ok {
			complexData["count"] = count
		}
		if sum, ok := dp["sum"].(float64); ok {
			complexData["sum"] = sum
		}
		if quantileValues, ok := dp["quantileValues"].([]interface{}); ok {
			complexData["quantileValues"] = quantileValues
		}
	}

	// Merge complex data into attributes
	if len(complexData) > 0 {
		if attributes == nil {
			attributes = make(map[string]interface{})
		}
		if attrsMap, ok := attributes.(map[string]interface{}); ok {
			attrsMap["_metricData"] = complexData
			attributesJSON, err = json.Marshal(attrsMap)
			if err != nil {
				return fmt.Errorf("failed to marshal attributes with metric data: %w", err)
			}
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