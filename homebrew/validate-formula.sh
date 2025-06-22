#!/bin/bash

# Simple validation script for Homebrew formula
echo "Validating SQLite OpenTelemetry Collector Homebrew formula..."

FORMULA_FILE="homebrew/sqlite-otel-collector.rb"

# Check if formula file exists
if [ ! -f "$FORMULA_FILE" ]; then
    echo "❌ Formula file not found: $FORMULA_FILE"
    exit 1
fi

echo "✅ Formula file exists"

# Check for required components
required_fields=(
    "class SqliteOtelCollector"
    "desc"
    "homepage"
    "version"
    "license"
    "url"
    "sha256"
    "def install"
    "test do"
    "service do"
)

for field in "${required_fields[@]}"; do
    if grep -q "$field" "$FORMULA_FILE"; then
        echo "✅ Found: $field"
    else
        echo "❌ Missing: $field"
    fi
done

# Check SHA256 format
if grep -E "sha256 \"[a-f0-9]{64}\"" "$FORMULA_FILE" > /dev/null; then
    echo "✅ SHA256 checksums appear valid"
else
    echo "❌ SHA256 checksums may be invalid"
fi

# Check URL format
if grep -E "https://github\.com/.*/releases/download/v[0-9]+\.[0-9]+\.[0-9]+/" "$FORMULA_FILE" > /dev/null; then
    echo "✅ GitHub release URLs appear valid"
else
    echo "❌ GitHub release URLs may be invalid"
fi

echo ""
echo "Formula validation complete!"
echo ""
echo "To test with Homebrew (if installed):"
echo "  brew audit --strict $FORMULA_FILE"
echo "  brew install --build-from-source $FORMULA_FILE"
echo "  brew test sqlite-otel-collector"