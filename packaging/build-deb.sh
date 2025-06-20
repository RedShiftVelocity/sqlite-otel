#!/bin/bash
set -e

# Script to build DEB package for sqlite-otel-collector

VERSION=${VERSION:-0.7.0}
PACKAGE_NAME="sqlite-otel-collector"
BUILD_DIR="$(pwd)/build/deb"

echo "Building DEB package for $PACKAGE_NAME version $VERSION"

# Clean and create build directory
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR"

# Create temporary build directory
TEMP_DIR=$(mktemp -d)
BUILD_ROOT="$TEMP_DIR/$PACKAGE_NAME-$VERSION"
mkdir -p "$BUILD_ROOT"

# Copy source files
echo "Preparing source files..."
cp -r . "$BUILD_ROOT/"
cd "$BUILD_ROOT"

# Clean unnecessary files
rm -rf .git .gitignore build/ dist/ *.db *.log

# Create debian directory
mkdir -p debian
cp -r packaging/deb/* debian/

# Create changelog
cat > debian/changelog << EOF
$PACKAGE_NAME ($VERSION-1) unstable; urgency=medium

  * Cross-platform build support
  * Execution logging with rotation
  * SQLite-only storage
  * Systemd service integration

 -- Claude Code <noreply@anthropic.com>  $(date -R)
EOF

# Create compat file
echo "13" > debian/compat

# Create source format
mkdir -p debian/source
echo "3.0 (native)" > debian/source/format

# Build package
echo "Building DEB package..."
cd "$TEMP_DIR"
dpkg-buildpackage -us -uc -b

# Copy built packages to dist
mkdir -p "$(dirs +0)/dist/deb"
cp "$TEMP_DIR"/*.deb "$(dirs +0)/dist/deb/" 2>/dev/null || true

# Cleanup
rm -rf "$TEMP_DIR"

echo "DEB packages built successfully:"
ls -la dist/deb/