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

	// Use atomic UPSERT with RETURNING clause to avoid race conditions
	var id int64
	var schemaURLValue interface{}
	if schemaURL != "" {
		schemaURLValue = schemaURL
	}
	
	// INSERT ... ON CONFLICT ... DO UPDATE SET id=id is a no-op update that allows RETURNING to work
	err = tx.QueryRow(`
		INSERT INTO resources (attributes, schema_url) VALUES (?, ?)
		ON CONFLICT(attributes, schema_url) DO UPDATE SET id=id
		RETURNING id`,
		string(attributesJSON), schemaURLValue,
	).Scan(&id)
	
	if err != nil {
		return 0, fmt.Errorf("failed to get or create resource: %w", err)
	}
	
	return id, nil
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

	// Use atomic UPSERT with RETURNING clause to avoid race conditions
	var id int64
	var versionValue, schemaURLValue interface{}
	if version != "" {
		versionValue = version
	}
	if schemaURL != "" {
		schemaURLValue = schemaURL
	}
	
	// INSERT ... ON CONFLICT ... DO UPDATE SET id=id is a no-op update that allows RETURNING to work
	err = tx.QueryRow(`
		INSERT INTO instrumentation_scopes (name, version, attributes, schema_url) VALUES (?, ?, ?, ?)
		ON CONFLICT(name, version, attributes, schema_url) DO UPDATE SET id=id
		RETURNING id`,
		name, versionValue, string(attributesJSON), schemaURLValue,
	).Scan(&id)
	
	if err != nil {
		return 0, fmt.Errorf("failed to get or create scope: %w", err)
	}
	
	return id, nil
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
	// Use atomic UPSERT with RETURNING clause to avoid race conditions
	var id int64
	
	// INSERT ... ON CONFLICT ... DO UPDATE SET id=id is a no-op update that allows RETURNING to work
	// Description and unit may change between calls, so we update them on conflict
	err := tx.QueryRow(`
		INSERT INTO metrics (name, description, unit, type, resource_id, scope_id) VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(name, type, resource_id, scope_id) DO UPDATE SET 
			description = excluded.description,
			unit = excluded.unit
		RETURNING id`,
		name, description, unit, metricType, resourceID, scopeID,
	).Scan(&id)
	
	if err != nil {
		return 0, fmt.Errorf("failed to get or create metric: %w", err)
	}
	
	return id, nil
}

// getOrDefault returns the value if it exists, otherwise returns the default
func getOrDefault(data map[string]interface{}, key string, defaultValue interface{}) interface{} {
	if val, ok := data[key]; ok {
		return val
	}
	return defaultValue
}