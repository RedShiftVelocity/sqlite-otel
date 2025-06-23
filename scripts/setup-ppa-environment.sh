#!/bin/bash

set -e

echo "🚀 Setting up Ubuntu PPA Development Environment"
echo "================================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running as root
if [ "$EUID" -eq 0 ]; then
    log_error "Don't run this script as root/sudo. It will prompt for sudo when needed."
    exit 1
fi

# Update package list
log_info "Updating package list..."
sudo apt-get update

# Install required packages
log_info "Installing Ubuntu PPA development tools..."
sudo apt-get install -y \
    ubuntu-dev-tools \
    devscripts \
    build-essential \
    debhelper \
    golang-go \
    libsqlite3-dev \
    gnupg \
    dput-ng

# Verify installation
log_info "Verifying installation..."

missing_tools=()
for tool in debuild dput dpkg-buildpackage; do
    if ! command -v "$tool" >/dev/null 2>&1; then
        missing_tools+=("$tool")
    fi
done

if [ ${#missing_tools[@]} -ne 0 ]; then
    log_error "Some tools are still missing: ${missing_tools[*]}"
    exit 1
fi

log_success "All required tools installed successfully!"

# Set up environment variables
log_info "Setting up environment variables..."

# Add to .bashrc if not already present
if ! grep -q "GPG_KEY_ID.*DE19D310F39E5FCB" ~/.bashrc; then
    echo "" >> ~/.bashrc
    echo "# PPA Development Environment" >> ~/.bashrc
    echo 'export GPG_KEY_ID="DE19D310F39E5FCB"' >> ~/.bashrc
    echo 'export DEBEMAIL="manishsinha.tech@gmail.com"' >> ~/.bashrc
    echo 'export DEBFULLNAME="Manish Sinha"' >> ~/.bashrc
    log_success "Environment variables added to ~/.bashrc"
else
    log_info "Environment variables already present in ~/.bashrc"
fi

# Set for current session
export GPG_KEY_ID="DE19D310F39E5FCB"
export DEBEMAIL="manishsinha.tech@gmail.com"
export DEBFULLNAME="Manish Sinha"

log_success "Environment variables set for current session"

# Verify GPG key
log_info "Verifying GPG key setup..."
if gpg --list-secret-keys | grep -q "DE19D310F39E5FCB"; then
    log_success "GPG key DE19D310F39E5FCB is available"
else
    log_error "GPG key DE19D310F39E5FCB not found"
    exit 1
fi

# Test build environment
log_info "Testing build environment..."
echo "Current environment:"
echo "  GPG_KEY_ID: $GPG_KEY_ID"
echo "  DEBEMAIL: $DEBEMAIL"
echo "  DEBFULLNAME: $DEBFULLNAME"
echo "  Go version: $(go version)"

log_success "PPA development environment setup complete!"
echo ""
log_info "Next steps:"
echo "  1. Build packages with: ./scripts/build-for-ppa.sh --version 0.8.0 --ppa ppa:manishsinha/sqlite-ote"
echo "  2. Monitor build status at: https://launchpad.net/~manishsinha/+archive/ubuntu/sqlite-ote"
echo ""
log_info "For new terminal sessions, run: source ~/.bashrc"