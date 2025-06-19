#!/bin/bash

# Test histogram metric
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
        ],
        "schemaUrl": "https://opentelemetry.io/schemas/1.4.0"
      },
      "scopeMetrics": [
        {
          "scope": {
            "name": "test-meter",
            "version": "1.0.0"
          },
          "metrics": [
            {
              "name": "http.request.duration",
              "description": "HTTP request duration",
              "unit": "ms",
              "histogram": {
                "dataPoints": [
                  {
                    "attributes": [
                      {
                        "key": "http.method",
                        "value": {
                          "stringValue": "POST"
                        }
                      }
                    ],
                    "startTimeUnixNano": "2024-01-01T12:00:00.000000000Z",
                    "timeUnixNano": "2024-01-01T12:00:10.000000000Z",
                    "count": "100",
                    "sum": 5000.5,
                    "bucketCounts": ["10", "20", "30", "25", "15"],
                    "explicitBounds": [10, 50, 100, 200, 500],
                    "exemplars": [
                      {
                        "timeUnixNano": "2024-01-01T12:00:05.000000000Z",
                        "asDouble": 45.2,
                        "traceId": "5b8efff798038103d269b633813fc60c",
                        "spanId": "eee19b7ec3c1b173"
                      }
                    ]
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