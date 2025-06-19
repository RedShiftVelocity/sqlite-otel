#!/bin/bash

# Test sum metric with integer value
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
              "name": "http.requests.total",
              "description": "Total HTTP requests",
              "unit": "1",
              "sum": {
                "dataPoints": [
                  {
                    "attributes": [
                      {
                        "key": "method",
                        "value": {
                          "stringValue": "GET"
                        }
                      }
                    ],
                    "startTimeUnixNano": "2024-01-01T12:00:00.000000000Z",
                    "timeUnixNano": "2024-01-01T12:00:10.000000000Z",
                    "asInt": "12345"
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