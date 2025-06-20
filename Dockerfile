# Multi-stage build for SQLite OTEL Collector

# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache make git

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN make build

# Runtime stage
FROM alpine:3.18

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 -S sqlite-otel && \
    adduser -u 1000 -S sqlite-otel -G sqlite-otel

# Create data directory
RUN mkdir -p /var/lib/sqlite-otel-collector && \
    chown -R sqlite-otel:sqlite-otel /var/lib/sqlite-otel-collector

# Copy binary from builder
COPY --from=builder /build/sqlite-otel-collector /usr/local/bin/sqlite-otel-collector

# Switch to non-root user
USER sqlite-otel

# Expose OTLP/HTTP port
EXPOSE 4318

# Set default database path for container
ENV DB_PATH=/var/lib/sqlite-otel-collector/otel-collector.db

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:4318/health || exit 1

# Run the collector
ENTRYPOINT ["/usr/local/bin/sqlite-otel-collector"]
CMD ["--db-path", "/var/lib/sqlite-otel-collector/otel-collector.db"]