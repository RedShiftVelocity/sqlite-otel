# Packaging for SQLite OTEL Collector

This directory contains packaging scripts and configurations for multiple package formats.

## Overview

The packaging scripts create installable packages for:

### DEB Packages (Debian/Ubuntu)
- **Debian** 10 (Buster) and later
- **Ubuntu** 20.04 LTS (Focal) and later
- Other Debian-based distributions

### RPM Packages (Red Hat/SUSE)
- **RHEL** (Red Hat Enterprise Linux) 7+
- **CentOS** 7+
- **Fedora** 30+
- **openSUSE** Leap 15+

## Features

- Automatic user/group creation (`sqlite-otel`)
- Systemd service integration with auto-start
- Security hardening with restricted permissions
- Proper packaging policy compliance (Debian/RPM)
- Pre/post installation scripts
- Clean uninstallation

## Prerequisites

### For DEB Packages
```bash
# Install DEB build tools
sudo apt-get update
sudo apt-get install -y build-essential debhelper dh-systemd devscripts
```

### For RPM Packages
```bash
# Install RPM build tools
sudo yum install -y rpm-build rpmdevtools    # RHEL/CentOS 7/8
sudo dnf install -y rpm-build rpmdevtools    # Fedora/RHEL 9+
sudo zypper install -y rpm-build             # openSUSE
```

## Building Packages

### Build DEB Package
```bash
# From the project root directory
make package-deb
```
The built DEB will be placed in `dist/deb/`

### Build RPM Package
```bash
# From the project root directory
make package-rpm
```
The built RPM will be placed in `dist/rpm/`

## Installation

### DEB Package Installation (Debian/Ubuntu)
```bash
# Install the DEB package
sudo dpkg -i dist/deb/sqlite-otel-collector_*.deb

# Or using apt (handles dependencies automatically)
sudo apt install ./dist/deb/sqlite-otel-collector_*.deb
```

If there are dependency issues with dpkg:
```bash
# Fix dependencies
sudo apt-get install -f
```

### RPM Package Installation (RHEL/Fedora/SUSE)
```bash
# Install the RPM package
sudo rpm -ivh dist/rpm/sqlite-otel-collector-*.rpm

# Or using package managers (handles dependencies automatically)
sudo yum install dist/rpm/sqlite-otel-collector-*.rpm      # RHEL/CentOS 7/8
sudo dnf install dist/rpm/sqlite-otel-collector-*.rpm      # Fedora/RHEL 9+
sudo zypper install dist/rpm/sqlite-otel-collector-*.rpm   # openSUSE
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

### DEB Package Removal (Debian/Ubuntu)
```bash
# Remove the package
sudo apt-get remove sqlite-otel-collector

# Remove package and configuration files
sudo apt-get purge sqlite-otel-collector

# Remove automatically installed dependencies
sudo apt-get autoremove
```

### RPM Package Removal (RHEL/Fedora/SUSE)
```bash
# Remove the package
sudo rpm -e sqlite-otel-collector

# Or using package managers
sudo yum remove sqlite-otel-collector     # RHEL/CentOS 7/8
sudo dnf remove sqlite-otel-collector     # Fedora/RHEL 9+
sudo zypper remove sqlite-otel-collector  # openSUSE
```

## Package Information

### DEB Package Information
View package details:
```bash
# Before installation
dpkg -I dist/deb/sqlite-otel-collector_*.deb

# After installation
dpkg -l | grep sqlite-otel-collector
apt show sqlite-otel-collector
```

### RPM Package Information
View package details:
```bash
# Before installation
rpm -qip dist/rpm/sqlite-otel-collector-*.rpm

# After installation
rpm -qi sqlite-otel-collector
rpm -ql sqlite-otel-collector  # List files
```

## Customization

### DEB Package Customization
The Debian packaging files can be customized:
- `deb/control` - Package metadata and dependencies
- `deb/rules` - Build rules
- `deb/preinst` - Pre-installation script
- `deb/postinst` - Post-installation script
- `deb/prerm` - Pre-removal script
- `deb/postrm` - Post-removal script

### RPM Package Customization
The RPM spec file can be customized in `rpm/sqlite-otel-collector.spec` for:
- Different installation paths
- Additional dependencies
- Custom pre/post scripts
- Configuration file packaging
- Build requirements and build options
- File permissions and ownership
