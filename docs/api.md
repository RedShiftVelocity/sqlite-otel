# API Reference

The SQLite OTEL Collector implements the OpenTelemetry Protocol (OTLP) over HTTP for receiving telemetry data.

## Base URL

```
http://localhost:4318
```

## Endpoints

### Health Check

#### `GET /health`

Health check endpoint to verify the collector is running.

**Response:**
```http
HTTP/1.1 200 OK
Content-Type: text/plain

OK
```

**Example:**
```bash
curl http://localhost:4318/health
```

### Traces

#### `POST /v1/traces`

Accepts trace data in OTLP format.

**Content-Type:** `application/json` or `application/x-protobuf`

**Request Body:** OTLP TracesData

**Example Request:**
```bash
curl -X POST http://localhost:4318/v1/traces \
  -H "Content-Type: application/json" \
  -d '{
    "resourceSpans": [{
      "resource": {
        "attributes": [
          {
            "key": "service.name",
            "value": {"stringValue": "my-service"}
          }
        ]
      },
      "scopeSpans": [{
        "spans": [{
          "traceId": "0123456789abcdef0123456789abcdef",
          "spanId": "0123456789abcdef", 
          "name": "my-operation",
          "kind": 1,
          "startTimeUnixNano": "1640995200000000000",
          "endTimeUnixNano": "1640995201000000000",
          "attributes": [
            {
              "key": "http.method",
              "value": {"stringValue": "GET"}
            }
          ]
        }]
      }]
    }]
  }'
```

**Response:**
```http
HTTP/1.1 200 OK
```

### Metrics

#### `POST /v1/metrics`

Accepts metrics data in OTLP format.

**Content-Type:** `application/json` or `application/x-protobuf`

**Request Body:** OTLP MetricsData

**Example Request:**
```bash
curl -X POST http://localhost:4318/v1/metrics \
  -H "Content-Type: application/json" \
  -d '{
    "resourceMetrics": [{
      "resource": {
        "attributes": [
          {
            "key": "service.name",
            "value": {"stringValue": "my-service"}
          }
        ]
      },
      "scopeMetrics": [{
        "metrics": [{
          "name": "requests_total",
          "description": "Total number of requests",
          "unit": "1",
          "sum": {
            "dataPoints": [{
              "attributes": [
                {
                  "key": "status",
                  "value": {"stringValue": "200"}
                }
              ],
              "asInt": "42",
              "timeUnixNano": "1640995200000000000"
            }],
            "aggregationTemporality": 2,
            "isMonotonic": true
          }
        }]
      }]
    }]
  }'
```

**Response:**
```http
HTTP/1.1 200 OK
```

### Logs

#### `POST /v1/logs`

Accepts log data in OTLP format.

**Content-Type:** `application/json` or `application/x-protobuf`

**Request Body:** OTLP LogsData

**Example Request:**
```bash
curl -X POST http://localhost:4318/v1/logs \
  -H "Content-Type: application/json" \
  -d '{
    "resourceLogs": [{
      "resource": {
        "attributes": [
          {
            "key": "service.name",
            "value": {"stringValue": "my-service"}
          }
        ]
      },
      "scopeLogs": [{
        "logRecords": [{
          "timeUnixNano": "1640995200000000000",
          "severityNumber": 9,
          "severityText": "INFO",
          "body": {
            "stringValue": "User logged in successfully"
          },
          "attributes": [
            {
              "key": "user.id",
              "value": {"stringValue": "12345"}
            }
          ]
        }]
      }]
    }]
  }'
```

**Response:**
```http
HTTP/1.1 200 OK
```

## Data Storage

All received telemetry data is stored in SQLite with the following schema:

### Spans Table

```sql
CREATE TABLE spans (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  trace_id TEXT NOT NULL,
  span_id TEXT NOT NULL,
  parent_span_id TEXT,
  name TEXT NOT NULL,
  kind INTEGER,
  start_time INTEGER,
  end_time INTEGER,
  duration INTEGER,
  service_name TEXT,
  service_version TEXT,
  status_code INTEGER,
  status_message TEXT,
  attributes TEXT, -- JSON
  events TEXT,     -- JSON
  links TEXT,      -- JSON
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  
  UNIQUE(trace_id, span_id)
);
```

### Metrics Table

