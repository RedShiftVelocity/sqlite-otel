#!/bin/bash
# Create release archives for tagged releases

set -e

if [ -z "${CIRCLE_TAG}" ]; then
  echo "Error: CIRCLE_TAG is not set. This job should only run on tagged releases."
  exit 1
fi

mkdir -p releases
cd binaries

for file in sqlite-otel-collector*; do
  if [ -f "$file" ]; then
    # Create archive with simple naming
    archive_name="${file}-${CIRCLE_TAG}.tar.gz"
    tar czf "../releases/${archive_name}" "$file"
    echo "Created releases/${archive_name}"
  fi
done