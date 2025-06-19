package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

// InsertMetricsData processes and inserts metrics data from OTLP JSON
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
		rmMap, ok := rm.(map[string]interface{})
		if !ok {
			continue
		}

		// Insert resource
		var resourceID int64
		if resource, ok := rmMap["resource"].(map[string]interface{}); ok {
			resourceID, err = InsertResource(tx, resource)
			if err != nil {
				return err
			}
		}

		// Process scope metrics
		scopeMetrics, ok := rmMap["scopeMetrics"].([]interface{})
		if !ok {
			continue
		}

		for _, sm := range scopeMetrics {
			smMap, ok := sm.(map[string]interface{})
			if !ok {
				continue
			}

			// Insert scope
			var scopeID int64
			if scope, ok := smMap["scope"].(map[string]interface{}); ok {
				scopeID, err = InsertScope(tx, scope)
				if err != nil {
					return err
				}
			}

			// Process metrics
			metrics, ok := smMap["metrics"].([]interface{})
			if !ok {
				continue
			}

			for _, metric := range metrics {
				metricMap, ok := metric.(map[string]interface{})
				if !ok {
					continue
				}

				if err := InsertMetric(tx, metricMap, resourceID, scopeID); err != nil {
					return err
				}
			}
		}
	}

	return tx.Commit()
}

// InsertMetric inserts a metric and its data points
func InsertMetric(tx *sql.Tx, metric map[string]interface{}, resourceID, scopeID int64) error {
	// Extract metric metadata
	name, _ := metric["name"].(string)
	description, _ := metric["description"].(string)
	unit, _ := metric["unit"].(string)

	// Determine metric type
	metricType := ""
	var dataPoints []interface{}

	if gauge, ok := metric["gauge"].(map[string]interface{}); ok {
		metricType = "gauge"
		dataPoints, _ = gauge["dataPoints"].([]interface{})
	} else if sum, ok := metric["sum"].(map[string]interface{}); ok {
		metricType = "sum"
		dataPoints, _ = sum["dataPoints"].([]interface{})
	} else if histogram, ok := metric["histogram"].(map[string]interface{}); ok {
		metricType = "histogram"
		dataPoints, _ = histogram["dataPoints"].([]interface{})
	} else if expHistogram, ok := metric["exponentialHistogram"].(map[string]interface{}); ok {
		metricType = "exponential_histogram"
		dataPoints, _ = expHistogram["dataPoints"].([]interface{})
	} else if summary, ok := metric["summary"].(map[string]interface{}); ok {
		metricType = "summary"
		dataPoints, _ = summary["dataPoints"].([]interface{})
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
		return err
	}

	// Insert data points
	for _, dp := range dataPoints {
		dpMap, ok := dp.(map[string]interface{})
		if !ok {
			continue
		}

		if err := InsertMetricDataPoint(tx, dpMap, metricID, metric); err != nil {
			return err
		}
	}

	return nil
}

// InsertMetricDataPoint inserts a metric data point
func InsertMetricDataPoint(tx *sql.Tx, dp map[string]interface{}, metricID int64, metric map[string]interface{}) error {
	// Common fields
	attributes, _ := dp["attributes"]
	attributesJSON, _ := json.Marshal(attributes)
	startTime := parseTimeNano(dp["startTimeUnixNano"])
	timeNano := parseTimeNano(dp["timeUnixNano"])
	flags := int64(0)
	if f, ok := dp["flags"].(float64); ok {
		flags = int64(f)
	}

	// Simplified insert - just store basic fields for now
	_, err := tx.Exec(`
		INSERT INTO metric_data_points (
			metric_id, attributes, start_time_unix_nano, time_unix_nano, flags
		) VALUES (?, ?, ?, ?, ?)`,
		metricID, string(attributesJSON), startTime, timeNano, flags,
	)
	
	return err
}