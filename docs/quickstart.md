# Quick Start

Get your SQLite OTEL Collector up and running in under 5 minutes.

## Step 1: Start the Collector

=== "Docker (Recommended)"

    ```bash
    # Start the collector
    docker run -d \
      --name sqlite-otel \
      -p 4318:4318 \
      -v sqlite-otel-data:/var/lib/sqlite-otel-collector \
      ghcr.io/redshiftvelocity/sqlite-otel:latest

    # Verify it's running
    docker logs sqlite-otel
    ```

=== "Binary"

    ```bash
    # Start the collector (runs in foreground)
    sqlite-otel

    # Or run in background
    sqlite-otel &
    ```

=== "Service"

    ```bash
    # If installed via package manager
    sudo systemctl start sqlite-otel-collector
    sudo systemctl status sqlite-otel-collector
    ```

## Step 2: Verify Collector is Running

```bash
# Test health endpoint
curl http://localhost:4318/health

# Expected response: 200 OK
```

## Step 3: Send Test Data

### Send a Test Trace

```bash
curl -X POST http://localhost:4318/v1/traces \
  -H "Content-Type: application/json" \
  -d '{
    "resourceSpans": [{
      "resource": {
        "attributes": [
          {
            "key": "service.name",
            "value": {"stringValue": "test-service"}
          }
        ]
      },
      "scopeSpans": [{
        "spans": [{
          "traceId": "0123456789abcdef0123456789abcdef",
          "spanId": "0123456789abcdef",
          "name": "test-span",
          "kind": 1,
          "startTimeUnixNano": "1640995200000000000",
          "endTimeUnixNano": "1640995201000000000"
        }]
      }]
    }]
  }'
```

### Send Test Metrics

```bash
curl -X POST http://localhost:4318/v1/metrics \
  -H "Content-Type: application/json" \
  -d '{
    "resourceMetrics": [{
      "resource": {
        "attributes": [
          {
            "key": "service.name", 
            "value": {"stringValue": "test-service"}
          }
        ]
      },
      "scopeMetrics": [{
        "metrics": [{
          "name": "test_counter",
          "description": "A test counter metric",
          "unit": "1",
          "sum": {
            "dataPoints": [{
              "attributes": [],
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

## Step 4: Query Your Data

The collector stores all data in SQLite. You can query it directly:

```bash
# Access the SQLite database
sqlite3 /var/lib/sqlite-otel-collector/otel-collector.db

# Or if running locally
sqlite3 ~/.local/share/sqlite-otel/otel-collector.db
```

### Sample Queries

```sql
-- List all traces
SELECT trace_id, span_id, name, service_name 
FROM spans 
ORDER BY start_time DESC 
LIMIT 10;

-- List all metrics
SELECT name, value, timestamp, service_name 
FROM metrics 
ORDER BY timestamp DESC 
LIMIT 10;

-- Count spans by service
SELECT service_name, COUNT(*) as span_count 
FROM spans 
GROUP BY service_name;
```

## Step 5: Configure Your Applications

Point your OpenTelemetry-instrumented applications to send data to the collector:

=== "Environment Variables"

    ```bash
    export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
    export OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf
    ```

=== "Go SDK"

    ```go
    import "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"

    exporter, err := otlptracehttp.New(
        context.Background(),
        otlptracehttp.WithEndpoint("http://localhost:4318"),
        otlptracehttp.WithInsecure(),
    )
    ```

=== "Python SDK"

    ```python
    from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter

    exporter = OTLPSpanExporter(
        endpoint="http://localhost:4318/v1/traces"
    )
    ```

## Common Configuration

### Custom Port

```bash
# Run on different port
sqlite-otel -port 9999

# Docker with custom port
docker run -d -p 9999:9999 \
  ghcr.io/redshiftvelocity/sqlite-otel:latest \
  -port 9999
```

### Custom Database Location

```bash
# Run with custom database path
sqlite-otel -db-path ./my-telemetry.db

# Docker with custom database
docker run -d -p 4318:4318 \
  -v $(pwd)/data:/data \
  ghcr.io/redshiftvelocity/sqlite-otel:latest \
  -db-path /data/telemetry.db
```

## Troubleshooting

!!! warning "Common Issues"

    **Port already in use**: Change the port with `-port 4319` or stop other services
    
    **Permission denied**: Make sure the user has write access to the database directory
    
    **Connection refused**: Check if the collector is running and firewall settings

```bash
# Check if collector is running
curl -f http://localhost:4318/health || echo "Collector not responding"

# Check Docker logs
docker logs sqlite-otel

# Check systemd logs
sudo journalctl -u sqlite-otel-collector -f
```

## Next Steps

- [Configuration Guide](configuration.md) - Customize collector behavior
- [CLI Reference](cli.md) - All command-line options
- [Deployment Guide](deployment.md) - Production deployment strategies

!!! success "You're Ready!"
    Your SQLite OTEL Collector is now collecting telemetry data. Start sending data from your applications!