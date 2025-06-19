package database

import (
	"strconv"
)

// Helper to safely get values from map
func getOrDefault(m map[string]interface{}, key string, defaultValue interface{}) interface{} {
	if val, ok := m[key]; ok {
		return val
	}
	return defaultValue
}

// Helper function to parse time from various formats
func parseTimeNano(timeValue interface{}) int64 {
	switch v := timeValue.(type) {
	case string:
		if t, err := strconv.ParseInt(v, 10, 64); err == nil {
			return t
		}
	case float64:
		return int64(v)
	}
	return 0
}