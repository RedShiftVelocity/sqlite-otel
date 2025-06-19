package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// getStringFromMap safely extracts a string value from a map[string]interface{}
func getStringFromMap(m map[string]interface{}, key string) (string, error) {
	var val string
	if v, ok := m[key]; ok && v != nil {
		if s, ok := v.(string); ok {
			val = s
		} else {
			return "", fmt.Errorf("key '%s' has invalid type: %T", key, v)
		}
	}
	return val, nil
}

// GetOrCreateResource finds or creates a resource and returns its ID
func GetOrCreateResource(tx *sql.Tx, resource map[string]interface{}) (int64, error) {
	// Explicitly handle attributes extraction
	attributes, ok := resource["attributes"]
	if !ok || attributes == nil {
		attributes = make(map[string]interface{}) // Default to empty map
	}
	
	schemaURL, err := getStringFromMap(resource, "schemaUrl")
	if err != nil {
		return 0, fmt.Errorf("resource %w", err)
	}
	
	// Marshal attributes to a canonical JSON string.
	// NOTE: Go's standard json.Marshal sorts map keys, which is essential
	// for the UNIQUE index on the attributes column to work correctly.
	attributesJSON, err := json.Marshal(attributes)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal resource attributes: %w", err)
	}

	// Use atomic INSERT ... ON CONFLICT DO NOTHING for compatibility with older SQLite
	// This approach works with SQLite 3.24.0+ (ON CONFLICT requires 3.24.0+)
	_, err = tx.Exec(`
		INSERT INTO resources (attributes, schema_url) VALUES (?, ?)
		ON CONFLICT(attributes, schema_url) DO NOTHING`,
		string(attributesJSON), schemaURL,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert resource: %w", err)
	}
	
	// Now select the ID of the guaranteed-to-exist row
	var id int64
	err = tx.QueryRow(`
		SELECT id FROM resources WHERE attributes = ? AND schema_url = ?`,
		string(attributesJSON), schemaURL,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to get resource id: %w", err)
	}
	
	return id, nil
}

// GetOrCreateScope finds or creates an instrumentation scope and returns its ID
func GetOrCreateScope(tx *sql.Tx, scope map[string]interface{}) (int64, error) {
	name, err := getStringFromMap(scope, "name")
	if err != nil {
		return 0, fmt.Errorf("scope %w", err)
	}
	version, err := getStringFromMap(scope, "version")
	if err != nil {
		return 0, fmt.Errorf("scope %w", err)
	}
	schemaURL, err := getStringFromMap(scope, "schemaUrl")
	if err != nil {
		return 0, fmt.Errorf("scope %w", err)
	}
	
	// Explicitly handle attributes extraction
	attributes, ok := scope["attributes"]
	if !ok || attributes == nil {
		attributes = make(map[string]interface{}) // Default to empty map
	}

	// Marshal attributes to a canonical JSON string.
	// NOTE: Go's standard json.Marshal sorts map keys, which is essential
	// for the UNIQUE index on the attributes column to work correctly.
	attributesJSON, err := json.Marshal(attributes)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal scope attributes: %w", err)
	}

	// Use atomic INSERT ... ON CONFLICT DO NOTHING for compatibility with older SQLite
	_, err = tx.Exec(`
		INSERT INTO instrumentation_scopes (name, version, attributes, schema_url) VALUES (?, ?, ?, ?)
		ON CONFLICT(name, version, attributes, schema_url) DO NOTHING`,
		name, version, string(attributesJSON), schemaURL,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert scope: %w", err)
	}
	
	// Now select the ID of the guaranteed-to-exist row
	var id int64
	err = tx.QueryRow(`
		SELECT id FROM instrumentation_scopes 
		WHERE name = ? AND version = ? AND attributes = ? AND schema_url = ?`,
		name, version, string(attributesJSON), schemaURL,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to get scope id: %w", err)
	}
	
	return id, nil
}

// parseTimeNano converts OTLP timestamp (string-encoded nanoseconds) to int64
func parseTimeNano(timeStr string) (int64, error) {
	// Trim whitespace for robustness
	timeStr = strings.TrimSpace(timeStr)
	if timeStr == "" {
		return 0, nil // Empty timestamp is not an error
	}
	// OTLP JSON sends timestamps as string-encoded nanoseconds since Unix epoch
	// e.g., "1672531200000000000" for 2023-01-01 00:00:00 UTC
	val, err := strconv.ParseInt(timeStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse timestamp '%s': %w", timeStr, err)
	}
	return val, nil
}

// GetOrCreateMetric finds or creates a metric and returns its ID
func GetOrCreateMetric(tx *sql.Tx, name, description, unit, metricType string, resourceID, scopeID int64) (int64, error) {
	// Use atomic INSERT ... ON CONFLICT DO NOTHING for compatibility with older SQLite
	// We don't update description/unit on conflict to maintain consistency - first definition wins
	_, err := tx.Exec(`
		INSERT INTO metrics (name, description, unit, metric_type, resource_id, scope_id) VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(name, metric_type, resource_id, scope_id) DO NOTHING`,
		name, description, unit, metricType, resourceID, scopeID,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert metric: %w", err)
	}
	
	// Now select the ID of the guaranteed-to-exist row
	var id int64
	err = tx.QueryRow(`
		SELECT id FROM metrics 
		WHERE name = ? AND metric_type = ? AND resource_id = ? AND scope_id = ?`,
		name, metricType, resourceID, scopeID,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to get metric id: %w", err)
	}
	
	return id, nil
}