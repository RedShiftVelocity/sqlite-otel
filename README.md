# OpenTelemetry SQLite Collector Service

## Overview
This project involves creating a standalone Go binary that functions as an OpenTelemetry collector service, designed to receive telemetry data and persist it to an embedded SQLite database. The service will be deployable as both an on-demand process and a system service on Linux systems.

## Requirements
- Go 1.21 or higher
- SQLite 3.24.0 or higher (for ON CONFLICT clause support)

## Core Architecture
The Go binary will implement a lightweight OTEL collector that listens on an ephemeral port for incoming telemetry data (traces, metrics, and logs) in OpenTelemetry Protocol (OTLP) format. Upon receiving data, it will immediately persist the information to a local SQLite database using an embedded database approach. This eliminates external dependencies and simplifies deployment while maintaining data durability.

The service will automatically bind to an available ephemeral port (typically in the 32768-65535 range) to avoid port conflicts, making it suitable for multi-instance deployments or development environments. The actual port will be logged at startup for client configuration.

## Command Line Interface
The binary accepts three optional command-line arguments with intelligent defaults:

**SQLite File Location (--db-path)**: Specifies the path for the SQLite database file. Default behavior will create a database file named `otel-collector.db` in the current working directory. For service deployment, this might default to `/var/lib/otel-collector/data.db` to follow Linux filesystem hierarchy standards.

**Log File Location (--log-file)**: Defines where the collector service writes its operational logs. The default will write to stdout for container/systemd compatibility, but can be overridden to write to `/var/log/otel-collector.log` for traditional service deployments.

**Logging Mode (--log-level)**: Controls the verbosity of operational logging with options like debug, info, warn, error. Default will be "info" level, providing sufficient operational visibility without overwhelming log output.

## Service Integration
The binary will include built-in service management capabilities for Linux systems. It will generate systemd service files and provide installation commands, enabling easy deployment as a system daemon. The service will handle graceful shutdown signals, ensuring data integrity during system restarts or maintenance.

Database schema will be automatically initialized on first startup, creating tables for traces, spans, metrics, and logs with appropriate indexing for performance. The SQLite database will use WAL (Write-Ahead Logging) mode to support concurrent access patterns if needed for future enhancements.

## Development Roadmap

### v0.1 - Core Foundation
- Go executable OTEL collector listening on ephemeral port
- Support for single OTEL protocol (OTLP/HTTP recommended)
- Basic build system using go build scripts only
- Minimal logging to stdout

### v0.2 - File Output
- Add file writing capability for collected telemetry data
- Append-only file output in structured format (JSON lines)
- Basic error handling for file operations

### v0.3 - Dual Storage
- Implement embedded SQLite database
- Write collected data to both file and SQLite simultaneously
- Database schema initialization

### v0.4 - Service Mode
- Linux systemd service integration
- Graceful shutdown handling
- Service installation and management scripts

### v0.5 - Execution Logging
- Add dedicated log file for service execution metadata
- Separate operational logs from telemetry data
- Log rotation and management capabilities

### v0.6 - SQLite Focus
- Remove file output for telemetry data (SQLite only)
- Maintain execution metadata logging to separate log file
- Performance optimizations for SQLite operations

### v0.7 - Distribution
- Cross-platform build system
- Package creation (RPM, DEB)
- CircleCI integration for automated builds and testing
- Release automation and CI/CD pipeline

### v1.0 - Public Release
- Comprehensive documentation
- Performance benchmarks
- Security review and hardening

## Feature Enhancements

### v1.1-1.3 - Configuration Options
- **v1.1**: SQLite file location argument (--db-path)
- **v1.2**: Log file location argument (--log-file)  
- **v1.3**: Logging mode argument (--log-level)

### v2.0 - Enhanced Public Release
- Complete command-line interface
- Configuration file support
- Advanced deployment options

## Stretch Goals

### v2.1+ - Extended Platform Support
- **Docker Containerization**: Multi-stage builds, minimal base images, health checks
- **Windows Service Support**: Native Windows service integration, MSI installer packages
- **Advanced Executable Arguments**: Configuration files, environment variable support, validation

### Future Considerations
- Multiple OTEL protocol support (gRPC, HTTP, custom)
- Horizontal scaling and clustering capabilities
- Web-based management interface
- Integration with major observability platforms