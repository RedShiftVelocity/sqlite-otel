#!/bin/bash

set -e

# PPA Build Script for SQLite OpenTelemetry Collector
# This script prepares source packages for upload to Ubuntu PPA using existing packaging/deb structure

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
PACKAGE_NAME="sqlite-otel-collector"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
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

# Check if required tools are installed
check_dependencies() {
    log_info "Checking required dependencies..."
    
    local missing_deps=()
    
    # Check for required packages
    for cmd in debuild dput gpg; do
        if ! command -v "$cmd" >/dev/null 2>&1; then
            missing_deps+=("$cmd")
        fi
    done
    
    # Check for Ubuntu development packages
    if ! dpkg -l 2>/dev/null | grep -q "ubuntu-dev-tools\|devscripts"; then
        missing_deps+=("devscripts")
    fi
    
    if [ ${#missing_deps[@]} -ne 0 ]; then
        log_error "Missing required dependencies: ${missing_deps[*]}"
        log_info "Install them with:"
        echo "  sudo apt-get update"
        echo "  sudo apt-get install -y ubuntu-dev-tools devscripts build-essential debhelper golang-go libsqlite3-dev gnupg dput-ng"
        exit 1
    fi
    
    log_success "All dependencies are installed"
}

# Get version information
get_version() {
    if [ -n "$1" ]; then
        VERSION="$1"
    else
        # Try to get version from git tag
        VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "0.8.0")
        # Remove leading 'v' if present
        VERSION=${VERSION#v}
    fi
    
    log_info "Using version: $VERSION"
}

# Setup temporary debian directory from packaging/deb
setup_debian_dir() {
    log_info "Setting up debian directory from packaging/deb..."
    
    cd "$PROJECT_ROOT"
    
    # Remove any existing debian directory
    rm -rf debian/
    
    # Copy packaging/deb to debian/
    cp -r packaging/deb/ debian/
    
    # Create missing files for PPA
    
    # Create debian/compat if it doesn't exist
    if [ ! -f debian/compat ]; then
        echo "13" > debian/compat
    fi
    
    # Create debian/source/format if it doesn't exist
    mkdir -p debian/source/
    if [ ! -f debian/source/format ]; then
        echo "3.0 (native)" > debian/source/format
    fi
    
    # Create debian/copyright if it doesn't exist
    if [ ! -f debian/copyright ]; then
        cat > debian/copyright << EOF
Format: https://www.debian.org/doc/packaging-manuals/copyright-format/1.0/
Upstream-Name: sqlite-otel-collector
Upstream-Contact: Manish Sinha <manishsinha.tech@gmail.com>
Source: https://github.com/RedShiftVelocity/sqlite-otel

Files: *
Copyright: 2025 Manish Sinha <manishsinha.tech@gmail.com>
License: MIT

License: MIT
 Permission is hereby granted, free of charge, to any person obtaining a copy
 of this software and associated documentation files (the "Software"), to deal
 in the Software without restriction, including without limitation the rights
 to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 copies of the Software, and to permit persons to whom the Software is
 furnished to do so, subject to the following conditions:
 .
 The above copyright notice and this permission notice shall be included in all
 copies or substantial portions of the Software.
 .
 THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 SOFTWARE.
EOF
    fi
    
    # Update debian/control for PPA requirements
    if grep -q "golang-go" debian/control; then
        log_info "debian/control already has golang-go dependency"
    else
        # Add golang-go to Build-Depends
        sed -i 's/Build-Depends: /Build-Depends: golang-go (>= 2:1.19~), /' debian/control
    fi
    
    # Make sure rules file is executable
    chmod +x debian/rules
    
    # Make sure maintainer scripts are executable if they exist
    for script in debian/*.postinst debian/*.prerm debian/*.postrm; do
        if [ -f "$script" ]; then
            chmod +x "$script"
        fi
    done
    
    log_success "Debian directory setup completed"
}

# Create or update changelog
create_changelog() {
    local ubuntu_release="$1"
    local ppa_version="$2"
    
    log_info "Creating changelog for Ubuntu $ubuntu_release..."
    
    cd "$PROJECT_ROOT"
    
    # Create changelog entry
    local changelog_entry="${PACKAGE_NAME} (${ppa_version}~${ubuntu_release}1) ${ubuntu_release}; urgency=medium"
    local temp_changelog=$(mktemp)
    
    {
        echo "$changelog_entry"
        echo ""
        echo "  * Build for Ubuntu $ubuntu_release"
        echo "  * OpenTelemetry collector with SQLite storage"
        echo "  * OTLP/HTTP endpoint support"
        echo "  * Systemd service integration"
        echo ""
        echo " -- $(git config user.name) <$(git config user.email)>  $(date -R)"
        echo ""
        
        # If existing changelog exists, append it
        if [ -f debian/changelog ]; then
            cat debian/changelog
        fi
    } > "$temp_changelog"
    
    mv "$temp_changelog" debian/changelog
    
    log_success "Changelog created for Ubuntu $ubuntu_release"
}

# Build source package for specific Ubuntu release
build_source_package() {
    local ubuntu_release="$1"
    local ppa_version="$2"
    
    log_info "Building source package for Ubuntu $ubuntu_release..."
    
    cd "$PROJECT_ROOT"
    
    # Create changelog for this release
    create_changelog "$ubuntu_release" "$ppa_version"
    
    # Build source package (use -d to override build dependency checks for source-only builds)
    debuild -S -sa -d -k"${GPG_KEY_ID:-}" || {
        log_error "Failed to build source package for $ubuntu_release"
        return 1
    }
    
    log_success "Source package built for Ubuntu $ubuntu_release"
}

# Upload to PPA
upload_to_ppa() {
    local ppa_name="$1"
    local changes_file="$2"
    
    if [ -z "$ppa_name" ]; then
        log_warning "No PPA name specified, skipping upload"
        return 0
    fi
    
    log_info "Uploading to PPA: $ppa_name"
    
    if [ ! -f "$changes_file" ]; then
        log_error "Changes file not found: $changes_file"
        return 1
    fi
    
    # Upload using dput
    dput "$ppa_name" "$changes_file" || {
        log_error "Failed to upload to PPA"
        return 1
    }
    
    log_success "Successfully uploaded to PPA: $ppa_name"
}

# Clean up
cleanup() {
    log_info "Cleaning up..."
    cd "$PROJECT_ROOT"
    rm -rf debian/
    log_success "Cleanup completed"
}

# Main build process
main() {
    local version=""
    local ppa_name=""
    # Default Ubuntu releases to build for
    # Currently only Jammy (22.04 LTS) due to distro-info compatibility
    # Noble (24.04 LTS) can be added when Launchpad build systems support it
    local ubuntu_releases=("jammy")
    local build_only=false
    local skip_cleanup=false
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -v|--version)
                version="$2"
                shift 2
                ;;
            -p|--ppa)
                ppa_name="$2"
                shift 2
                ;;
            -r|--release)
                ubuntu_releases=("$2")
                shift 2
                ;;
            --build-only)
                build_only=true
                shift
                ;;
            --no-cleanup)
                skip_cleanup=true
                shift
                ;;
            -h|--help)
                echo "Usage: $0 [OPTIONS]"
                echo ""
                echo "Options:"
                echo "  -v, --version VERSION    Package version (default: auto-detect from git)"
                echo "  -p, --ppa PPA_NAME      PPA to upload to (e.g., ppa:yourname/yourppa)"
                echo "  -r, --release RELEASE   Ubuntu release to build for (default: jammy)"
                echo "  --build-only            Only build source packages, don't upload"
                echo "  --no-cleanup            Don't remove debian/ directory after build"
                echo "  -h, --help              Show this help message"
                echo ""
                echo "Examples:"
                echo "  $0 --version 0.8.0 --ppa ppa:manishsinha/sqlite-otel"
                echo "  $0 --build-only --release focal"
                echo ""
                echo "Environment variables:"
                echo "  GPG_KEY_ID              GPG key ID for signing packages"
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    # Trap cleanup
    if [ "$skip_cleanup" = false ]; then
        trap cleanup EXIT
    fi
    
    # Header
    echo "======================================================"
    echo "  Ubuntu PPA Build Script"
    echo "  Package: $PACKAGE_NAME"
    echo "======================================================"
    
    # Check dependencies
    check_dependencies
    
    # Get version
    get_version "$version"
    
    # Setup debian directory
    setup_debian_dir
    
    # Build for each Ubuntu release
    local built_packages=()
    
    for release in "${ubuntu_releases[@]}"; do
        log_info "Processing Ubuntu $release..."
        
        local ppa_version="${VERSION}-1ubuntu1"
        
        if build_source_package "$release" "$ppa_version"; then
            local changes_file="../${PACKAGE_NAME}_${ppa_version}~${release}1_source.changes"
            built_packages+=("$release:$changes_file")
        else
            log_error "Failed to build package for $release"
            continue
        fi
    done
    
    # Upload packages if PPA is specified and not build-only
    if [ "$build_only" = false ] && [ -n "$ppa_name" ]; then
        log_info "Uploading packages to PPA..."
        
        for package_info in "${built_packages[@]}"; do
            local release="${package_info%:*}"
            local changes_file="${package_info#*:}"
            
            log_info "Uploading $release package..."
            upload_to_ppa "$ppa_name" "$changes_file"
        done
    fi
    
    # Summary
    echo ""
    echo "======================================================"
    log_success "Build process completed!"
    echo "======================================================"
    echo ""
    log_info "Built packages for releases: ${ubuntu_releases[*]}"
    
    if [ "$build_only" = true ]; then
        log_info "Source packages are ready for manual upload"
        log_info "Upload manually with: dput <ppa-name> <changes-file>"
    elif [ -n "$ppa_name" ]; then
        log_info "Packages uploaded to: $ppa_name"
        log_info "Check build status at: https://launchpad.net/~yourname/+archive/ubuntu/yourppa/+packages"
    else
        log_info "No PPA specified - packages built but not uploaded"
    fi
    
    echo ""
    log_info "Next steps:"
    echo "  1. Monitor build status on Launchpad"
    echo "  2. Test packages after successful builds"
    echo "  3. Update installation documentation with PPA instructions"
}

# Run main function
main "$@"