# Installation Guide

This guide covers all installation methods for SQLite OpenTelemetry Collector.

## 📋 Table of Contents

- [System Requirements](#system-requirements)
- [Quick Install](#quick-install)
- [Package Managers](#package-managers)
  - [Debian/Ubuntu](#debianubuntu)
  - [RHEL/CentOS/Fedora](#rhelcentosfedora)
  - [macOS](#macos)
- [Container Installation](#container-installation)
  - [Docker](#docker)
  - [Kubernetes](#kubernetes)
- [Binary Installation](#binary-installation)
- [Building from Source](#building-from-source)
- [Verification](#verification)
- [Uninstallation](#uninstallation)

## System Requirements

### Minimum Requirements

- **OS**: Linux (kernel 3.10+), macOS 10.15+, Windows 10+
- **Architecture**: amd64, arm64, armv7
- **Memory**: 128MB RAM
- **Disk**: 100MB free space (plus data storage)
- **SQLite**: 3.24.0+ (bundled with binary)

### Recommended Requirements

- **Memory**: 512MB RAM
- **Disk**: 1GB free space
- **Network**: Stable connection for telemetry ingestion

## Quick Install

### One-Line Install (Linux/macOS)

```bash
curl -sSL https://install.sqlite-otel.io | bash
```

This script:
- Detects your OS and architecture
- Downloads the appropriate binary
- Installs to `/usr/local/bin`
- Sets up systemd service (Linux)
- Creates necessary directories

## Package Managers

### Debian/Ubuntu

#### Using APT Repository

```bash
# 1. Add GPG key
curl -fsSL https://packages.sqlite-otel.io/gpg.key | sudo gpg --dearmor -o /usr/share/keyrings/sqlite-otel-archive-keyring.gpg

# 2. Add repository
echo "deb [signed-by=/usr/share/keyrings/sqlite-otel-archive-keyring.gpg] https://packages.sqlite-otel.io/deb stable main" | sudo tee /etc/apt/sources.list.d/sqlite-otel.list

# 3. Update and install
sudo apt update
sudo apt install sqlite-otel-collector

# 4. Start service
sudo systemctl enable --now sqlite-otel-collector
```

#### Using .deb Package

```bash
# Download latest .deb
wget https://github.com/RedShiftVelocity/sqlite-otel/releases/latest/download/sqlite-otel-collector_0.7.0_amd64.deb

# Install
sudo dpkg -i sqlite-otel-collector_0.7.0_amd64.deb

# Fix dependencies if needed
sudo apt-get install -f
```

#### Configuration

The Debian package:
- Installs binary to `/usr/bin/sqlite-otel-collector`
- Creates systemd service `sqlite-otel-collector.service`
- Creates user/group `sqlite-otel`
- Sets up directories:
  - `/var/lib/sqlite-otel-collector/` (data)
  - `/var/log/sqlite-otel-collector/` (logs)
  - `/etc/sqlite-otel/` (future config)

### RHEL/CentOS/Fedora

#### Using YUM/DNF Repository

```bash
# 1. Add repository
sudo tee /etc/yum.repos.d/sqlite-otel.repo << EOF
[sqlite-otel]
name=SQLite OTEL Collector
baseurl=https://packages.sqlite-otel.io/rpm/stable/\$basearch
enabled=1
gpgcheck=1
gpgkey=https://packages.sqlite-otel.io/gpg.key
EOF

# 2. Install
sudo yum install sqlite-otel-collector  # or dnf on Fedora

# 3. Start service
sudo systemctl enable --now sqlite-otel-collector
```

#### Using .rpm Package

```bash
# Download latest .rpm
wget https://github.com/RedShiftVelocity/sqlite-otel/releases/latest/download/sqlite-otel-collector-0.7.0-1.x86_64.rpm

# Install
sudo rpm -ivh sqlite-otel-collector-0.7.0-1.x86_64.rpm
```

### macOS

#### Using Homebrew

```bash
# Add tap
brew tap redshiftvelocity/sqlite-otel

# Install
brew install sqlite-otel-collector

# Start as service
brew services start sqlite-otel-collector

# Or run directly
sqlite-otel-collector
```

#### Manual macOS Install

```bash
# Download binary
curl -LO https://github.com/RedShiftVelocity/sqlite-otel/releases/latest/download/sqlite-otel-collector-darwin-amd64.tar.gz

# Extract
tar xzf sqlite-otel-collector-darwin-amd64.tar.gz

# Move to PATH
sudo mv sqlite-otel-collector /usr/local/bin/

# Make executable
sudo chmod +x /usr/local/bin/sqlite-otel-collector

# Create LaunchDaemon (optional)
sudo tee /Library/LaunchDaemons/io.sqlite-otel.collector.plist << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>io.sqlite-otel.collector</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/sqlite-otel-collector</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/var/log/sqlite-otel-collector.log</string>
    <key>StandardErrorPath</key>
    <string>/var/log/sqlite-otel-collector.error.log</string>
</dict>
</plist>
EOF

# Load service
sudo launchctl load /Library/LaunchDaemons/io.sqlite-otel.collector.plist
```

## Container Installation

### Docker

#### Docker Hub

```bash
# Pull latest image
docker pull redshiftvelocity/sqlite-otel:latest

# Run with persistent storage
docker run -d \
  --name sqlite-otel \
  -p 4318:4318 \
  -v sqlite-otel-data:/var/lib/sqlite-otel-collector \
  -v sqlite-otel-logs:/var/log \
  --restart unless-stopped \
  redshiftvelocity/sqlite-otel:latest
```

#### Docker Compose

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  sqlite-otel:
    image: redshiftvelocity/sqlite-otel:latest
    container_name: sqlite-otel-collector
    ports:
      - "4318:4318"
    volumes:
      - ./data:/var/lib/sqlite-otel-collector
      - ./logs:/var/log
    environment:
      - LOG_MAX_SIZE=50
      - LOG_MAX_BACKUPS=5
      - LOG_COMPRESS=true
    restart: unless-stopped
    user: "1000:1000"  # Run as non-root
    read_only: true     # Security hardening
    tmpfs:
      - /tmp
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:4318/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 10s
```

Run:

```bash
docker-compose up -d
```

### Kubernetes

#### Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sqlite-otel-collector
  namespace: monitoring
spec:
  replicas: 1
  selector:
    matchLabels:
      app: sqlite-otel-collector
  template:
    metadata:
      labels:
        app: sqlite-otel-collector
    spec:
      serviceAccountName: sqlite-otel-collector
      securityContext:
        runAsNonRoot: true
        runAsUser: 65534
        fsGroup: 65534
      containers:
      - name: collector
        image: redshiftvelocity/sqlite-otel:latest
        ports:
        - containerPort: 4318
          name: otlp-http
        env:
        - name: LOG_MAX_SIZE
          value: "50"
        volumeMounts:
        - name: data
          mountPath: /var/lib/sqlite-otel-collector
        - name: logs
          mountPath: /var/log
        resources:
          requests:
            memory: "256Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 4318
          initialDelaySeconds: 10
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /health
            port: 4318
          initialDelaySeconds: 5
          periodSeconds: 10
      volumes:
      - name: data
        persistentVolumeClaim:
          claimName: sqlite-otel-data
      - name: logs
        emptyDir: {}
---
apiVersion: v1
kind: Service
metadata:
  name: sqlite-otel-collector
  namespace: monitoring
spec:
  selector:
    app: sqlite-otel-collector
  ports:
  - port: 4318
    targetPort: 4318
    name: otlp-http
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: sqlite-otel-data
  namespace: monitoring
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
```

#### Helm Chart

```bash
# Add repository
helm repo add sqlite-otel https://charts.sqlite-otel.io
helm repo update

# Install
helm install sqlite-otel sqlite-otel/collector \
  --namespace monitoring \
  --create-namespace \
  --set persistence.enabled=true \
  --set persistence.size=10Gi
```

## Binary Installation

### Download Prebuilt Binaries

1. Visit [Releases](https://github.com/RedShiftVelocity/sqlite-otel/releases)
2. Download the appropriate archive for your platform:
   - `sqlite-otel-collector-linux-amd64.tar.gz`
   - `sqlite-otel-collector-linux-arm64.tar.gz`
   - `sqlite-otel-collector-darwin-amd64.tar.gz`
   - `sqlite-otel-collector-darwin-arm64.tar.gz`
   - `sqlite-otel-collector-windows-amd64.zip`

### Linux Installation

```bash
# Download
curl -LO https://github.com/RedShiftVelocity/sqlite-otel/releases/latest/download/sqlite-otel-collector-linux-amd64.tar.gz

# Extract
tar xzf sqlite-otel-collector-linux-amd64.tar.gz

# Install
sudo install -m 755 sqlite-otel-collector /usr/local/bin/

# Create systemd service
sudo tee /etc/systemd/system/sqlite-otel-collector.service << EOF
[Unit]
Description=SQLite OpenTelemetry Collector
After=network.target

[Service]
Type=simple
User=sqlite-otel
Group=sqlite-otel
ExecStart=/usr/local/bin/sqlite-otel-collector
Restart=always
RestartSec=5

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/sqlite-otel-collector /var/log

[Install]
WantedBy=multi-user.target
EOF

# Create user
sudo useradd --system --no-create-home --shell /bin/false sqlite-otel

# Create directories
sudo mkdir -p /var/lib/sqlite-otel-collector /var/log
sudo chown sqlite-otel:sqlite-otel /var/lib/sqlite-otel-collector

# Enable and start
sudo systemctl daemon-reload
sudo systemctl enable --now sqlite-otel-collector
```

## Building from Source

### Prerequisites

- Go 1.21 or higher
- Git
- Make (optional)

### Build Steps

```bash
# Clone repository
git clone https://github.com/RedShiftVelocity/sqlite-otel.git
cd sqlite-otel

# Build with make
make build

# Or build directly
go build -o sqlite-otel-collector \
  -ldflags="-s -w -X main.Version=$(git describe --tags --always)"

# Install
sudo install -m 755 sqlite-otel-collector /usr/local/bin/
```

### Cross-Compilation

```bash
# Build for multiple platforms
make build-all

# Or specific platform
GOOS=linux GOARCH=arm64 go build -o sqlite-otel-collector-linux-arm64
```

## Verification

### Check Installation

```bash
# Check version
sqlite-otel-collector --version

# Test run
sqlite-otel-collector --port 4319

# Check service status (systemd)
sudo systemctl status sqlite-otel-collector

# Check service logs
sudo journalctl -u sqlite-otel-collector -f
```

### Test Telemetry

```bash
# Send test trace
curl -X POST http://localhost:4318/v1/traces \
  -H "Content-Type: application/json" \
  -d '{
    "resourceSpans": [{
      "resource": {
        "attributes": [{
          "key": "service.name",
          "value": {"stringValue": "test-service"}
        }]
      },
      "scopeSpans": [{
        "spans": [{
          "name": "test-span",
          "kind": 1,
          "traceId": "7bba9f33312b3dbb8b2c2c62bb7abe2d",
          "spanId": "1b2c3d4e5f6g7h8i",
          "startTimeUnixNano": "1651234567890123456",
          "endTimeUnixNano": "1651234567890223456"
        }]
      }]
    }]
  }'

# Check database
sqlite3 /var/lib/sqlite-otel-collector/otel-collector.db "SELECT COUNT(*) FROM traces;"
```

## Uninstallation

### Package Manager Uninstall

```bash
# Debian/Ubuntu
sudo apt remove sqlite-otel-collector
sudo apt purge sqlite-otel-collector  # Also remove config

# RHEL/CentOS
sudo yum remove sqlite-otel-collector

# macOS
brew uninstall sqlite-otel-collector
brew untap redshiftvelocity/sqlite-otel
```

### Manual Uninstall

```bash
# Stop service
sudo systemctl stop sqlite-otel-collector
sudo systemctl disable sqlite-otel-collector

# Remove files
sudo rm /usr/local/bin/sqlite-otel-collector
sudo rm /etc/systemd/system/sqlite-otel-collector.service
sudo systemctl daemon-reload

# Remove data (optional)
sudo rm -rf /var/lib/sqlite-otel-collector
sudo rm -rf /var/log/sqlite-otel-collector*

# Remove user
sudo userdel sqlite-otel
```

### Docker Uninstall

```bash
# Stop and remove container
docker stop sqlite-otel
docker rm sqlite-otel

# Remove image
docker rmi redshiftvelocity/sqlite-otel:latest

# Remove volumes (optional - this deletes data!)
docker volume rm sqlite-otel-data sqlite-otel-logs
```

## Next Steps

- [Configuration Guide](../configuration/README.md) - Configure the collector
- [Deployment Guide](../deployment/README.md) - Production deployment patterns
- [Operations Guide](../operations/README.md) - Monitoring and maintenance