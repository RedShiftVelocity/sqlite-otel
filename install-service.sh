#!/bin/bash
# install-service.sh - Install sqlite-otel-collector as a systemd service

set -e

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo "Please run as root (use sudo)"
    exit 1
fi

echo "Installing sqlite-otel-collector service..."

# Build the binary with optimizations
echo "Building sqlite-otel binary..."
go build -ldflags="-s -w" -o sqlite-otel .

# Create system user for the service
if ! id -u sqlite-otel >/dev/null 2>&1; then
    echo "Creating sqlite-otel system user..."
    useradd --system --no-create-home --shell /bin/false sqlite-otel
fi

# Install binary
echo "Installing binary to /usr/local/bin..."
install -m 755 sqlite-otel /usr/local/bin/

# Create data directory
echo "Creating data directory..."
mkdir -p /var/lib/sqlite-otel-collector
chown sqlite-otel:sqlite-otel /var/lib/sqlite-otel-collector

# Install systemd service
echo "Installing systemd service..."
install -m 644 sqlite-otel-collector.service /etc/systemd/system/

# Reload systemd and enable service
echo "Enabling service..."
systemctl daemon-reload
systemctl enable sqlite-otel-collector.service

echo "Installation complete!"
echo ""
echo "To start the service: sudo systemctl start sqlite-otel-collector"
echo "To check status: sudo systemctl status sqlite-otel-collector"
echo "To view logs: sudo journalctl -u sqlite-otel-collector -f"