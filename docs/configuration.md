# Configuration

Configure the SQLite OTEL Collector to suit your specific needs.

## Command Line Options

The collector can be configured using command-line flags:

| Flag | Description | Default | Example |
|------|-------------|---------|---------|
| `-port` | Port to listen on | `4318` | `-port 9090` |
| `-db-path` | SQLite database file path | Auto-detected | `-db-path ./data.db` |
| `-log-file` | Log file path | Auto-detected | `-log-file ./app.log` |
| `-log-max-size` | Max log file size (MB) | `100` | `-log-max-size 50` |
| `-log-max-backups` | Max number of log backups | `7` | `-log-max-backups 3` |
| `-log-max-age` | Max log file age (days) | `30` | `-log-max-age 14` |
| `-log-compress` | Compress rotated logs | `true` | `-log-compress=false` |
| `-version` | Show version and exit | - | `-version` |

## Path Detection

The collector automatically detects appropriate paths based on the execution context:

### User Mode (Default)
- **Database**: `~/.local/share/sqlite-otel/otel-collector.db`
- **Logs**: `~/.local/state/sqlite-otel/execution.log`

### Service Mode (systemd)
- **Database**: `/var/lib/sqlite-otel-collector/otel-collector.db`
- **Logs**: `/var/log/sqlite-otel-collector.log`

## Examples

### Basic Usage

```bash
# Default configuration
sqlite-otel

# Custom port
sqlite-otel -port 8080

# Custom database location
sqlite-otel -db-path /data/telemetry.db
```

### Production Configuration

```bash
# Production setup with custom paths and log rotation
sqlite-otel \
  -port 4318 \
  -db-path /opt/telemetry/data.db \
  -log-file /var/log/sqlite-otel/collector.log \
  -log-max-size 50 \
  -log-max-backups 10 \
  -log-max-age 90 \
  -log-compress=true
```

### Docker Configuration

```bash
# Docker with environment variables
docker run -d \
  --name sqlite-otel \
  -p 4318:4318 \
  -v sqlite-data:/var/lib/sqlite-otel-collector \
  -v sqlite-logs:/var/log \
  ghcr.io/redshiftvelocity/sqlite-otel:latest \
  -log-max-size 25 \
  -log-max-backups 5
```

## Environment Variables

While the collector primarily uses command-line flags, you can set some environment variables for container deployments:

```bash
# Set in Docker or Kubernetes
export SQLITE_OTEL_PORT=4318
export SQLITE_OTEL_DB_PATH=/data/collector.db
```

## Database Configuration

### SQLite Settings

The collector automatically configures SQLite with optimal settings:

- **WAL Mode**: Enabled for better concurrency
- **Synchronous**: NORMAL for balance of safety and performance  
- **Journal Mode**: WAL (Write-Ahead Logging)
- **Foreign Keys**: Enabled for data integrity

### Schema Overview

The collector creates these tables automatically:

```sql
-- Traces and spans
CREATE TABLE spans (
  trace_id TEXT,
  span_id TEXT,
  parent_span_id TEXT,
  name TEXT,
  kind INTEGER,
  start_time INTEGER,
  end_time INTEGER,
  service_name TEXT,
  attributes TEXT -- JSON
);

-- Metrics
CREATE TABLE metrics (
  name TEXT,
  description TEXT,
  unit TEXT,
  type TEXT,
  value REAL,
  timestamp INTEGER,
  service_name TEXT,
  attributes TEXT -- JSON
);

-- Logs
CREATE TABLE logs (
  timestamp INTEGER,
  severity TEXT,
  body TEXT,
  service_name TEXT,
  trace_id TEXT,
  span_id TEXT,
  attributes TEXT -- JSON
);
```

## Log Rotation

The collector includes built-in log rotation to manage disk usage:

### Rotation Triggers

- **Size-based**: When log file exceeds `log-max-size` MB
- **Time-based**: Checked daily, removes files older than `log-max-age` days
- **Count-based**: Keeps only `log-max-backups` old files

### Rotation Behavior

```bash
# Example log rotation
collector.log          # Current log
collector.log.1         # Previous log (compressed if enabled)
collector.log.2.gz      # Older log (compressed)
collector.log.3.gz      # Oldest kept log
```

### Disable Rotation

```bash
# Disable rotation (not recommended for production)
sqlite-otel \
  -log-max-size 0 \
  -log-max-backups 0 \
  -log-max-age 0
```

## Performance Tuning

### For High-Volume Deployments

```bash
# Optimized for high throughput
sqlite-otel \
  -port 4318 \
  -db-path /fast-storage/telemetry.db \
  -log-max-size 200 \
  -log-max-backups 3
```

### For Resource-Constrained Environments

```bash
# Minimal resource usage
sqlite-otel \
  -port 4318 \
  -log-max-size 10 \
  -log-max-backups 2 \
  -log-max-age 7
```

## Security Considerations

### File Permissions

Ensure proper file permissions:

```bash
# Set secure permissions for database
chmod 600 /path/to/database.db
chown sqlite-otel:sqlite-otel /path/to/database.db

# Set secure permissions for logs
chmod 640 /path/to/collector.log
chown sqlite-otel:adm /path/to/collector.log
```

### Network Security

```bash
# Bind to specific interface (not all interfaces)
# Note: Currently binds to all interfaces (0.0.0.0)
# Consider using firewall rules or reverse proxy for additional security
```

## Systemd Service Configuration

When installed via package manager, edit the systemd service:

```bash
# Edit service configuration
sudo systemctl edit sqlite-otel-collector

# Add override configuration
[Service]
ExecStart=
ExecStart=/usr/bin/sqlite-otel-collector -port 4318 -log-max-size 50
```

## Configuration Validation

Validate your configuration before deployment:

```bash
# Test configuration
sqlite-otel -version

# Test with dry-run (check paths are accessible)
sqlite-otel -db-path /test/path.db
# Should fail if path is not writable
```

## Next Steps

- [CLI Reference](cli.md) - Complete command-line reference
- [Deployment Guide](deployment.md) - Production deployment strategies  
- [API Reference](api.md) - OTLP endpoint documentation