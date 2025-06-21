#!/bin/bash
set -e

# Script to build DEB package for sqlite-otel-collector

VERSION=${VERSION:-0.7.0}
PACKAGE_NAME="sqlite-otel-collector"
BUILD_DIR="$(pwd)/build/deb"
ORIGINAL_DIR="$(pwd)"

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

# Verify binary exists before proceeding
if [ ! -f "sqlite-otel" ]; then
    echo "ERROR: Binary sqlite-otel not found in source directory"
    echo "Please run 'make build' before packaging"
    exit 1
fi

# Clean unnecessary files but preserve the binary
rm -rf .git .gitignore build/ dist/ *.db *.log
# Keep the sqlite-otel binary - it's needed for packaging

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

 -- Manish Sinha <manishsinha.tech@gmail.com>  $(date -R)
EOF

# Create compat file
echo "13" > debian/compat

# Create source format
mkdir -p debian/source
echo "3.0 (native)" > debian/source/format

# Build package
echo "Building DEB package..."

# Try building with dependencies first
if dpkg-buildpackage -us -uc -b 2>/dev/null; then
    echo "Built successfully with all dependencies"
else
    echo "Warning: Build dependencies not satisfied, building with -d flag"
    echo "This is expected in CI environments without full Go development packages"
    dpkg-buildpackage -us -uc -b -d
fi

# Copy built packages to dist
mkdir -p "$ORIGINAL_DIR/dist/deb"

# Copy DEB file (dpkg-buildpackage creates it in the parent directory of the build root)
if ls "$TEMP_DIR"/../*.deb 1> /dev/null 2>&1; then
    cp "$TEMP_DIR"/../*.deb "$ORIGINAL_DIR/dist/deb/"
elif ls "$TEMP_DIR"/*.deb 1> /dev/null 2>&1; then
    cp "$TEMP_DIR"/*.deb "$ORIGINAL_DIR/dist/deb/"
else
    echo "ERROR: No DEB file found"
    exit 1
fi

# Cleanup
rm -rf "$TEMP_DIR"

echo "DEB packages built successfully:"
ls -la "$ORIGINAL_DIR/dist/deb/"