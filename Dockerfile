# Multi-stage Dockerfile for SQLite OpenTelemetry Collector
# Produces a ~19.5MB optimized image with security hardening

# ================================
# Build Stage - Create the binary
# ================================
FROM golang:1.21-alpine AS builder

# Install build dependencies
# - make: Required for Makefile-based build process
# - git: Required for Go module downloads and version info
# - gcc, musl-dev: Required for CGO (go-sqlite3 needs CGO)
RUN apk add --no-cache make git gcc musl-dev

# Set working directory for build
WORKDIR /build

# Copy dependency files first for better Docker layer caching
# This allows dependency downloads to be cached when source code changes
COPY go.mod go.sum ./
RUN go mod download

# Copy all source code
COPY . .

# Build the optimized binary with CGO enabled for go-sqlite3
# Set CGO_ENABLED=1 explicitly for sqlite3 support
ENV CGO_ENABLED=1
RUN make build

# ================================
# Runtime Stage - Minimal image
# ================================
FROM alpine:3.18

# Install minimal runtime dependencies
# - ca-certificates: For HTTPS certificate validation
# - tzdata: For timezone support in logs and timestamps  
# - wget: For health check implementation
RUN apk add --no-cache ca-certificates tzdata wget

# Create non-root user for security
# UID/GID 1000 for compatibility with most systems
RUN addgroup -g 1000 -S sqlite-otel && \
    adduser -u 1000 -S sqlite-otel -G sqlite-otel

# Create application data directory with proper ownership
RUN mkdir -p /var/lib/sqlite-otel-collector && \
    chown -R sqlite-otel:sqlite-otel /var/lib/sqlite-otel-collector

# Copy the compiled binary from build stage
COPY --from=builder /build/sqlite-otel /usr/bin/sqlite-otel-collector

# Switch to non-root user for security
USER sqlite-otel

# Expose the standard OTLP/HTTP port
EXPOSE 4318

# Configure health check for container orchestration
# Checks /health endpoint every 30s with 3s timeout
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:4318/health || exit 1

# Configure container startup
ENTRYPOINT ["/usr/bin/sqlite-otel-collector"]
CMD ["--db-path", "/var/lib/sqlite-otel-collector/otel-collector.db"]