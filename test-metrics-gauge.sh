#!/bin/bash

# Test gauge metric
curl -X POST http://localhost:4320/v1/metrics \
  -H "Content-Type: application/json" \
  -d '{
  "resourceMetrics": [
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
      "scopeMetrics": [
        {
          "scope": {
            "name": "test-meter",
            "version": "1.0.0"
          },
          "metrics": [
            {
              "name": "cpu.utilization",
              "description": "CPU utilization percentage",
              "unit": "%",
              "gauge": {
                "dataPoints": [
                  {
                    "attributes": [
                      {
                        "key": "cpu",
                        "value": {
                          "stringValue": "cpu0"
                        }
                      }
                    ],
                    "startTimeUnixNano": "2024-01-01T12:00:00.000000000Z",
                    "timeUnixNano": "2024-01-01T12:00:10.000000000Z",
                    "asDouble": 45.3
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