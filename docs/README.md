# SQLite OpenTelemetry Collector Documentation

<div align="center">
  <img src="../site/assets/logo.svg" alt="SQLite OTEL Collector" width="200">
  
  [![GitHub Release](https://img.shields.io/github/v/release/RedShiftVelocity/sqlite-otel)](https://github.com/RedShiftVelocity/sqlite-otel/releases)
  [![Docker Pulls](https://img.shields.io/docker/pulls/redshiftvelocity/sqlite-otel)](https://hub.docker.com/r/redshiftvelocity/sqlite-otel)
  [![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://opensource.org/licenses/MIT)
  
  **A lightweight, embedded OpenTelemetry collector with SQLite storage**
</div>

## 📋 Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Installation](#installation)
- [Configuration](#configuration)
- [Deployment](#deployment)
- [Operations](#operations)
- [Reference](#reference)
- [Support](#support)

## Overview

SQLite OpenTelemetry Collector is a lightweight, single-binary telemetry collector that persists traces, metrics, and logs directly to an embedded SQLite database. Perfect for edge deployments, development environments, and resource-constrained systems.

### Key Features

- 🚀 **Single Binary** - No external dependencies, just download and run
- 💾 **Embedded Storage** - SQLite database for reliable local persistence
- 🔒 **Security First** - Runs as non-root with comprehensive hardening
- 📊 **Full Telemetry** - Supports traces, metrics, and logs via OTLP/HTTP
- 🔄 **Log Rotation** - Built-in rotation with compression support
- 📦 **Easy Deployment** - Native packages for major platforms

### Use Cases

- **Edge Computing** - Collect telemetry at edge locations with intermittent connectivity
- **Development** - Local telemetry collection without complex infrastructure
- **Embedded Systems** - Lightweight collector for resource-constrained environments
- **Compliance** - Keep sensitive telemetry data on-premises

## Quick Start

Get up and running in under 5 minutes:

### Docker

```bash
docker run -d \
  --name sqlite-otel \
  -p 4318:4318 \
  -v sqlite-otel-data:/var/lib/sqlite-otel-collector \
  redshiftvelocity/sqlite-otel:latest
```

### Binary

```bash
# Download latest release
curl -LO https://github.com/RedShiftVelocity/sqlite-otel/releases/latest/download/sqlite-otel-collector-linux-amd64.tar.gz
tar xzf sqlite-otel-collector-linux-amd64.tar.gz

# Run collector
./sqlite-otel-collector
```

### Send Test Data

```bash
# Send a test trace
curl -X POST http://localhost:4318/v1/traces \
  -H "Content-Type: application/json" \
  -d @examples/trace.json

# Send test metrics
curl -X POST http://localhost:4318/v1/metrics \
  -H "Content-Type: application/json" \
  -d @examples/metrics.json
```

## Installation

### 📦 Package Managers

<details>
<summary><b>Ubuntu/Debian</b></summary>

```bash
# Add repository
curl -s https://packages.sqlite-otel.io/setup.deb.sh | sudo bash

# Install
sudo apt-get update
sudo apt-get install sqlite-otel-collector

# Start service
sudo systemctl enable --now sqlite-otel-collector
```

</details>

<details>
<summary><b>RHEL/CentOS/Fedora</b></summary>

```bash
# Add repository
curl -s https://packages.sqlite-otel.io/setup.rpm.sh | sudo bash

# Install
sudo yum install sqlite-otel-collector

# Start service
sudo systemctl enable --now sqlite-otel-collector
```

</details>

<details>
<summary><b>macOS (Homebrew)</b></summary>

```bash
# Add tap
brew tap redshiftvelocity/sqlite-otel

# Install
brew install sqlite-otel-collector

# Start service
brew services start sqlite-otel-collector
```

</details>

### 🐳 Docker

<details>
<summary><b>Docker Run</b></summary>

```bash
docker run -d \
  --name sqlite-otel \
  -p 4318:4318 \
  -v sqlite-otel-data:/var/lib/sqlite-otel-collector \
  -v sqlite-otel-logs:/var/log \
  --restart unless-stopped \
  redshiftvelocity/sqlite-otel:latest \
  --log-max-size 50 \
  --log-max-backups 5
```

</details>

<details>
<summary><b>Docker Compose</b></summary>

```yaml
version: '3.8'

services:
  sqlite-otel:
    image: redshiftvelocity/sqlite-otel:latest
    container_name: sqlite-otel
    ports:
      - "4318:4318"
    volumes:
      - sqlite-data:/var/lib/sqlite-otel-collector
      - sqlite-logs:/var/log
    environment:
      - LOG_MAX_SIZE=50
      - LOG_MAX_BACKUPS=5
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:4318/health"]
      interval: 30s
      timeout: 10s
      retries: 3

volumes:
  sqlite-data:
  sqlite-logs:
```

</details>

### 📥 Manual Installation

See the [installation guide](./installation/README.md) for:
- Binary installation
- Building from source
- Systemd service setup
- Verification steps

## Configuration

### Command Line Options

```bash
sqlite-otel-collector [options]

Options:
  --port            Port to listen on (default: 4318)
  --db-path         SQLite database path (default: platform-specific)
  --log-file        Log file path (default: platform-specific)
  --log-max-size    Maximum log size in MB before rotation (default: 100)
  --log-max-backups Maximum number of old log files (default: 7)
  --log-max-age     Maximum days to retain old files (default: 30)
  --log-compress    Compress rotated log files (default: true)
  --version         Show version information
  --help            Show help message
```

### Default Paths

| Component | User Mode | Service Mode |
|-----------|-----------|--------------|
| Database | `~/.local/share/sqlite-otel/otel-collector.db` | `/var/lib/sqlite-otel-collector/otel-collector.db` |
| Logs | `~/.local/state/sqlite-otel/execution.log` | `/var/log/sqlite-otel-collector.log` |

### Environment Variables

All command-line options can be set via environment variables:

```bash
export SQLITE_OTEL_PORT=4318
export SQLITE_OTEL_DB_PATH=/custom/path/database.db
export SQLITE_OTEL_LOG_FILE=/custom/path/app.log
export SQLITE_OTEL_LOG_MAX_SIZE=200
```

## Deployment

### Architecture Patterns

- **Standalone**: Single instance for local collection
- **Sidecar**: Deploy alongside applications in containers
- **DaemonSet**: One collector per node in Kubernetes
- **Edge Gateway**: Aggregate telemetry at edge locations

See the [deployment guide](./deployment/README.md) for detailed patterns and examples.

### Security Considerations

- Runs as dedicated `sqlite-otel` user (non-root)
- Systemd hardening with namespace isolation
- Read-only root filesystem in containers
- TLS support for encrypted communication

## Operations

### Monitoring

The collector exposes operational metrics:

- **Health Check**: `GET /health`
- **Metrics**: Internal metrics about processing
- **Logs**: Structured JSON logs with levels

### Database Management

```bash
# View database size
sqlite3 /var/lib/sqlite-otel-collector/otel-collector.db "SELECT page_count * page_size as size FROM pragma_page_count(), pragma_page_size();"

# Export data
sqlite3 /var/lib/sqlite-otel-collector/otel-collector.db ".mode json" "SELECT * FROM traces;" > traces.json

# Backup database
sqlite3 /var/lib/sqlite-otel-collector/otel-collector.db ".backup /backup/otel-backup.db"
```

### Troubleshooting

See the [troubleshooting guide](./operations/troubleshooting.md) for common issues and solutions.

## Reference

- [API Reference](./reference/api.md) - OTLP endpoints and formats
- [Database Schema](./reference/schema.md) - SQLite table structures
- [Configuration Reference](./reference/configuration.md) - All options
- [Performance Tuning](./reference/performance.md) - Optimization guide

## Support

- 📖 [Documentation](https://sqlite-otel.io/docs)
- 💬 [GitHub Discussions](https://github.com/RedShiftVelocity/sqlite-otel/discussions)
- 🐛 [Issue Tracker](https://github.com/RedShiftVelocity/sqlite-otel/issues)
- 📧 [Mailing List](https://groups.google.com/g/sqlite-otel)

## License

SQLite OpenTelemetry Collector is licensed under the [MIT License](../LICENSE).