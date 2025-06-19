#!/bin/bash

# Test various error cases

echo "Test 1: Missing resource (should fail)"
curl -X POST http://localhost:4320/v1/metrics \
  -H "Content-Type: application/json" \
  -d '{
  "resourceMetrics": [
    {
      "scopeMetrics": [
        {
          "scope": {
            "name": "test-meter"
          },
          "metrics": [
            {
              "name": "test.metric",
              "gauge": {
                "dataPoints": [
                  {
                    "asDouble": 1.0
                  }
                ]
              }
            }
          ]
        }
      ]
    }
  ]
}'

echo -e "\n\nTest 2: Invalid timestamp (should fail)"
curl -X POST http://localhost:4320/v1/metrics \
  -H "Content-Type: application/json" \
  -d '{
  "resourceMetrics": [
    {
      "resource": {
        "attributes": []
      },
      "scopeMetrics": [
        {
          "scope": {
            "name": "test-meter"
          },
          "metrics": [
            {
              "name": "test.metric",
              "gauge": {
                "dataPoints": [
                  {
                    "timeUnixNano": "invalid-timestamp",
                    "asDouble": 1.0
                  }
                ]
              }
            }
          ]
        }
      ]
    }
  ]
}'

echo -e "\n\nTest 3: Invalid integer value (should fail)"
curl -X POST http://localhost:4320/v1/metrics \
  -H "Content-Type: application/json" \
  -d '{
  "resourceMetrics": [
    {
      "resource": {
        "attributes": []
      },
      "scopeMetrics": [
        {
          "scope": {
            "name": "test-meter"
          },
          "metrics": [
            {
              "name": "test.counter",
              "sum": {
                "dataPoints": [
                  {
                    "asInt": "not-a-number"
                  }
                ]
              }
            }
          ]
        }
      ]
    }
  ]
}'