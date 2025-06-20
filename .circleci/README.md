# CircleCI Configuration

This directory contains the CircleCI configuration for continuous integration and deployment.

## Overview

The CI/CD pipeline consists of three main jobs:

1. **Test**: Runs all tests with race detection and generates coverage reports
2. **Build**: Builds binaries for all Linux platforms  
3. **Release**: Creates release archives when a version tag is pushed

## Workflow

- **test-and-build**: Runs on all branches and tags
  - Test job runs first
  - Build job runs after tests pass
  - Release job only runs on version tags (v*)

## Features

- Go module caching for faster builds
- Static analysis with `go vet`
- Test result storage and coverage reports
- Multi-platform builds (Linux, macOS, Windows)
- Automatic release archive creation for tags
- Artifact storage for all builds
- Error handling with `set -e` in shell scripts
- Centralized binary name configuration

## Usage

1. Push changes to any branch to trigger tests and builds
2. Create a version tag (e.g., `v0.7.0`) to trigger release workflow
3. View test results and artifacts in CircleCI dashboard

## Configuration Details

- **Go Version**: 1.21
- **Binary Name**: Configured via BINARY_NAME environment variable
- **Test Options**: Race detection enabled, coverage reports, static analysis
- **Build Targets**: Current platform + all supported platforms (Linux, macOS, Windows)
- **Release Format**: tar.gz archives with format: `{binary}-{version}-{platform}.tar.gz`
- **Error Handling**: Shell scripts use `set -e` for fail-fast behavior