# sqlite-otel Makefile

# Binary name
BINARY_NAME=sqlite-otel
SERVICE_NAME=sqlite-otel-collector

# Version information
MAJOR_MINOR?=v0.7
VERSION?=$(shell \
	if git describe --tags --exact-match HEAD >/dev/null 2>&1; then \
		git describe --tags --exact-match HEAD; \
	else \
		count=$$(git rev-list --count HEAD 2>/dev/null || echo "0"); \
		echo "$(MAJOR_MINOR).$$count"; \
	fi \
)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go build flags
LDFLAGS=-ldflags "-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}"
BUILDFLAGS=-trimpath

# Platforms to build for
PLATFORMS=linux/amd64 linux/arm64 linux/arm darwin/amd64 darwin/arm64 windows/amd64

# Default target
all: build

# Build for current platform
build:
	@echo "Building ${BINARY_NAME} ${VERSION} for current platform..."
	go build ${BUILDFLAGS} ${LDFLAGS} -o ${BINARY_NAME} .

# Run the binary
run: build
	./${BINARY_NAME}

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f ${BINARY_NAME}
	rm -f ${BINARY_NAME}.exe
	rm -rf dist/
	rm -rf releases/
	rm -f coverage.out coverage.html

# Run tests
test:
	@echo "Running tests..."
	go test -race -coverprofile=coverage.out ./...

# Run tests with coverage report
test-coverage: test
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Build for all platforms
build-all:
	@echo "Building for all platforms..."
	@mkdir -p dist
	@for platform in $(PLATFORMS); do \
		GOOS=$$(echo $$platform | cut -d/ -f1); \
		GOARCH=$$(echo $$platform | cut -d/ -f2); \
		output_name=${BINARY_NAME}-$$GOOS-$$GOARCH; \
		if [ "$$GOOS" = "windows" ]; then output_name="$$output_name.exe"; fi; \
		echo "Building $$output_name..."; \
		GOOS=$$GOOS GOARCH=$$GOARCH go build ${BUILDFLAGS} ${LDFLAGS} -o dist/$$output_name .; \
	done
	@echo "Build complete. Binaries in dist/"

# Build only Linux binaries (for CI/CD)
build-linux:
	@echo "Building Linux binaries..."
	@mkdir -p dist
	@for arch in amd64 arm64 arm; do \
		echo "Building Linux $$arch binary..."; \
		GOOS=linux GOARCH=$$arch go build ${BUILDFLAGS} ${LDFLAGS} -o dist/${BINARY_NAME}-linux-$$arch .; \
	done

# Install binary to system
install: build
	@echo "Installing ${BINARY_NAME} to /usr/local/bin..."
	@echo "Note: You may need to run 'sudo make install' for system-wide installation"
	@install -m 755 ${BINARY_NAME} /usr/local/bin/ || (echo "Error: Permission denied. Try 'sudo make install'" && exit 1)

# Uninstall binary from system
uninstall:
	@echo "Removing ${BINARY_NAME} from /usr/local/bin..."
	@echo "Note: You may need to run 'sudo make uninstall' for system-wide removal"
	@rm -f /usr/local/bin/${BINARY_NAME} || (echo "Error: Permission denied. Try 'sudo make uninstall'" && exit 1)

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install from https://golangci-lint.run/"; \
		exit 1; \
	fi

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	go mod tidy

# Verify dependencies
verify:
	@echo "Verifying dependencies..."
	go mod verify

# Create release archives
# Note: Uses tar.gz format for all platforms. Windows users may prefer .zip archives.
# Consider using zip for Windows binaries in production releases.
release: build-all
	@echo "Creating release archives..."
	@mkdir -p releases
	@cd dist && for file in *; do \
		if [ -f "$$file" ]; then \
			tar czf ../releases/$$file-${VERSION}.tar.gz $$file; \
			echo "Created releases/$$file-${VERSION}.tar.gz"; \
		fi; \
	done

# Development build (with race detector)
dev:
	@echo "Building development version with race detector..."
	go build -race ${LDFLAGS} -o ${BINARY_NAME} .

# Run with example flags
run-example: build
	./${BINARY_NAME} -port 0 -db-path ./test.db

# Show version
version:
	@echo "Version: ${VERSION}"
	@echo "Git Commit: ${GIT_COMMIT}"
	@echo "Build Time: ${BUILD_TIME}"

# Help
help:
	@echo "Available targets:"
	@echo "  make build          - Build for current platform"
	@echo "  make build-all      - Build for all platforms"
	@echo "  make build-linux    - Build Linux binaries only"
	@echo "  make test           - Run tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make install        - Install binary to /usr/local/bin"
	@echo "  make uninstall      - Remove binary from /usr/local/bin"
	@echo "  make fmt            - Format code"
	@echo "  make lint           - Run linter"
	@echo "  make tidy           - Tidy dependencies"
	@echo "  make release        - Create release archives"
	@echo "  make dev            - Build with race detector"
	@echo "  make run-example    - Run with example flags"
	@echo "  make version        - Show version information"
	@echo "  make help           - Show this help"
	@echo "  make package-deb    - Build DEB package"
	@echo "  make package-rpm    - Build RPM package"

# Package targets
package-deb:
	./packaging/scripts/build-deb.sh

package-rpm:
	./packaging/scripts/build-rpm.sh

.PHONY: all build run clean test test-coverage build-all build-linux install uninstall fmt lint tidy verify release dev run-example version help package-deb package-rpm
