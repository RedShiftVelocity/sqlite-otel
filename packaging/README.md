# RPM Packaging for SQLite OTEL Collector

This directory contains RPM packaging scripts and configurations.

## Overview

The packaging scripts create installable RPM packages for:
- **RHEL** (Red Hat Enterprise Linux) 7+
- **CentOS** 7+
- **Fedora** 30+
- **openSUSE** Leap 15+

## Features

- Automatic user/group creation (`sqlite-otel`)
- Systemd service integration with auto-start
- Security hardening with restricted permissions
- Proper file system hierarchy compliance
- Pre/post installation scripts

## Prerequisites

```bash
# Install RPM build tools
sudo yum install -y rpm-build rpmdevtools    # RHEL/CentOS 7/8
sudo dnf install -y rpm-build rpmdevtools    # Fedora/RHEL 9+
sudo zypper install -y rpm-build             # openSUSE
```

## Building the RPM Package

```bash
# From the project root directory
make package-rpm
```

The built RPM will be placed in `dist/rpm/`

## Installation

```bash
# Install the RPM package
sudo rpm -ivh dist/rpm/sqlite-otel-collector-*.rpm

# Or using yum/dnf
sudo yum install dist/rpm/sqlite-otel-collector-*.rpm
sudo dnf install dist/rpm/sqlite-otel-collector-*.rpm
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
sudo rpm -e sqlite-otel-collector

# Or using yum/dnf
sudo yum remove sqlite-otel-collector
sudo dnf remove sqlite-otel-collector
```

## Customization

The RPM spec file can be customized in `rpm/sqlite-otel-collector.spec` for:
- Different installation paths
- Additional dependencies
- Custom pre/post scripts
- Configuration file packaging