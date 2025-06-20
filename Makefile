# sqlite-otel Makefile

# Binary name
BINARY_NAME=sqlite-otel-collector

# Default target
all: build

# Build the binary
build:
	go build -o $(BINARY_NAME) .

# Run the binary
run: build
	./$(BINARY_NAME)

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)

# Run tests
test:
	go test ./...

# Build for all platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME)-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build -o $(BINARY_NAME)-darwin-amd64 .
	GOOS=windows GOARCH=amd64 go build -o $(BINARY_NAME)-windows-amd64.exe .

# Package targets
package-rpm:
	./packaging/scripts/build-rpm.sh

.PHONY: build run clean test build-all package-rpm