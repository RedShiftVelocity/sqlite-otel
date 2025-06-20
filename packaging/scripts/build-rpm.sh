#!/bin/bash
set -e

# Script to build RPM package for sqlite-otel-collector

VERSION=${VERSION:-0.7.0}
PACKAGE_NAME="sqlite-otel-collector"
BUILD_DIR="$(pwd)/build/rpm"
SOURCES_DIR="$BUILD_DIR/SOURCES"
SPECS_DIR="$BUILD_DIR/SPECS"
RPMS_DIR="$BUILD_DIR/RPMS"

echo "Building RPM package for $PACKAGE_NAME version $VERSION"

# Clean and create build directories
rm -rf "$BUILD_DIR"
mkdir -p "$SOURCES_DIR" "$SPECS_DIR" "$RPMS_DIR"

# Create source tarball
echo "Creating source tarball..."
TEMP_DIR=$(mktemp -d)
mkdir -p "$TEMP_DIR/$PACKAGE_NAME-$VERSION"

# Copy source files
cp -r . "$TEMP_DIR/$PACKAGE_NAME-$VERSION/"
cd "$TEMP_DIR/$PACKAGE_NAME-$VERSION"

# Clean unnecessary files
rm -rf .git .gitignore build/ dist/ *.db *.log

# Create tarball
cd "$TEMP_DIR"
tar czf "$SOURCES_DIR/$PACKAGE_NAME-$VERSION.tar.gz" "$PACKAGE_NAME-$VERSION"
rm -rf "$TEMP_DIR"

# Copy spec file
cp packaging/rpm/$PACKAGE_NAME.spec "$SPECS_DIR/"

# Update version in spec file
sed -i "s/Version:.*$/Version:        $VERSION/" "$SPECS_DIR/$PACKAGE_NAME.spec"

# Build RPM
echo "Building RPM..."
rpmbuild --define "_topdir $BUILD_DIR" \
         --define "_version $VERSION" \
         -ba "$SPECS_DIR/$PACKAGE_NAME.spec"

# Copy built RPMs to dist
mkdir -p dist/rpm
find "$RPMS_DIR" -name "*.rpm" -exec cp {} dist/rpm/ \;

echo "RPM packages built successfully:"
ls -la dist/rpm/