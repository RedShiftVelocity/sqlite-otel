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
	schemaURL, _ := resource["schemaUrl"].(string)
	
	attributesJSON, err := json.Marshal(attributes)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal resource attributes: %w", err)
	}

	// Try to find existing resource
	var id int64
	err = tx.QueryRow(
		"SELECT id FROM resources WHERE attributes = ? AND (schema_url = ? OR (schema_url IS NULL AND ? IS NULL))",
		string(attributesJSON), schemaURL, schemaURL,
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
		string(attributesJSON), schemaURL,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert resource: %w", err)
	}

	return result.LastInsertId()
}

// GetOrCreateScope finds or creates an instrumentation scope and returns its ID
func GetOrCreateScope(tx *sql.Tx, scope map[string]interface{}) (int64, error) {
	name, _ := scope["name"].(string)
	version, _ := scope["version"].(string)
	attributes, _ := scope["attributes"]
	schemaURL, _ := scope["schemaUrl"].(string)

	attributesJSON, err := json.Marshal(attributes)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal scope attributes: %w", err)
	}

	// Try to find existing scope
	var id int64
	err = tx.QueryRow(
		`SELECT id FROM instrumentation_scopes 
		WHERE name = ? AND (version = ? OR (version IS NULL AND ? IS NULL)) 
		AND attributes = ? AND (schema_url = ? OR (schema_url IS NULL AND ? IS NULL))`,
		name, version, version, string(attributesJSON), schemaURL, schemaURL,
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
		name, version, string(attributesJSON), schemaURL,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert scope: %w", err)
	}

	return result.LastInsertId()
}

// parseTimeNano converts a time string to Unix nanoseconds
func parseTimeNano(timeStr string) int64 {
	t, err := time.Parse(time.RFC3339Nano, timeStr)
	if err != nil {
		return 0
	}
	return t.UnixNano()
}

// getOrDefault returns the value if it exists, otherwise returns the default
func getOrDefault(data map[string]interface{}, key string, defaultValue interface{}) interface{} {
	if val, ok := data[key]; ok {
		return val
	}
	return defaultValue
}