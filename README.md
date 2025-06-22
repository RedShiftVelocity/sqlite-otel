# OpenTelemetry SQLite Collector Service

## Overview
This project involves creating a standalone Go binary that functions as an OpenTelemetry collector service, designed to receive telemetry data and persist it to an embedded SQLite database. The service will be deployable as both an on-demand process and a system service on Linux systems.

## Requirements
- Go 1.21 or higher
- SQLite 3.24.0 or higher (for ON CONFLICT clause support)

## Installation

### Homebrew (macOS & Linux)
```bash
# Install directly from GitHub
brew install --formula https://raw.githubusercontent.com/RedShiftVelocity/sqlite-otel/main/homebrew/sqlite-otel-collector.rb

# Start immediately
sqlite-otel-collector

# Or run as background service
brew services start sqlite-otel-collector
```

### Download Pre-built Binaries
Download the latest release from [GitHub Releases](https://github.com/RedShiftVelocity/sqlite-otel/releases):

```bash
# Linux x86_64
curl -LO https://github.com/RedShiftVelocity/sqlite-otel/releases/latest/download/sqlite-otel-linux-amd64
chmod +x sqlite-otel-linux-amd64
./sqlite-otel-linux-amd64

# macOS Intel
curl -LO https://github.com/RedShiftVelocity/sqlite-otel/releases/latest/download/sqlite-otel-darwin-amd64
chmod +x sqlite-otel-darwin-amd64
./sqlite-otel-darwin-amd64

# macOS Apple Silicon
curl -LO https://github.com/RedShiftVelocity/sqlite-otel/releases/latest/download/sqlite-otel-darwin-arm64
chmod +x sqlite-otel-darwin-arm64
./sqlite-otel-darwin-arm64
```

### Package Managers
```bash
# Ubuntu/Debian
curl -LO https://github.com/RedShiftVelocity/sqlite-otel/releases/latest/download/sqlite-otel-collector_0.8.0_amd64.deb
sudo dpkg -i sqlite-otel-collector_0.8.0_amd64.deb

# Verify installation
sudo systemctl start sqlite-otel-collector
sudo systemctl status sqlite-otel-collector
```

## Quick Start

### Binary Usage
```bash
# Run with default settings
./sqlite-otel

# Run with custom port and database
./sqlite-otel -port 4318 -db-path ./my-data.db

# Show version information
./sqlite-otel -version

# Show help
./sqlite-otel -h
```

### Docker Usage
```bash
# Pull latest development image
docker pull ghcr.io/redshiftvelocity/sqlite-otel:latest

# Run with default settings
docker run -d --name sqlite-otel -p 4318:4318 \
  ghcr.io/redshiftvelocity/sqlite-otel:latest

# Run with custom configuration
docker run -d --name sqlite-otel -p 4318:4318 \
  -v $(pwd)/data:/var/lib/sqlite-otel-collector \
  ghcr.io/redshiftvelocity/sqlite-otel:latest

# Run development version
docker run -d --name sqlite-otel-dev -p 4318:4318 \
  ghcr.io/redshiftvelocity/sqlite-otel:v0.7.95-dev

# Check container logs
docker logs sqlite-otel

# Test the service
curl -X POST http://localhost:4318/v1/traces \
  -H "Content-Type: application/json" \
  -d '{"resourceSpans": []}'
```

### Service Installation
```bash
# Install to /usr/local/bin (requires sudo)
sudo make install

# Uninstall
sudo make uninstall
```

## Semantic Versioning

The project uses semantic versioning with automatic revision numbers:

- **Format**: `v{major}.{minor}.{commit_count}`
- **Development**: `v0.7.86` (86 total commits)
- **Tagged releases**: `v1.0.0` (uses exact git tag)
- **Configurable**: Update `MAJOR_MINOR` in Makefile to change version

## Core Architecture
The Go binary will implement a lightweight OTEL collector that listens on an ephemeral port for incoming telemetry data (traces, metrics, and logs) in OpenTelemetry Protocol (OTLP) format. Upon receiving data, it will immediately persist the information to a local SQLite database using an embedded database approach. This eliminates external dependencies and simplifies deployment while maintaining data durability.

The service will automatically bind to an available ephemeral port (typically in the 32768-65535 range) to avoid port conflicts, making it suitable for multi-instance deployments or development environments. The actual port will be logged at startup for client configuration.

### Command Line Interface

The binary accepts the following command-line arguments:

| Flag | Description | Default |
|------|-------------|---------|
| `-port` | Port to listen on | `4318` (OTLP/HTTP standard) |
| `-db-path` | Path to SQLite database file | User mode: `~/.local/share/sqlite-otel/otel-collector.db`<br>Service mode: `/var/lib/sqlite-otel-collector/otel-collector.db` |
| `-log-file` | Path to log file for execution metadata | User mode: `~/.local/state/sqlite-otel/execution.log`<br>Service mode: `/var/log/sqlite-otel-collector.log` |
| `-log-max-size` | Maximum log file size in MB before rotation | `100` |
| `-log-max-backups` | Maximum number of old log files to keep | `7` |
| `-log-max-age` | Maximum number of days to keep old log files | `30` |
| `-log-compress` | Compress rotated log files | `true` |
| `-version` | Show version information | - |

### Path Detection

The application automatically detects whether it's running in:
- **User mode**: Uses XDG Base Directory specification (`~/.local/share`, `~/.local/state`)
- **Service mode**: Uses system directories (`/var/lib`, `/var/log`) when running under systemd

## Service Integration
The binary will include built-in service management capabilities for Linux systems. It will generate systemd service files and provide installation commands, enabling easy deployment as a system daemon. The service will handle graceful shutdown signals, ensuring data integrity during system restarts or maintenance.

Database schema will be automatically initialized on first startup, creating tables for traces, spans, metrics, and logs with appropriate indexing for performance. The SQLite database will use WAL (Write-Ahead Logging) mode to support concurrent access patterns if needed for future enhancements.

## Wish to contribute?

### Building
```bash
# Build for current platform
make build

# Build for all platforms (Linux, macOS, Windows)
make build-all

# Build only Linux binaries
make build-linux

# Development build with race detector
make dev
```

### Testing
```bash
# Run tests
make test

# Run tests with coverage report
make test-coverage

# Format code
make fmt

# Verify dependencies
make verify
```

### Distribution
```bash
# Create release archives for all platforms
make release

# Clean build artifacts
make clean

# Show version information
make version

# Show all available commands
make help
```

### Makefile Commands

The project includes a comprehensive Makefile with the following targets:

| Command | Description | Example Output |
|---------|-------------|----------------|
| `make help` | Display all available commands | Lists all make targets with descriptions |
| `make version` | Show version information | `Version: v0.7.86`<br>`Git Commit: 1fe7bdb`<br>`Build Time: 2025-06-21_04:51:28` |
| `make build` | Build for current platform | `Building sqlite-otel v0.7.86 for current platform...` |
| `make build-all` | Build for all platforms | Creates binaries in `dist/` for Linux, macOS, Windows |
| `make build-linux` | Build Linux binaries only | Creates `sqlite-otel-linux-{amd64,arm64,arm}` in `dist/` |
| `make test` | Run tests with race detection | `ok github.com/RedShiftVelocity/sqlite-otel/logging 8.872s coverage: 58.1%` |
| `make test-coverage` | Generate HTML coverage report | Creates `coverage.html` |
| `make clean` | Remove build artifacts | Cleans `dist/`, `releases/`, coverage files |
| `make fmt` | Format Go code | Formats all `.go` files |
| `make lint` | Run golangci-lint | Requires golangci-lint installation |
| `make tidy` | Tidy Go modules | `go mod tidy` |
| `make verify` | Verify dependencies | `all modules verified` |
| `make release` | Create versioned archives | Creates `.tar.gz` files in `releases/` |
| `make dev` | Development build with race detector | Build with `-race` flag |
| `make run-example` | Run with example flags | Starts server on random port with test DB |
| `make install` | Install to `/usr/local/bin` | Requires `sudo` for system installation |
| `make uninstall` | Remove from `/usr/local/bin` | Requires `sudo` for system removal |

#### Cross-Platform Builds
```bash
$ make build-all
Building for all platforms...
Building sqlite-otel-linux-amd64...
Building sqlite-otel-linux-arm64...
Building sqlite-otel-linux-arm...
Building sqlite-otel-darwin-amd64...
Building sqlite-otel-darwin-arm64...
Building sqlite-otel-windows-amd64.exe...
Build complete. Binaries in dist/

$ ls -la dist/
-rwxrwxr-x sqlite-otel-darwin-amd64      5.5MB
-rwxrwxr-x sqlite-otel-darwin-arm64      5.3MB  
-rwxrwxr-x sqlite-otel-linux-amd64       6.8MB
-rwxrwxr-x sqlite-otel-linux-arm         5.2MB
-rwxrwxr-x sqlite-otel-linux-arm64       5.1MB
-rwxrwxr-x sqlite-otel-windows-amd64.exe 5.5MB
```

#### Release Archives
```bash
$ make release
Creating release archives...
Created releases/sqlite-otel-darwin-amd64-v0.7.86.tar.gz
Created releases/sqlite-otel-darwin-arm64-v0.7.86.tar.gz
Created releases/sqlite-otel-linux-amd64-v0.7.86.tar.gz
Created releases/sqlite-otel-linux-arm-v0.7.86.tar.gz
Created releases/sqlite-otel-linux-arm64-v0.7.86.tar.gz
Created releases/sqlite-otel-windows-amd64.exe-v0.7.86.tar.gz
```

#### Version Information
```bash
$ make version
Version: v0.7.86
Git Commit: 1fe7bdb
Build Time: 2025-06-21_04:51:28

$ ./sqlite-otel -version
sqlite-otel-collector v0.7.86
Build Time: 2025-06-21_04:51:28
Git Commit: 1fe7bdb
```

## Development Roadmap

### Future Enhancements

#### v1.1-1.3 - Configuration Options
- **v1.1**: SQLite file location argument (--db-path)
- **v1.2**: Log file location argument (--log-file)  
- **v1.3**: Logging mode argument (--log-level)

#### v2.0 - Enhanced Public Release
- Complete command-line interface
- Configuration file support
- Advanced deployment options

#### v2.1+ - Extended Platform Support
- **Docker Containerization**: Multi-stage builds, minimal base images, health checks
- **Windows Service Support**: Native Windows service integration, MSI installer packages
- **Advanced Executable Arguments**: Configuration files, environment variable support, validation

#### Future Considerations
- Multiple OTEL protocol support (gRPC, HTTP, custom)
- Horizontal scaling and clustering capabilities
- Web-based management interface
- Integration with major observability platforms