#!/bin/bash
# autogen.sh - Generate autotools build system files

set -e

# Check for required tools
check_tool() {
    if ! command -v "$1" >/dev/null 2>&1; then
        echo "Error: $1 is required but not found"
        echo "Please install autotools: apt-get install autotools-dev autoconf automake"
        exit 1
    fi
}

echo "Checking for required autotools..."
check_tool autoconf
check_tool automake
check_tool aclocal

# Create m4 directory if it doesn't exist
mkdir -p m4

echo "Running aclocal..."
aclocal -I m4

echo "Running autoconf..."
autoconf

echo "Running automake..."
automake --add-missing --copy --foreign

echo ""
echo "Autotools setup complete!"
echo ""
echo "Now you can run:"
echo "  ./configure [options]"
echo "  make"
echo "  make install"
echo ""
echo "For cross-compilation:"
echo "  ./configure GOOS=linux GOARCH=arm64"
echo ""
echo "For debug build:"
echo "  ./configure --enable-debug"
echo ""
echo "For systemd integration:"
echo "  ./configure --with-systemd"
echo ""