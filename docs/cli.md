# CLI Reference

Complete command-line interface reference for the SQLite OTEL Collector.

## Synopsis

```bash
sqlite-otel [OPTIONS]
```

## Options

### Core Options

#### `-port`
**Type**: Integer  
**Default**: `4318`  
**Description**: Port number for the OTLP HTTP server to listen on.

```bash
sqlite-otel -port 9090
```

#### `-db-path`
**Type**: String  
**Default**: Auto-detected based on execution context  
**Description**: Path to the SQLite database file.

```bash
sqlite-otel -db-path /data/telemetry.db
```

**Auto-detection rules:**
- **User mode**: `~/.local/share/sqlite-otel/otel-collector.db`
- **Service mode**: `/var/lib/sqlite-otel-collector/otel-collector.db`

### Logging Options

#### `-log-file`
**Type**: String  
**Default**: Auto-detected based on execution context  
**Description**: Path to the log file for execution metadata.

```bash
sqlite-otel -log-file /var/log/my-collector.log
```

**Auto-detection rules:**
- **User mode**: `~/.local/state/sqlite-otel/execution.log`
- **Service mode**: `/var/log/sqlite-otel-collector.log`

#### `-log-max-size`
**Type**: Integer  
**Default**: `100`  
**Description**: Maximum size of log file in MB before rotation.

```bash
sqlite-otel -log-max-size 50
```

#### `-log-max-backups`
**Type**: Integer  
**Default**: `7`  
**Description**: Maximum number of old log files to keep.

```bash
sqlite-otel -log-max-backups 3
```

#### `-log-max-age`
**Type**: Integer  
**Default**: `30`  
**Description**: Maximum number of days to keep old log files.

```bash
sqlite-otel -log-max-age 14
```

#### `-log-compress`
**Type**: Boolean  
**Default**: `true`  
**Description**: Compress rotated log files using gzip.

```bash
sqlite-otel -log-compress=false
```

### Utility Options

#### `-version`
**Type**: Flag  
**Description**: Display version information and exit.

```bash
sqlite-otel -version
```

**Output example:**
```
sqlite-otel-collector v0.7.95
Build Time: 2025-06-22_12:34:56
Git Commit: abc1234
```

#### `-h`, `-help`
**Type**: Flag  
**Description**: Display help information and exit.

```bash
sqlite-otel -h
sqlite-otel -help
```

## Usage Examples

### Basic Usage

```bash
# Start with defaults (port 4318, auto-detected paths)
sqlite-otel

# Start on custom port
sqlite-otel -port 8080

# Start with custom database
sqlite-otel -db-path ./my-data.db
```

### Production Configuration

```bash
# Full production configuration
sqlite-otel \
  -port 4318 \
  -db-path /opt/telemetry/collector.db \
  -log-file /var/log/sqlite-otel/collector.log \
  -log-max-size 100 \
  -log-max-backups 10 \
  -log-max-age 90 \
  -log-compress=true
```

### Development Configuration

```bash
# Development setup with local paths
sqlite-otel \
  -port 4318 \
  -db-path ./dev-data.db \
  -log-file ./dev.log \
  -log-max-size 10 \
  -log-max-backups 2
```

### Minimal Resource Configuration

```bash
# For resource-constrained environments
sqlite-otel \
  -port 4318 \
  -log-max-size 5 \
  -log-max-backups 1 \
  -log-max-age 3 \
  -log-compress=true
```

## Exit Codes

| Code | Description |
|------|-------------|
| `0` | Success |
| `1` | General error |
| `2` | Invalid command-line arguments |
| `3` | Database initialization error |
| `4` | Server startup error |
| `5` | Permission denied |

## Signal Handling

The collector handles the following signals gracefully:

### SIGTERM
Graceful shutdown - closes database connections and stops the server.

```bash
# Send graceful shutdown signal
kill -TERM <pid>
```

### SIGINT (Ctrl+C)
Immediate shutdown - same as SIGTERM but triggered by Ctrl+C.

### SIGUSR1
Log rotation trigger - forces log rotation immediately.

```bash
# Force log rotation
kill -USR1 <pid>
```

## Environment Context Detection

The collector automatically detects its execution context:

### User Mode Detection
- Not running under systemd
- No special privileges
- Uses XDG Base Directory specification

### Service Mode Detection
- Running under systemd (detected via `INVOCATION_ID` environment variable)
- Typically running as dedicated user
- Uses system-wide paths

## Validation

### Path Validation
The collector validates all file paths on startup:

- **Database path**: Must be writable directory
- **Log path**: Must be writable directory
- **File permissions**: Checks read/write access

### Port Validation
- Must be between 1 and 65535
- Must not be already in use
- Must be accessible (not blocked by firewall)

### Configuration Validation
```bash
# Test configuration without starting server
sqlite-otel -port 4318 -db-path /test/path.db
# Will exit with error if path is not accessible
```

## Docker Usage

When running in Docker, pass CLI arguments to the container:

```bash
# Basic Docker usage
docker run -p 4318:4318 ghcr.io/redshiftvelocity/sqlite-otel:latest

# Docker with custom arguments
docker run -p 9090:9090 \
  ghcr.io/redshiftvelocity/sqlite-otel:latest \
  -port 9090 -log-max-size 25

# Docker with volume mounts
docker run -p 4318:4318 \
  -v $(pwd)/data:/data \
  ghcr.io/redshiftvelocity/sqlite-otel:latest \
  -db-path /data/collector.db
```

## Systemd Integration

When installed as a systemd service, arguments are configured in the service file:

```ini
[Unit]
Description=SQLite OpenTelemetry Collector
After=network.target

[Service]
Type=simple
User=sqlite-otel
ExecStart=/usr/bin/sqlite-otel-collector -port 4318
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

Override with systemctl:

```bash
# Edit service configuration
sudo systemctl edit sqlite-otel-collector

# Add custom configuration
[Service]
ExecStart=
ExecStart=/usr/bin/sqlite-otel-collector -port 4318 -log-max-size 50
```

## Troubleshooting

### Common Issues

**Port already in use:**
```bash
sqlite-otel -port 4319  # Use different port
```

**Permission denied:**
```bash
# Check path permissions
ls -la /path/to/database/
# Ensure user has write access
```

**Database locked:**
```bash
# Check for other processes using the database
lsof /path/to/database.db
```

### Debugging

Enable verbose logging by examining the log file:

```bash
# Follow logs in real-time
tail -f ~/.local/state/sqlite-otel/execution.log

# Or for systemd service
sudo journalctl -u sqlite-otel-collector -f
```

## See Also

- [Configuration Guide](configuration.md) - Detailed configuration options
- [Quick Start](quickstart.md) - Getting started guide  
- [API Reference](api.md) - OTLP endpoint documentation