```sql
CREATE TABLE metrics (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  description TEXT,
  unit TEXT,
  type TEXT, -- gauge, counter, histogram, summary
  value REAL,
  value_int INTEGER,
  timestamp INTEGER,
  service_name TEXT,
  service_version TEXT,
  attributes TEXT, -- JSON
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### Logs Table

```sql
CREATE TABLE logs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  timestamp INTEGER,
  severity_number INTEGER,
  severity_text TEXT,
  body TEXT,
  service_name TEXT,
  service_version TEXT,
  trace_id TEXT,
  span_id TEXT,
  attributes TEXT, -- JSON
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

## Error Responses

### 400 Bad Request
Invalid request format or malformed OTLP data.

```json
{
  "error": "Invalid OTLP data format",
  "details": "Missing required field: resourceSpans"
}
```

### 413 Payload Too Large
Request body exceeds maximum size limit.

```json
{
  "error": "Request too large",
  "details": "Maximum request size is 10MB"
}
```

### 500 Internal Server Error
Database or server error.

```json
{
  "error": "Internal server error",
  "details": "Database write failed"
}
```

## Rate Limiting

Currently, no rate limiting is implemented. For production deployments, consider:

- Using a reverse proxy (nginx, Apache) with rate limiting
- Implementing network-level rate limiting
- Monitoring resource usage

## Authentication

The collector currently does not implement authentication. For production:

- Deploy behind a reverse proxy with authentication
- Use network-level security (VPN, private networks)
- Implement firewall rules

## SDK Configuration

Configure your OpenTelemetry SDKs to send data to the collector:

### Environment Variables

```bash
export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
export OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf
```

### Go SDK

```go
import (
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
    "go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
)

// Trace exporter
traceExporter, err := otlptracehttp.New(
    context.Background(),
    otlptracehttp.WithEndpoint("http://localhost:4318"),
    otlptracehttp.WithInsecure(),
)

// Metric exporter  
metricExporter, err := otlpmetrichttp.New(
    context.Background(),
    otlpmetrichttp.WithEndpoint("http://localhost:4318"),
    otlpmetrichttp.WithInsecure(),
)
```

### Python SDK

```python
from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter
from opentelemetry.exporter.otlp.proto.http.metric_exporter import OTLPMetricExporter

# Trace exporter
trace_exporter = OTLPSpanExporter(
    endpoint="http://localhost:4318/v1/traces"
)

# Metric exporter
metric_exporter = OTLPMetricExporter(
    endpoint="http://localhost:4318/v1/metrics"
)
```

### Java SDK

```java
import io.opentelemetry.exporter.otlp.http.trace.OtlpHttpSpanExporter;
import io.opentelemetry.exporter.otlp.http.metrics.OtlpHttpMetricExporter;

// Trace exporter
OtlpHttpSpanExporter traceExporter = OtlpHttpSpanExporter.builder()
    .setEndpoint("http://localhost:4318/v1/traces")
    .build();

// Metric exporter
OtlpHttpMetricExporter metricExporter = OtlpHttpMetricExporter.builder()
    .setEndpoint("http://localhost:4318/v1/metrics")
    .build();
```

### Node.js SDK

```javascript
const { OTLPTraceExporter } = require('@opentelemetry/exporter-otlp-http');
const { OTLPMetricExporter } = require('@opentelemetry/exporter-otlp-http');

// Trace exporter
const traceExporter = new OTLPTraceExporter({
  url: 'http://localhost:4318/v1/traces',
});

// Metric exporter
const metricExporter = new OTLPMetricExporter({
  url: 'http://localhost:4318/v1/metrics',
});
```

## Query Examples

Access the SQLite database directly to query your data:

### Recent Traces
```sql
SELECT 
  trace_id,
  name,
  service_name,
  duration,
  datetime(start_time/1000000000, 'unixepoch') as start_time
FROM spans 
WHERE start_time > (strftime('%s', 'now') - 3600) * 1000000000
ORDER BY start_time DESC
LIMIT 20;
```

### Error Spans
```sql
SELECT 
  trace_id,
  span_id,
  name,
  service_name,
  status_message
FROM spans 
WHERE status_code > 0
ORDER BY start_time DESC;
```

### Metrics by Service
```sql
SELECT 
  service_name,
  name,
  AVG(value) as avg_value,
  COUNT(*) as count
FROM metrics 
WHERE timestamp > (strftime('%s', 'now') - 3600) * 1000000000
GROUP BY service_name, name
ORDER BY service_name, name;
```

## See Also

- [OpenTelemetry Protocol Specification](https://github.com/open-telemetry/opentelemetry-proto)
- [Quick Start Guide](quickstart.md) - Getting started
- [Configuration](configuration.md) - Collector configuration