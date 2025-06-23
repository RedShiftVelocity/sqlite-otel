# Ubuntu PPA Setup Guide

This guide walks you through creating and managing a Personal Package Archive (PPA) on Launchpad for the SQLite OpenTelemetry Collector.

## Prerequisites

### 1. Launchpad Account
- Create account at https://launchpad.net/
- Verify your email address
- Complete your profile

### 2. GPG Key Setup

Generate and upload a GPG key to Launchpad:

```bash
# Generate GPG key (if you don't have one)
gpg --full-generate-key

# List your keys to get the key ID
gpg --list-secret-keys --keyid-format LONG

# Export your public key
gpg --armor --export YOUR_KEY_ID

# Copy the output and paste it into your Launchpad profile at:
# https://launchpad.net/~yourusername/+editpgpkeys
```

### 3. Install Development Tools

```bash
sudo apt-get update
sudo apt-get install -y \
    ubuntu-dev-tools \
    devscripts \
    build-essential \
    debhelper \
    golang-go \
    libsqlite3-dev \
    gnupg \
    dh-systemd \
    python3-launchpadlib
```

## Step 1: Create the PPA

Run the automated PPA creation script:

```bash
cd /path/to/sqlite-otel
python3 scripts/create-ppa.py
```

This will:
1. Authenticate with Launchpad (opens browser for OAuth)
2. Create a PPA named `sqlite-otel`
3. Configure the PPA settings

**Alternative: Manual Creation**

If you prefer to create the PPA manually:
1. Go to https://launchpad.net/~yourusername/+activate-ppa
2. Fill in PPA details:
   - **Name**: `sqlite-otel`
   - **Display name**: `SQLite OpenTelemetry Collector`
   - **Description**: See example below

```
Personal Package Archive for SQLite OpenTelemetry Collector

This PPA provides packages for the SQLite OpenTelemetry Collector, a lightweight 
OpenTelemetry collector that stores telemetry data in SQLite databases.

Features:
- OTLP/HTTP endpoint support for traces, metrics, and logs
- SQLite database storage for efficient data management
- Systemd service integration with security hardening
- Support for Ubuntu 20.04 LTS, 22.04 LTS, 23.10, and 24.04 LTS
- Multi-architecture support (amd64, arm64, armhf)

Homepage: https://github.com/RedShiftVelocity/sqlite-otel
```

## Step 2: Build and Upload Packages

### Set Environment Variables

```bash
export GPG_KEY_ID="YOUR_GPG_KEY_ID"
export DEBEMAIL="your.email@example.com"
export DEBFULLNAME="Your Full Name"
```

### Build for All Ubuntu Releases

```bash
# Build and upload to your PPA
./scripts/build-for-ppa.sh \
    --version 0.8.0 \
    --ppa ppa:yourusername/sqlite-otel
```

### Build for Specific Release

```bash
# Build only for Ubuntu 22.04 LTS
./scripts/build-for-ppa.sh \
    --version 0.8.0 \
    --release jammy \
    --ppa ppa:yourusername/sqlite-otel
```

### Build Without Uploading

```bash
# Build packages locally for testing
./scripts/build-for-ppa.sh \
    --version 0.8.0 \
    --build-only
```

## Step 3: Monitor Build Status

1. Go to your PPA page: https://launchpad.net/~yourusername/+archive/ubuntu/sqlite-otel
2. Check build status for each Ubuntu release
3. View build logs if there are failures

## Step 4: Test Installation

Once builds are successful, test the installation:

```bash
# Add your PPA
sudo add-apt-repository ppa:yourusername/sqlite-otel
sudo apt update

# Install the package
sudo apt install sqlite-otel-collector

# Test the service
sudo systemctl start sqlite-otel-collector
sudo systemctl status sqlite-otel-collector

# Test functionality
curl http://localhost:4318/health
```

## Troubleshooting

### Common Build Failures

1. **Missing Build Dependencies**
   - Update `packaging/deb/control` to include required dependencies
   - Ensure `golang-go` and `libsqlite3-dev` are in `Build-Depends`

2. **GPG Signing Issues**
   - Verify your GPG key is uploaded to Launchpad
   - Check that `GPG_KEY_ID` environment variable is set correctly
   - Ensure your key hasn't expired

3. **Version Conflicts**
   - Use unique version numbers for each upload
   - Follow the pattern: `version-1ubuntu1~release1`

### Checking Build Logs

If a build fails:
1. Go to your PPA page on Launchpad
2. Click on the failed build
3. View the build log to identify the issue
4. Fix the issue and upload a new version

### Manual Package Upload

If the automated script fails, you can upload manually:

```bash
# Build source package
./scripts/build-for-ppa.sh --build-only --release jammy

# Upload manually
dput ppa:yourusername/sqlite-otel ../sqlite-otel-collector_*_source.changes
```

## Managing the PPA

### Updating Packages

To release a new version:

1. Update the version in your code
2. Commit and tag the release
3. Run the build script with the new version:

```bash
./scripts/build-for-ppa.sh --version 0.9.0 --ppa ppa:yourusername/sqlite-otel
```

### Deleting Packages

To remove a package from the PPA:
1. Go to your PPA page on Launchpad
2. Click "View package details"
3. Click "Delete packages"

### PPA Settings

Configure PPA settings at: https://launchpad.net/~yourusername/+archive/ubuntu/sqlite-otel/+edit

- **Description**: Update as needed
- **Dependencies**: Leave as "Primary Archive for Ubuntu"
- **Publishing**: Keep enabled
- **Commercial**: Leave disabled (for open source projects)

## Security Considerations

1. **Keep GPG Keys Secure**: Store private keys safely
2. **Version Control**: Don't commit GPG keys to version control
3. **Access Control**: Only grant PPA access to trusted maintainers
4. **Package Signing**: All packages must be signed with uploaded GPG key

## Advanced Configuration

### Multiple Maintainers

To add team members to your PPA:
1. Go to your PPA page
2. Click "Change details"
3. Add team members under "Authorized uploaders"

### Custom Dependencies

If your package needs dependencies from other PPAs:
1. Go to PPA settings
2. Add dependency PPAs under "Dependencies"

### Build Recipe

For automated builds from source control:
1. Create a build recipe
2. Link to your GitHub repository
3. Configure automatic builds on new tags

## Getting Help

- **Launchpad Help**: https://help.launchpad.net/
- **Ubuntu Packaging Guide**: https://packaging.ubuntu.com/
- **Debian Policy Manual**: https://www.debian.org/doc/debian-policy/
- **Ask Ubuntu**: https://askubuntu.com/ (tag with 'launchpad', 'ppa')

## Next Steps

After successfully creating your PPA:

1. **Documentation**: Update project README with PPA installation instructions
2. **CI/CD**: Consider automating PPA uploads in your release pipeline  
3. **Testing**: Set up automated testing of PPA packages
4. **Monitoring**: Monitor download statistics and user feedback
5. **Maintenance**: Plan for regular updates and security patches