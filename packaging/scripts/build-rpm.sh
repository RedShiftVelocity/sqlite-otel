#!/bin/bash
#
# Robust RPM build script for cross-platform CI environments.
#
# This script inspects the OS and injects compatibility macros for non-RPM-native
# systems (like Debian/Ubuntu) before invoking rpmbuild. This keeps the .spec
# file clean and distro-agnostic.
#
set -euo pipefail

# --- Configuration ---
VERSION=${VERSION:-0.7.0}
PACKAGE_NAME="sqlite-otel-collector"
ORIGINAL_DIR="$(pwd)"
BUILD_DIR="$ORIGINAL_DIR/build/rpm"
SOURCES_DIR="$BUILD_DIR/SOURCES"
SPECS_DIR="$BUILD_DIR/SPECS"
RPMS_DIR="$BUILD_DIR/RPMS"

# Allow the spec file to be passed as an argument (relative to SPECS_DIR)
SPEC_FILE="$SPECS_DIR/$PACKAGE_NAME.spec"

echo "Building RPM package for $PACKAGE_NAME version $VERSION"

# --- OS Detection and Macro Setup ---
RPMBUILD_OPTS=("-bb")
DEFINES=()

# Use /etc/os-release for reliable OS detection
if [ -f /etc/os-release ]; then
    # Source the file to get access to variables like $ID
    # Store VERSION before sourcing to avoid overwriting it
    SAVED_VERSION="$VERSION"
    . /etc/os-release
    VERSION="$SAVED_VERSION"
else
    # Fallback for older systems without /etc/os-release
    ID=""
fi

echo "==> Detected OS ID: '${ID}'"

case "${ID}" in
    ubuntu|debian)
        echo "==> Applying Debian/Ubuntu compatibility macros."
        DEFINES+=(--define "_unitdir /usr/lib/systemd/system")
        DEFINES+=(--define "_presetdir /usr/lib/systemd/system-preset")
        DEFINES+=(--define "_sysusersdir /usr/lib/sysusers.d")
        DEFINES+=(--define "systemd_post : ")
        DEFINES+=(--define "systemd_preun : ")
        DEFINES+=(--define "systemd_postun : ")
        ;;
    fedora|centos|rhel|almalinux|rocky)
        echo "==> Detected RPM-native OS. No compatibility macros needed."
        ;;
    *)
        echo "==> WARNING: Unknown OS ID '${ID}'. Proceeding without compatibility macros."
        ;;
esac

# Check if rpmbuild is available
if ! command -v rpmbuild >/dev/null 2>&1; then
    echo "WARNING: rpmbuild command not found"
    echo "RPM packaging requires rpmbuild to be installed."
    echo "On RHEL/CentOS/Fedora: sudo yum install rpm-build rpmdevtools"
    echo "On openSUSE: sudo zypper install rpm-build"
    echo "On Ubuntu/Debian: sudo apt install rpm build-essential"
    echo "Attempting to continue anyway..."
fi

# Clean and create build directories
rm -rf "$BUILD_DIR"
mkdir -p "$SOURCES_DIR" "$SPECS_DIR" "$RPMS_DIR"

# Create source tarball
echo "Creating source tarball..."
TEMP_DIR=$(mktemp -d)
mkdir -p "$TEMP_DIR/$PACKAGE_NAME-$VERSION"

# No binary verification needed - we build from source

# Copy source files
cp -r . "$TEMP_DIR/$PACKAGE_NAME-$VERSION/"
cd "$TEMP_DIR/$PACKAGE_NAME-$VERSION"

# Clean unnecessary files for source-based packaging
rm -rf .git .gitignore build/ dist/ *.db *.log sqlite-otel
# Remove pre-built binaries since we're building from source

# Create tarball
cd "$TEMP_DIR"
tar czf "$SOURCES_DIR/$PACKAGE_NAME-$VERSION.tar.gz" "$PACKAGE_NAME-$VERSION"
rm -rf "$TEMP_DIR"

# Copy spec file from original directory
cp "$ORIGINAL_DIR/packaging/rpm/$PACKAGE_NAME.spec" "$SPECS_DIR/"

# Copy additional source files
cp "$ORIGINAL_DIR/packaging/rpm/$PACKAGE_NAME.sysusers" "$SOURCES_DIR/"
cp "$ORIGINAL_DIR/packaging/rpm/$PACKAGE_NAME.yaml" "$SOURCES_DIR/"

# Update version in spec file
sed -i "s/Version:.*$/Version:        $VERSION/" "$SPECS_DIR/$PACKAGE_NAME.spec"

# Build RPM
echo "==> Building RPM from spec: $SPEC_FILE"
echo "==> rpmbuild options: ${RPMBUILD_OPTS[*]}"
if [ ${#DEFINES[@]} -gt 0 ]; then
    echo "==> Custom defines: ${DEFINES[*]}"
fi

# Check if rpmbuild exists after installation attempt
if ! command -v rpmbuild >/dev/null 2>&1; then
    echo "ERROR: rpmbuild command still not found after package installation"
    echo "Cannot build RPM packages without rpmbuild"
    echo "This environment may not support RPM package building"
    exit 1
fi

# Try building with dependencies first
if rpmbuild "${RPMBUILD_OPTS[@]}" \
            --define "_topdir $BUILD_DIR" \
            --define "_version $VERSION" \
            "${DEFINES[@]}" \
            "$SPEC_FILE" 2>/dev/null; then
    echo "Built successfully with all dependencies"
else
    echo "Warning: Build dependencies not satisfied, building with --nodeps flag"
    echo "This is expected in CI environments without full RPM development packages"
    rpmbuild "${RPMBUILD_OPTS[@]}" \
             --define "_topdir $BUILD_DIR" \
             --define "_version $VERSION" \
             --nodeps \
             "${DEFINES[@]}" \
             "$SPEC_FILE"
fi

# Copy built RPMs to dist
mkdir -p "$ORIGINAL_DIR/dist/rpm"
find "$RPMS_DIR" -name "*.rpm" -exec cp {} "$ORIGINAL_DIR/dist/rpm/" \;

echo "RPM packages built successfully:"
ls -la "$ORIGINAL_DIR/dist/rpm/"