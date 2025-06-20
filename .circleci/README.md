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
- Test result storage and coverage reports
- Multi-platform Linux builds (amd64, arm64, arm)
- Automatic release archive creation for tags
- Artifact storage for all builds

## Usage

1. Push changes to any branch to trigger tests and builds
2. Create a version tag (e.g., `v0.7.0`) to trigger release workflow
3. View test results and artifacts in CircleCI dashboard

## Configuration Details

- **Go Version**: 1.21
- **Test Options**: Race detection enabled, coverage reports
- **Build Targets**: Current platform + all Linux architectures
- **Release Format**: tar.gz archives with version in filename