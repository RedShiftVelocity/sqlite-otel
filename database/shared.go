package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// GetOrCreateResource finds or creates a resource and returns its ID
func GetOrCreateResource(tx *sql.Tx, resource map[string]interface{}) (int64, error) {
	attributes, _ := resource["attributes"]
	
	var schemaURL string
	if su, ok := resource["schemaUrl"]; ok && su != nil {
		if s, ok := su.(string); ok {
			schemaURL = s
		} else {
			return 0, fmt.Errorf("resource schemaUrl has invalid type: %T", su)
		}
	}
	
	attributesJSON, err := json.Marshal(attributes)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal resource attributes: %w", err)
	}

	// Try to find existing resource
	var id int64
	var schemaURLValue interface{}
	if schemaURL != "" {
		schemaURLValue = schemaURL
	}
	
	err = tx.QueryRow(
		"SELECT id FROM resources WHERE attributes = ? AND schema_url IS ?",
		string(attributesJSON), schemaURLValue,
	).Scan(&id)
	
	if err == nil {
		return id, nil // Found existing resource
	}
	
	if err != sql.ErrNoRows {
		return 0, fmt.Errorf("failed to query resource: %w", err)
	}

	// Create new resource
	result, err := tx.Exec(
		"INSERT INTO resources (attributes, schema_url) VALUES (?, ?)",
		string(attributesJSON), schemaURLValue,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert resource: %w", err)
	}

	return result.LastInsertId()
}

// GetOrCreateScope finds or creates an instrumentation scope and returns its ID
func GetOrCreateScope(tx *sql.Tx, scope map[string]interface{}) (int64, error) {
	var name, version, schemaURL string
	
	// Safe extraction of name
	if n, ok := scope["name"]; ok && n != nil {
		if s, ok := n.(string); ok {
			name = s
		} else {
			return 0, fmt.Errorf("scope name has invalid type: %T", n)
		}
	}
	
	// Safe extraction of version
	if v, ok := scope["version"]; ok && v != nil {
		if s, ok := v.(string); ok {
			version = s
		} else {
			return 0, fmt.Errorf("scope version has invalid type: %T", v)
		}
	}
	
	// Safe extraction of schemaUrl
	if su, ok := scope["schemaUrl"]; ok && su != nil {
		if s, ok := su.(string); ok {
			schemaURL = s
		} else {
			return 0, fmt.Errorf("scope schemaUrl has invalid type: %T", su)
		}
	}
	
	attributes, _ := scope["attributes"]

	attributesJSON, err := json.Marshal(attributes)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal scope attributes: %w", err)
	}

	// Try to find existing scope
	var id int64
	var versionValue, schemaURLValue interface{}
	if version != "" {
		versionValue = version
	}
	if schemaURL != "" {
		schemaURLValue = schemaURL
	}
	
	err = tx.QueryRow(
		`SELECT id FROM instrumentation_scopes 
		WHERE name = ? AND version IS ? 
		AND attributes = ? AND schema_url IS ?`,
		name, versionValue, string(attributesJSON), schemaURLValue,
	).Scan(&id)
	
	if err == nil {
		return id, nil // Found existing scope
	}
	
	if err != sql.ErrNoRows {
		return 0, fmt.Errorf("failed to query scope: %w", err)
	}

	// Create new scope
	result, err := tx.Exec(
		"INSERT INTO instrumentation_scopes (name, version, attributes, schema_url) VALUES (?, ?, ?, ?)",
		name, versionValue, string(attributesJSON), schemaURLValue,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert scope: %w", err)
	}

	return result.LastInsertId()
}

// parseTimeNano converts a time string to Unix nanoseconds
func parseTimeNano(timeStr string) (int64, error) {
	if timeStr == "" {
		return 0, nil // Empty timestamp is not an error
	}
	t, err := time.Parse(time.RFC3339Nano, timeStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse time '%s': %w", timeStr, err)
	}
	return t.UnixNano(), nil
}

// GetOrCreateMetric finds or creates a metric and returns its ID
func GetOrCreateMetric(tx *sql.Tx, name, description, unit, metricType string, resourceID, scopeID int64) (int64, error) {
	// Try to find existing metric
	var id int64
	err := tx.QueryRow(
		`SELECT id FROM metrics 
		WHERE name = ? AND type = ? AND resource_id = ? AND scope_id = ?`,
		name, metricType, resourceID, scopeID,
	).Scan(&id)
	
	if err == nil {
		return id, nil // Found existing metric
	}
	
	if err != sql.ErrNoRows {
		return 0, fmt.Errorf("failed to query metric: %w", err)
	}

	// Create new metric
	result, err := tx.Exec(
		"INSERT INTO metrics (name, description, unit, type, resource_id, scope_id) VALUES (?, ?, ?, ?, ?, ?)",
		name, description, unit, metricType, resourceID, scopeID,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert metric: %w", err)
	}

	return result.LastInsertId()
}

// getOrDefault returns the value if it exists, otherwise returns the default
func getOrDefault(data map[string]interface{}, key string, defaultValue interface{}) interface{} {
	if val, ok := data[key]; ok {
		return val
	}
	return defaultValue
}