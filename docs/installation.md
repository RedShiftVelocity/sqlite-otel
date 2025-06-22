# Installation

## System Requirements

- **OS**: Linux, macOS, or Windows
- **Memory**: Minimum 64MB RAM
- **Storage**: 10MB for binary + space for telemetry data
- **Network**: Port 4318 (configurable)

## Binary Installation

### Linux

```bash
# AMD64
wget https://github.com/RedShiftVelocity/sqlite-otel/releases/latest/download/sqlite-otel-linux-amd64
chmod +x sqlite-otel-linux-amd64
sudo mv sqlite-otel-linux-amd64 /usr/local/bin/sqlite-otel

# ARM64
wget https://github.com/RedShiftVelocity/sqlite-otel/releases/latest/download/sqlite-otel-linux-arm64
chmod +x sqlite-otel-linux-arm64
sudo mv sqlite-otel-linux-arm64 /usr/local/bin/sqlite-otel
```

### macOS

```bash
# Intel
wget https://github.com/RedShiftVelocity/sqlite-otel/releases/latest/download/sqlite-otel-darwin-amd64
chmod +x sqlite-otel-darwin-amd64
sudo mv sqlite-otel-darwin-amd64 /usr/local/bin/sqlite-otel

# Apple Silicon
wget https://github.com/RedShiftVelocity/sqlite-otel/releases/latest/download/sqlite-otel-darwin-arm64
chmod +x sqlite-otel-darwin-arm64
sudo mv sqlite-otel-darwin-arm64 /usr/local/bin/sqlite-otel
```

## Docker Installation

!!! info "Recommended for Production"
    Docker provides the easiest deployment method with built-in security and isolation.

```bash
# Pull latest development image
docker pull ghcr.io/redshiftvelocity/sqlite-otel:latest

# Run with default settings
docker run -d --name sqlite-otel -p 4318:4318 \
  ghcr.io/redshiftvelocity/sqlite-otel:latest

# Run with persistent storage
docker run -d --name sqlite-otel -p 4318:4318 \
  -v sqlite-otel-data:/var/lib/sqlite-otel-collector \
  ghcr.io/redshiftvelocity/sqlite-otel:latest
```

## Package Installation

### Debian/Ubuntu (.deb)

```bash
# Download and install
wget https://github.com/RedShiftVelocity/sqlite-otel/releases/latest/download/sqlite-otel-collector_amd64.deb
sudo dpkg -i sqlite-otel-collector_amd64.deb

# Start service
sudo systemctl enable --now sqlite-otel-collector
```

### RHEL/CentOS/Fedora (.rpm)

```bash
# Download and install
wget https://github.com/RedShiftVelocity/sqlite-otel/releases/latest/download/sqlite-otel-collector-amd64.rpm
sudo rpm -ivh sqlite-otel-collector-amd64.rpm

# Start service
sudo systemctl enable --now sqlite-otel-collector
```

## Verification

After installation, verify the collector is working:

```bash
# Check version
sqlite-otel --version

# Test health endpoint
curl http://localhost:4318/health
```

!!! success "Installation Complete"
    Your SQLite OTEL Collector is now ready to receive telemetry data!

[Next: Quick Start →](quickstart.md){ .md-button .md-button--primary }