# Packaging for SQLite OTEL Collector

This directory contains packaging scripts and configurations for building RPM and DEB packages.

## Overview

The packaging scripts create installable packages for:
- **RPM-based systems** (RHEL, CentOS, Fedora, openSUSE)
- **DEB-based systems** (Debian, Ubuntu)

## Features

- Automatic user/group creation (`sqlite-otel`)
- Systemd service integration
- Security hardening with restricted permissions
- Proper file system hierarchy compliance
- Pre/post installation scripts

## Building Packages

### Prerequisites

#### For RPM:
```bash
# Install RPM build tools
sudo yum install -y rpm-build rpmdevtools    # RHEL/CentOS
sudo dnf install -y rpm-build rpmdevtools    # Fedora
sudo zypper install -y rpm-build             # openSUSE
```

#### For DEB:
```bash
# Install DEB build tools
sudo apt-get install -y build-essential debhelper dh-systemd devscripts
```

### Build Commands

```bash
# Build RPM package
make package-rpm

# Build DEB package  
make package-deb

# Build all packages
make package-all
```

Built packages will be placed in:
- `dist/rpm/` - RPM packages
- `dist/deb/` - DEB packages

## Installation

### RPM-based systems:
```bash
sudo rpm -ivh dist/rpm/sqlite-otel-collector-*.rpm
# or
sudo yum install dist/rpm/sqlite-otel-collector-*.rpm
```

### DEB-based systems:
```bash
sudo dpkg -i dist/deb/sqlite-otel-collector_*.deb
# or
sudo apt install ./dist/deb/sqlite-otel-collector_*.deb
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
- **Logs**: `/var/log/sqlite-otel-collector.log` (when configured)
- **Config**: `/etc/sqlite-otel-collector/` (future use)

## Security

The service runs as the `sqlite-otel` user with:
- No new privileges
- Private tmp directory
- Read-only system access (except specified paths)
- Protected home directories
- Restricted network families (IPv4/IPv6 only)
- No capabilities

## Uninstallation

### RPM:
```bash
sudo rpm -e sqlite-otel-collector
```

### DEB:
```bash
sudo apt-get remove sqlite-otel-collector
```