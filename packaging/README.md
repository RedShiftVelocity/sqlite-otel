# DEB Packaging for SQLite OTEL Collector

This directory contains Debian packaging scripts and configurations.

## Overview

The packaging scripts create installable DEB packages for:
- **Debian** 10 (Buster) and later
- **Ubuntu** 20.04 LTS (Focal) and later
- Other Debian-based distributions

## Features

- Automatic user/group creation (`sqlite-otel`)
- Systemd service integration with auto-start
- Security hardening with restricted permissions
- Proper Debian policy compliance
- Pre/post installation scripts
- Clean uninstallation

## Prerequisites

```bash
# Install DEB build tools
sudo apt-get update
sudo apt-get install -y build-essential debhelper dh-systemd devscripts
```

## Building the DEB Package

```bash
# From the project root directory
make package-deb
```

The built DEB will be placed in `dist/deb/`

## Installation

```bash
# Install the DEB package
sudo dpkg -i dist/deb/sqlite-otel-collector_*.deb

# Or using apt (handles dependencies)
sudo apt install ./dist/deb/sqlite-otel-collector_*.deb
```

If there are dependency issues:
```bash
# Fix dependencies
sudo apt-get install -f
```

## Service Management

After installation, the service can be managed with systemctl:

```bash
# Start the service
sudo systemctl start sqlite-otel-collector

# Enable on boot
sudo systemctl enable sqlite-otel-collector

# Check status
sudo systemctl status sqlite-otel-collector

# View logs
sudo journalctl -u sqlite-otel-collector -f
```

## File Locations

- **Binary**: `/usr/bin/sqlite-otel-collector`
- **Service**: `/lib/systemd/system/sqlite-otel-collector.service`
- **Database**: `/var/lib/sqlite-otel-collector/otel-collector.db`
- **Logs**: Via journald (systemd journal)

## Security

The service runs as the `sqlite-otel` user with comprehensive security hardening:
- No new privileges
- Private tmp directory
- Read-only system access (except specified paths)
- Protected kernel tunables and modules
- Restricted system calls
- Memory execution protection
- Device access restrictions

## Uninstallation

```bash
# Remove the package
sudo apt-get remove sqlite-otel-collector

# Remove package and configuration files
sudo apt-get purge sqlite-otel-collector

# Remove automatically installed dependencies
sudo apt-get autoremove
```

## Package Information

View package details:
```bash
# Before installation
dpkg -I dist/deb/sqlite-otel-collector_*.deb

# After installation
dpkg -l | grep sqlite-otel-collector
apt show sqlite-otel-collector
```

## Customization

The Debian packaging files can be customized:
- `deb/control` - Package metadata and dependencies
- `deb/rules` - Build rules
- `deb/preinst` - Pre-installation script
- `deb/postinst` - Post-installation script