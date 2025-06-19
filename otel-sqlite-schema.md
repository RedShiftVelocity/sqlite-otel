# OpenTelemetry SQLite Database Schemas

Based on the official OTLP specification and data models from [OpenTelemetry Protocol](https://opentelemetry.io/docs/specs/otlp/) and the [opentelemetry-proto repository](https://github.com/open-telemetry/opentelemetry-proto).

## Traces Schema

### Core Trace Tables

```sql
-- Resource table (shared across all telemetry types)
CREATE TABLE IF NOT EXISTS resources (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    attributes TEXT NOT NULL, -- JSON object of resource attributes
    schema_url TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Instrumentation Scope table (shared across all telemetry types)  
CREATE TABLE IF NOT EXISTS instrumentation_scopes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    version TEXT,
    attributes TEXT, -- JSON object of scope attributes
    schema_url TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Main spans table
CREATE TABLE IF NOT EXISTS spans (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    
    -- Core span identifiers (from OTLP Span message)
    trace_id TEXT NOT NULL, -- 16-byte hex string (32 chars)
    span_id TEXT NOT NULL,  -- 8-byte hex string (16 chars)
    parent_span_id TEXT,    -- 8-byte hex string (16 chars), NULL for root spans
    trace_state TEXT,       -- W3C trace-state header value
    
    -- Span metadata
    name TEXT NOT NULL,
    kind INTEGER NOT NULL,  -- SpanKind enum: 0=UNSPECIFIED, 1=INTERNAL, 2=SERVER, 3=CLIENT, 4=PRODUCER, 5=CONSUMER
    
    -- Timing (nanoseconds since Unix epoch)
    start_time_unix_nano INTEGER NOT NULL,
    end_time_unix_nano INTEGER NOT NULL,
    
    -- Span data
    attributes TEXT,        -- JSON object of span attributes
    dropped_attributes_count INTEGER DEFAULT 0,
    
    -- Status
    status_code INTEGER DEFAULT 0,    -- 0=UNSET, 1=OK, 2=ERROR
    status_message TEXT,
    
    -- Flags and additional metadata
    flags INTEGER DEFAULT 0,          -- SpanFlags bit field
    
    -- Foreign keys
    resource_id INTEGER,
    scope_id INTEGER,
    
    -- Collection metadata
    ingested_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (resource_id) REFERENCES resources(id),
    FOREIGN KEY (scope_id) REFERENCES instrumentation_scopes(id)
);

-- Span events table
CREATE TABLE IF NOT EXISTS span_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    span_id INTEGER NOT NULL,
    
    time_unix_nano INTEGER NOT NULL,
    name TEXT NOT NULL,
    attributes TEXT, -- JSON object
    dropped_attributes_count INTEGER DEFAULT 0,
    
    FOREIGN KEY (span_id) REFERENCES spans(id) ON DELETE CASCADE
);

-- Span links table  
CREATE TABLE IF NOT EXISTS span_links (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    span_id INTEGER NOT NULL,
    
    trace_id TEXT NOT NULL,
    span_id_linked TEXT NOT NULL,
    trace_state TEXT,
    attributes TEXT, -- JSON object
    dropped_attributes_count INTEGER DEFAULT 0,
    flags INTEGER DEFAULT 0,
    
    FOREIGN KEY (span_id) REFERENCES spans(id) ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_spans_trace_id ON spans(trace_id);
CREATE INDEX IF NOT EXISTS idx_spans_span_id ON spans(span_id);
CREATE INDEX IF NOT EXISTS idx_spans_parent_span_id ON spans(parent_span_id);
CREATE INDEX IF NOT EXISTS idx_spans_start_time ON spans(start_time_unix_nano);
CREATE INDEX IF NOT EXISTS idx_spans_name ON spans(name);
CREATE INDEX IF NOT EXISTS idx_span_events_span_id ON span_events(span_id);
CREATE INDEX IF NOT EXISTS idx_span_links_span_id ON span_links(span_id);
```

## Metrics Schema

```sql
-- Main metrics table
CREATE TABLE IF NOT EXISTS metrics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    
    -- Metric metadata
    name TEXT NOT NULL,
    description TEXT,
    unit TEXT,
    type TEXT NOT NULL,     -- gauge, sum, histogram, exponential_histogram, summary
    
    -- Foreign keys
    resource_id INTEGER,
    scope_id INTEGER,
    
    -- Collection metadata
    ingested_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (resource_id) REFERENCES resources(id),
    FOREIGN KEY (scope_id) REFERENCES instrumentation_scopes(id)
);

-- Metric data points table (normalized for all metric types)
CREATE TABLE IF NOT EXISTS metric_data_points (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    metric_id INTEGER NOT NULL,
    
    -- Common fields for all data point types
    attributes TEXT,        -- JSON object of data point attributes
    start_time_unix_nano INTEGER,
    time_unix_nano INTEGER NOT NULL,
    flags INTEGER DEFAULT 0,
    
    -- Value fields (type-specific, most will be NULL for any given point)
    -- For Gauge and Sum
    value_double REAL,
    value_int INTEGER,
    
    -- For Sum only
    aggregation_temporality INTEGER, -- 0=UNSPECIFIED, 1=DELTA, 2=CUMULATIVE
    is_monotonic BOOLEAN,
    
    -- For Histogram and ExponentialHistogram
    count INTEGER,
    sum_value REAL,
    min_value REAL,
    max_value REAL,
    
    -- For Histogram
    bucket_counts TEXT,     -- JSON array of bucket counts
    explicit_bounds TEXT,   -- JSON array of bucket boundaries
    
    -- For ExponentialHistogram  
    scale INTEGER,
    zero_count INTEGER,
    positive_offset INTEGER,
    positive_bucket_counts TEXT, -- JSON array
    negative_offset INTEGER,
    negative_bucket_counts TEXT, -- JSON array
    
    -- For Summary
    quantile_values TEXT,   -- JSON array of {quantile: value} objects
    
    FOREIGN KEY (metric_id) REFERENCES metrics(id) ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_metrics_name ON metrics(name);
CREATE INDEX IF NOT EXISTS idx_metrics_type ON metrics(type);
CREATE INDEX IF NOT EXISTS idx_metric_data_points_metric_id ON metric_data_points(metric_id);
CREATE INDEX IF NOT EXISTS idx_metric_data_points_time ON metric_data_points(time_unix_nano);
```

## Logs Schema

```sql
-- Main log records table
CREATE TABLE IF NOT EXISTS log_records (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    
    -- Timing
    time_unix_nano INTEGER,         -- When the event occurred
    observed_time_unix_nano INTEGER NOT NULL, -- When it was observed by the collection system
    
    -- Severity
    severity_number INTEGER,        -- Numeric severity (1-24 range)
    severity_text TEXT,            -- Textual severity level
    
    -- Content
    body TEXT,                     -- The log message/content (AnyValue as JSON)
    attributes TEXT,               -- JSON object of log attributes
    
    -- Trace correlation
    trace_id TEXT,                 -- 16-byte hex string, links to trace
    span_id TEXT,                  -- 8-byte hex string, links to span
    trace_flags INTEGER DEFAULT 0, -- W3C trace flags
    
    -- Foreign keys
    resource_id INTEGER,
    scope_id INTEGER,
    
    -- Collection metadata  
    ingested_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (resource_id) REFERENCES resources(id),
    FOREIGN KEY (scope_id) REFERENCES instrumentation_scopes(id)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_log_records_time ON log_records(time_unix_nano);
CREATE INDEX IF NOT EXISTS idx_log_records_observed_time ON log_records(observed_time_unix_nano);
CREATE INDEX IF NOT EXISTS idx_log_records_severity ON log_records(severity_number);
CREATE INDEX IF NOT EXISTS idx_log_records_trace_id ON log_records(trace_id);
CREATE INDEX IF NOT EXISTS idx_log_records_span_id ON log_records(span_id);
```

## OTLP JSON Data Format Examples

### Trace JSON Structure (OTLP)
```json
{
  "resourceSpans": [
    {
      "resource": {
        "attributes": [
          {"key": "service.name", "value": {"stringValue": "my-service"}},
          {"key": "service.version", "value": {"stringValue": "1.0.0"}}
        ]
      },
      "scopeSpans": [
        {
          "scope": {
            "name": "my-instrumentation-library",
            "version": "1.0.0"
          },
          "spans": [
            {
              "traceId": "1234567890abcdef1234567890abcdef",
              "spanId": "1234567890abcdef",
              "parentSpanId": "fedcba0987654321",
              "name": "HTTP GET /api/users",
              "kind": 2,
              "startTimeUnixNano": "1640995200000000000",
              "endTimeUnixNano": "1640995200100000000",
              "attributes": [
                {"key": "http.method", "value": {"stringValue": "GET"}},
                {"key": "http.status_code", "value": {"intValue": "200"}}
              ],
              "status": {
                "code": 1
              }
            }
          ]
        }
      ]
    }
  ]
}
```

### Metric JSON Structure (OTLP)
```json
{
  "resourceMetrics": [
    {
      "resource": {
        "attributes": [
          {"key": "service.name", "value": {"stringValue": "my-service"}}
        ]
      },
      "scopeMetrics": [
        {
          "scope": {"name": "my-meter"},
          "metrics": [
            {
              "name": "http_requests_total",
              "description": "Total number of HTTP requests",
              "unit": "1",
              "sum": {
                "dataPoints": [
                  {
                    "attributes": [
                      {"key": "method", "value": {"stringValue": "GET"}}
                    ],
                    "timeUnixNano": "1640995200000000000",
                    "asInt": "42"
                  }
                ],
                "aggregationTemporality": 2,
                "isMonotonic": true
              }
            }
          ]
        }
      ]
    }
  ]
}
```

### Log JSON Structure (OTLP)
```json
{
  "resourceLogs": [
    {
      "resource": {
        "attributes": [
          {"key": "service.name", "value": {"stringValue": "my-service"}}
        ]
      },
      "scopeLogs": [
        {
          "scope": {"name": "my-logger"},
          "logRecords": [
            {
              "timeUnixNano": "1640995200000000000",
              "observedTimeUnixNano": "1640995200000000000",
              "severityNumber": 9,
              "severityText": "INFO",
              "body": {"stringValue": "User login successful"},
              "attributes": [
                {"key": "user.id", "value": {"stringValue": "12345"}},
                {"key": "session.id", "value": {"stringValue": "abc123"}}
              ],
              "traceId": "1234567890abcdef1234567890abcdef",
              "spanId": "1234567890abcdef"
            }
          ]
        }
      ]
    }
  ]
}
```

## Schema Design Notes

1. **Normalization**: Resources and instrumentation scopes are normalized into separate tables since they're shared across multiple telemetry records.

2. **JSON Storage**: Complex nested structures (attributes, arrays) are stored as JSON text for flexibility while maintaining the OTLP structure.

3. **Time Fields**: All time fields use INTEGER to store nanosecond Unix timestamps as specified in OTLP.

4. **Hex String IDs**: Trace and Span IDs are stored as hex strings (not binary) for easier querying and debugging.

5. **Type Flexibility**: The schemas accommodate all OTLP data types while keeping the structure queryable.

6. **Performance**: Indexes are created on commonly queried fields like trace_id, time fields, and names.

This schema maintains full fidelity with the OTLP specification while providing efficient storage and query capabilities for your SQLite collector.