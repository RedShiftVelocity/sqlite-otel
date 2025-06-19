package database

import (
	"testing"
)

func TestInsertLogsData(t *testing.T) {
	// Initialize test database
	if err := InitDB(":memory:"); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer CloseDB()

	// Test 1: Valid log with all fields
	validLog := map[string]interface{}{
		"resourceLogs": []interface{}{
			map[string]interface{}{
				"resource": map[string]interface{}{
					"attributes": []interface{}{
						map[string]interface{}{
							"key": "service.name",
							"value": map[string]interface{}{
								"stringValue": "test-service",
							},
						},
					},
				},
				"scopeLogs": []interface{}{
					map[string]interface{}{
						"scope": map[string]interface{}{
							"name":    "test-logger",
							"version": "1.0.0",
						},
						"logRecords": []interface{}{
							map[string]interface{}{
								"timeUnixNano":         "2024-01-01T12:00:00.000000000Z",
								"observedTimeUnixNano": "2024-01-01T12:00:01.000000000Z",
								"severityNumber":       float64(9),
								"severityText":         "INFO",
								"body": map[string]interface{}{
									"stringValue": "Test log message",
								},
								"attributes": []interface{}{
									map[string]interface{}{
										"key": "user.id",
										"value": map[string]interface{}{
											"stringValue": "user123",
										},
									},
								},
								"traceId": "5b8efff798038103d269b633813fc60c",
								"spanId":  "eee19b7ec3c1b173",
								"flags":   float64(1),
							},
						},
					},
				},
			},
		},
	}

	if err := InsertLogsData(validLog); err != nil {
		t.Errorf("Failed to insert valid log: %v", err)
	}

	// Test 2: Minimal log (no trace context, no attributes)
	minimalLog := map[string]interface{}{
		"resourceLogs": []interface{}{
			map[string]interface{}{
				"resource": map[string]interface{}{
					"attributes": []interface{}{},
				},
				"scopeLogs": []interface{}{
					map[string]interface{}{
						"scope": map[string]interface{}{
							"name": "minimal-logger",
						},
						"logRecords": []interface{}{
							map[string]interface{}{
								"body": map[string]interface{}{
									"stringValue": "Minimal log message",
								},
							},
						},
					},
				},
			},
		},
	}

	if err := InsertLogsData(minimalLog); err != nil {
		t.Errorf("Failed to insert minimal log: %v", err)
	}

	// Test 3: Invalid timestamp
	invalidTimestampLog := map[string]interface{}{
		"resourceLogs": []interface{}{
			map[string]interface{}{
				"resource": map[string]interface{}{
					"attributes": []interface{}{},
				},
				"scopeLogs": []interface{}{
					map[string]interface{}{
						"scope": map[string]interface{}{
							"name": "test-logger",
						},
						"logRecords": []interface{}{
							map[string]interface{}{
								"timeUnixNano": "invalid-timestamp",
								"body": map[string]interface{}{
									"stringValue": "This should fail",
								},
							},
						},
					},
				},
			},
		},
	}

	if err := InsertLogsData(invalidTimestampLog); err == nil {
		t.Error("Expected error for invalid timestamp, but got none")
	}

	// Test 4: Invalid severityText type
	invalidSeverityLog := map[string]interface{}{
		"resourceLogs": []interface{}{
			map[string]interface{}{
				"resource": map[string]interface{}{
					"attributes": []interface{}{},
				},
				"scopeLogs": []interface{}{
					map[string]interface{}{
						"scope": map[string]interface{}{
							"name": "test-logger",
						},
						"logRecords": []interface{}{
							map[string]interface{}{
								"severityText": 123, // Should be string
								"body": map[string]interface{}{
									"stringValue": "This should fail",
								},
							},
						},
					},
				},
			},
		},
	}

	if err := InsertLogsData(invalidSeverityLog); err == nil {
		t.Error("Expected error for invalid severityText type, but got none")
	}

	// Test 5: Malformed resourceLog
	malformedLog := map[string]interface{}{
		"resourceLogs": []interface{}{
			"not a map", // Should be map[string]interface{}
		},
	}

	if err := InsertLogsData(malformedLog); err == nil {
		t.Error("Expected error for malformed resourceLog, but got none")
	}

	// Query to verify logs were inserted
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM log_records").Scan(&count); err != nil {
		t.Fatalf("Failed to query log records: %v", err)
	}

	// Should have 2 successful insertions
	if count != 2 {
		t.Errorf("Expected 2 log records, got %d", count)
	}
}