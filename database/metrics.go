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

	// Insert data points based on type
	for _, dp := range dataPoints {
		dpMap, ok := dp.(map[string]interface{})
		if !ok {
			continue
		}

		switch metricType {
		case "gauge":
			err = InsertGaugeDataPoint(tx, dpMap, metricID)
		case "sum":
			err = InsertSumDataPoint(tx, dpMap, metricID, metric)
		case "histogram":
			err = InsertHistogramDataPoint(tx, dpMap, metricID)
		case "exponential_histogram":
			err = InsertExponentialHistogramDataPoint(tx, dpMap, metricID)
		case "summary":
			err = InsertSummaryDataPoint(tx, dpMap, metricID)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

// InsertGaugeDataPoint inserts a gauge data point
func InsertGaugeDataPoint(tx *sql.Tx, dp map[string]interface{}, metricID int64) error {
	attributes := getOrDefault(dp, "attributes", []interface{}{})
	attributesJSON, _ := json.Marshal(attributes)
	
	startTime := parseTimeNano(dp["startTimeUnixNano"])
	timeNano := parseTimeNano(dp["timeUnixNano"])
	flags := int64(0)
	if f, ok := dp["flags"].(float64); ok {
		flags = int64(f)
	}

	var valueDouble sql.NullFloat64
	var valueInt sql.NullInt64

	if val, ok := dp["asDouble"].(float64); ok {
		valueDouble = sql.NullFloat64{Float64: val, Valid: true}
	} else if val, ok := dp["asInt"].(string); ok {
		if intVal, err := parseIntValue(val); err == nil {
			valueInt = sql.NullInt64{Int64: intVal, Valid: true}
		}
	}

	_, err := tx.Exec(`
		INSERT INTO metric_data_points (
			metric_id, attributes, start_time_unix_nano, time_unix_nano, flags,
			value_double, value_int
		) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		metricID, string(attributesJSON), startTime, timeNano, flags,
		valueDouble, valueInt,
	)
	return err
}

// InsertSumDataPoint inserts a sum data point
func InsertSumDataPoint(tx *sql.Tx, dp map[string]interface{}, metricID int64, metric map[string]interface{}) error {
	attributes := getOrDefault(dp, "attributes", []interface{}{})
	attributesJSON, _ := json.Marshal(attributes)
	
	startTime := parseTimeNano(dp["startTimeUnixNano"])
	timeNano := parseTimeNano(dp["timeUnixNano"])
	flags := int64(0)
	if f, ok := dp["flags"].(float64); ok {
		flags = int64(f)
	}

	var valueDouble sql.NullFloat64
	var valueInt sql.NullInt64

	if val, ok := dp["asDouble"].(float64); ok {
		valueDouble = sql.NullFloat64{Float64: val, Valid: true}
	} else if val, ok := dp["asInt"].(string); ok {
		if intVal, err := parseIntValue(val); err == nil {
			valueInt = sql.NullInt64{Int64: intVal, Valid: true}
		}
	}

	// Get sum-specific fields
	sum := metric["sum"].(map[string]interface{})
	aggregationTemporality := int64(0)
	if at, ok := sum["aggregationTemporality"].(float64); ok {
		aggregationTemporality = int64(at)
	}
	isMonotonic, _ := sum["isMonotonic"].(bool)

	_, err := tx.Exec(`
		INSERT INTO metric_data_points (
			metric_id, attributes, start_time_unix_nano, time_unix_nano, flags,
			value_double, value_int, aggregation_temporality, is_monotonic
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		metricID, string(attributesJSON), startTime, timeNano, flags,
		valueDouble, valueInt, aggregationTemporality, isMonotonic,
	)
	return err
}

// InsertHistogramDataPoint inserts a histogram data point
func InsertHistogramDataPoint(tx *sql.Tx, dp map[string]interface{}, metricID int64) error {
	attributes := getOrDefault(dp, "attributes", []interface{}{})
	attributesJSON, _ := json.Marshal(attributes)
	
	startTime := parseTimeNano(dp["startTimeUnixNano"])
	timeNano := parseTimeNano(dp["timeUnixNano"])
	flags := int64(0)
	if f, ok := dp["flags"].(float64); ok {
		flags = int64(f)
	}

	count := int64(0)
	if c, ok := dp["count"].(string); ok {
		count, _ = parseIntValue(c)
	}

	var sumValue sql.NullFloat64
	if s, ok := dp["sum"].(float64); ok {
		sumValue = sql.NullFloat64{Float64: s, Valid: true}
	}

	var minValue, maxValue sql.NullFloat64
	if min, ok := dp["min"].(float64); ok {
		minValue = sql.NullFloat64{Float64: min, Valid: true}
	}
	if max, ok := dp["max"].(float64); ok {
		maxValue = sql.NullFloat64{Float64: max, Valid: true}
	}

	bucketCountsJSON, _ := json.Marshal(dp["bucketCounts"])
	explicitBoundsJSON, _ := json.Marshal(dp["explicitBounds"])

	_, err := tx.Exec(`
		INSERT INTO metric_data_points (
			metric_id, attributes, start_time_unix_nano, time_unix_nano, flags,
			count, sum_value, min_value, max_value, bucket_counts, explicit_bounds
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		metricID, string(attributesJSON), startTime, timeNano, flags,
		count, sumValue, minValue, maxValue, string(bucketCountsJSON), string(explicitBoundsJSON),
	)
	return err
}

// InsertExponentialHistogramDataPoint inserts an exponential histogram data point
func InsertExponentialHistogramDataPoint(tx *sql.Tx, dp map[string]interface{}, metricID int64) error {
	attributes := getOrDefault(dp, "attributes", []interface{}{})
	attributesJSON, _ := json.Marshal(attributes)
	
	startTime := parseTimeNano(dp["startTimeUnixNano"])
	timeNano := parseTimeNano(dp["timeUnixNano"])
	flags := int64(0)
	if f, ok := dp["flags"].(float64); ok {
		flags = int64(f)
	}

	count := int64(0)
	if c, ok := dp["count"].(string); ok {
		count, _ = parseIntValue(c)
	}

	var sumValue sql.NullFloat64
	if s, ok := dp["sum"].(float64); ok {
		sumValue = sql.NullFloat64{Float64: s, Valid: true}
	}

	scale := int64(0)
	if s, ok := dp["scale"].(float64); ok {
		scale = int64(s)
	}

	zeroCount := int64(0)
	if zc, ok := dp["zeroCount"].(string); ok {
		zeroCount, _ = parseIntValue(zc)
	}

	// Handle positive and negative buckets
	var positiveOffset, negativeOffset int64
	var positiveBucketCounts, negativeBucketCounts string

	if positive, ok := dp["positive"].(map[string]interface{}); ok {
		if po, ok := positive["offset"].(float64); ok {
			positiveOffset = int64(po)
		}
		posCountsJSON, _ := json.Marshal(positive["bucketCounts"])
		positiveBucketCounts = string(posCountsJSON)
	}

	if negative, ok := dp["negative"].(map[string]interface{}); ok {
		if no, ok := negative["offset"].(float64); ok {
			negativeOffset = int64(no)
		}
		negCountsJSON, _ := json.Marshal(negative["bucketCounts"])
		negativeBucketCounts = string(negCountsJSON)
	}

	_, err := tx.Exec(`
		INSERT INTO metric_data_points (
			metric_id, attributes, start_time_unix_nano, time_unix_nano, flags,
			count, sum_value, scale, zero_count,
			positive_offset, positive_bucket_counts,
			negative_offset, negative_bucket_counts
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		metricID, string(attributesJSON), startTime, timeNano, flags,
		count, sumValue, scale, zeroCount,
		positiveOffset, positiveBucketCounts,
		negativeOffset, negativeBucketCounts,
	)
	return err
}

// InsertSummaryDataPoint inserts a summary data point
func InsertSummaryDataPoint(tx *sql.Tx, dp map[string]interface{}, metricID int64) error {
	attributes := getOrDefault(dp, "attributes", []interface{}{})
	attributesJSON, _ := json.Marshal(attributes)
	
	startTime := parseTimeNano(dp["startTimeUnixNano"])
	timeNano := parseTimeNano(dp["timeUnixNano"])
	flags := int64(0)
	if f, ok := dp["flags"].(float64); ok {
		flags = int64(f)
	}

	count := int64(0)
	if c, ok := dp["count"].(string); ok {
		count, _ = parseIntValue(c)
	}

	var sumValue sql.NullFloat64
	if s, ok := dp["sum"].(float64); ok {
		sumValue = sql.NullFloat64{Float64: s, Valid: true}
	}

	quantileValuesJSON, _ := json.Marshal(dp["quantileValues"])

	_, err := tx.Exec(`
		INSERT INTO metric_data_points (
			metric_id, attributes, start_time_unix_nano, time_unix_nano, flags,
			count, sum_value, quantile_values
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		metricID, string(attributesJSON), startTime, timeNano, flags,
		count, sumValue, string(quantileValuesJSON),
	)
	return err
}

// parseIntValue parses integer values from string or float64
func parseIntValue(val interface{}) (int64, error) {
	switch v := val.(type) {
	case string:
		var intVal int64
		_, err := fmt.Sscanf(v, "%d", &intVal)
		return intVal, err
	case float64:
		return int64(v), nil
	default:
		return 0, fmt.Errorf("unsupported type for int value")
	}
}