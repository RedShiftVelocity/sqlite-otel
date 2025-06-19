#!/bin/bash

# Test 1: Basic log with all fields
echo "Test 1: Basic log with all fields"
curl -X POST http://localhost:4321/v1/logs \
  -H "Content-Type: application/json" \
  -d '{
  "resourceLogs": [
    {
      "resource": {
        "attributes": [
          {
            "key": "service.name",
            "value": {
              "stringValue": "test-service"
            }
          }
        ]
      },
      "scopeLogs": [
        {
          "scope": {
            "name": "test-logger",
            "version": "1.0.0"
          },
          "logRecords": [
            {
              "timeUnixNano": "2024-01-01T12:00:00.000000000Z",
              "observedTimeUnixNano": "2024-01-01T12:00:01.000000000Z",
              "severityNumber": 9,
              "severityText": "INFO",
              "body": {
                "stringValue": "Test log message"
              },
              "attributes": [
                {
                  "key": "user.id",
                  "value": {
                    "stringValue": "user123"
                  }
                }
              ],
              "traceId": "5b8efff798038103d269b633813fc60c",
              "spanId": "eee19b7ec3c1b173",
              "flags": 1
            }
          ]
        }
      ]
    }
  ]
}'
echo

# Test 2: Log with minimal fields
echo "Test 2: Log with minimal fields (no trace context, no attributes)"
curl -X POST http://localhost:4321/v1/logs \
  -H "Content-Type: application/json" \
  -d '{
  "resourceLogs": [
    {
      "resource": {
        "attributes": []
      },
      "scopeLogs": [
        {
          "scope": {
            "name": "minimal-logger"
          },
          "logRecords": [
            {
              "body": {
                "stringValue": "Minimal log message"
              }
            }
          ]
        }
      ]
    }
  ]
}'
echo

# Test 3: Log with complex body (structured)
echo "Test 3: Log with complex structured body"
curl -X POST http://localhost:4321/v1/logs \
  -H "Content-Type: application/json" \
  -d '{
  "resourceLogs": [
    {
      "resource": {
        "attributes": [
          {
            "key": "service.name",
            "value": {
              "stringValue": "test-service"
            }
          }
        ]
      },
      "scopeLogs": [
        {
          "scope": {
            "name": "structured-logger",
            "version": "2.0.0"
          },
          "logRecords": [
            {
              "timeUnixNano": "2024-01-01T12:00:05.000000000Z",
              "severityNumber": 13,
              "severityText": "ERROR",
              "body": {
                "kvlistValue": {
                  "values": [
                    {
                      "key": "error",
                      "value": {
                        "stringValue": "Connection timeout"
                      }
                    },
                    {
                      "key": "retry_count",
                      "value": {
                        "intValue": "3"
                      }
                    }
                  ]
                }
              },
              "attributes": [
                {
                  "key": "component",
                  "value": {
                    "stringValue": "database"
                  }
                }
              ]
            }
          ]
        }
      ]
    }
  ]
}'
echo

# Test 4: Invalid timestamp
echo "Test 4: Log with invalid timestamp (should fail)"
curl -X POST http://localhost:4321/v1/logs \
  -H "Content-Type: application/json" \
  -d '{
  "resourceLogs": [
    {
      "resource": {
        "attributes": []
      },
      "scopeLogs": [
        {
          "scope": {
            "name": "test-logger"
          },
          "logRecords": [
            {
              "timeUnixNano": "invalid-timestamp",
              "body": {
                "stringValue": "This should fail"
              }
            }
          ]
        }
      ]
    }
  ]
}'
echo

# Test 5: Invalid severityText type
echo "Test 5: Log with invalid severityText type (should fail)"
curl -X POST http://localhost:4321/v1/logs \
  -H "Content-Type: application/json" \
  -d '{
  "resourceLogs": [
    {
      "resource": {
        "attributes": []
      },
      "scopeLogs": [
        {
          "scope": {
            "name": "test-logger"
          },
          "logRecords": [
            {
              "severityText": 123,
              "body": {
                "stringValue": "This should fail"
              }
            }
          ]
        }
      ]
    }
  ]
}'
echo

# Give collector time to process
sleep 1

echo "Checking database for stored logs..."