# Homebrew Support for SQLite OpenTelemetry Collector

This directory contains the Homebrew formula for SQLite OpenTelemetry Collector.

## Installation

### Option 1: Direct Formula Installation (Recommended)

```bash
# Download and install the formula locally
curl -O https://raw.githubusercontent.com/RedShiftVelocity/sqlite-otel/main/homebrew/sqlite-otel-collector.rb
brew install --build-from-source ./sqlite-otel-collector.rb

# Alternatively, for development/testing (from local repository)
brew install --build-from-source ./homebrew/sqlite-otel-collector.rb
```

**Note**: Homebrew requires local formula files for non-checksummed installs. The download step ensures you have the latest formula locally.

### Option 2: Custom Tap (Future)

Once we create a custom Homebrew tap, installation will be:

```bash
# Add our tap (future)
brew tap redshiftvelocity/sqlite-otel

# Install from tap (future)  
brew install sqlite-otel-collector
```

## Usage

### Start the Collector

```bash
# Run directly
sqlite-otel-collector

# Run as background service
brew services start sqlite-otel-collector

# Check service status
brew services list | grep sqlite-otel
```

### Send Test Data

```bash
# Test traces endpoint
curl -X POST http://localhost:4318/v1/traces \
  -H "Content-Type: application/json" \
  -d '{"resourceSpans":[{"spans":[{"name":"test-span","kind":1}]}]}'

# Test metrics endpoint  
curl -X POST http://localhost:4318/v1/metrics \
  -H "Content-Type: application/json" \
  -d '{"resourceMetrics":[{"metrics":[{"name":"test-metric"}]}]}'
```

### Configuration

The collector accepts the same configuration options as the standalone binary:

```bash
# Custom database path
sqlite-otel-collector --db-path /custom/path/telemetry.db

# Custom port
sqlite-otel-collector --port 8080

# Enable debug logging
sqlite-otel-collector --log-level debug
```

## File Locations

When installed via Homebrew:

- **Binary**: `/opt/homebrew/bin/sqlite-otel-collector` (Apple Silicon) or `/usr/local/bin/sqlite-otel-collector` (Intel)
- **Data Directory**: `/opt/homebrew/var/lib/sqlite-otel-collector/` (Apple Silicon) or `/usr/local/var/lib/sqlite-otel-collector/` (Intel)
- **Logs**: `/opt/homebrew/var/log/sqlite-otel-collector.log` (Apple Silicon) or `/usr/local/var/log/sqlite-otel-collector.log` (Intel)
- **Service**: Managed by `brew services`

## Service Management

```bash
# Start service
brew services start sqlite-otel-collector

# Stop service  
brew services stop sqlite-otel-collector

# Restart service
brew services restart sqlite-otel-collector

# View service status
brew services list | grep sqlite-otel-collector

# View logs
tail -f $(brew --prefix)/var/log/sqlite-otel-collector.log
```

## Uninstalling

```bash
# Stop service if running
brew services stop sqlite-otel-collector

# Uninstall formula
brew uninstall sqlite-otel-collector

# Optional: Remove data directory
rm -rf $(brew --prefix)/var/lib/sqlite-otel-collector/
```

## Platform Support

The formula supports:

- **macOS**: Intel (x86_64) and Apple Silicon (ARM64)
- **Linux**: x86_64, ARM64, and ARM (via Homebrew on Linux)

## Updating

To update to a new version:

```bash
# If installed from direct URL, reinstall
brew uninstall sqlite-otel-collector
brew install --formula https://raw.githubusercontent.com/RedShiftVelocity/sqlite-otel/main/homebrew/sqlite-otel-collector.rb

# Future: Update from tap
brew update && brew upgrade sqlite-otel-collector
```

## Development

### Testing the Formula

```bash
# Install locally for testing
brew install --build-from-source ./homebrew/sqlite-otel-collector.rb

# Run formula audit
brew audit --strict ./homebrew/sqlite-otel-collector.rb

# Test installation
brew test sqlite-otel-collector
```

### Creating a Custom Tap

To create a dedicated Homebrew tap in the future:

1. Create repository: `redshiftvelocity/homebrew-sqlite-otel`
2. Move formula to: `Formula/sqlite-otel-collector.rb`
3. Users can then: `brew tap redshiftvelocity/sqlite-otel`

## Troubleshooting

### Permission Issues

```bash
# Fix data directory permissions
sudo chown -R $(whoami) $(brew --prefix)/var/lib/sqlite-otel-collector/
```

### Port Conflicts

```bash
# Check what's using port 4318
lsof -i :4318

# Use alternative port
sqlite-otel-collector --port 8080
```

### Service Won't Start

```bash
# Check service logs
brew services list
tail -f $(brew --prefix)/var/log/sqlite-otel-collector.log

# Restart service
brew services restart sqlite-otel-collector
```

## Support

- **Issues**: [GitHub Issues](https://github.com/RedShiftVelocity/sqlite-otel/issues)
- **Documentation**: [Project README](https://github.com/RedShiftVelocity/sqlite-otel)
- **Releases**: [GitHub Releases](https://github.com/RedShiftVelocity/sqlite-otel/releases)