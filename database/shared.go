package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// GetOrCreateResource finds or creates a resource and returns its ID
func GetOrCreateResource(tx *sql.Tx, resource map[string]interface{}) (int64, error) {
	// Explicitly handle attributes extraction
	attributes, ok := resource["attributes"]
	if !ok || attributes == nil {
		attributes = make(map[string]interface{}) // Default to empty map
	}
	
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
	
	// Explicitly handle attributes extraction
	attributes, ok := scope["attributes"]
	if !ok || attributes == nil {
		attributes = make(map[string]interface{}) // Default to empty map
	}

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
	// Use atomic INSERT ... ON CONFLICT DO NOTHING for compatibility with older SQLite
	// We don't update description/unit on conflict to maintain consistency - first definition wins
	_, err := tx.Exec(`
		INSERT INTO metrics (name, description, unit, type, resource_id, scope_id) VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(name, type, resource_id, scope_id) DO NOTHING`,
		name, description, unit, metricType, resourceID, scopeID,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert metric: %w", err)
	}
	
	// Now select the ID of the guaranteed-to-exist row
	var id int64
	err = tx.QueryRow(`
		SELECT id FROM metrics 
		WHERE name = ? AND type = ? AND resource_id = ? AND scope_id = ?`,
		name, metricType, resourceID, scopeID,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to get metric id: %w", err)
	}
	
	return id, nil
}