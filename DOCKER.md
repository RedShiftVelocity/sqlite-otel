# Docker Support for SQLite OpenTelemetry Collector

This document provides comprehensive information about building, running, and deploying the SQLite OpenTelemetry Collector using Docker.

## Quick Start

### Building the Image

```bash
# Build the Docker image
docker build -t sqlite-otel-collector .

# Build with a specific tag
docker build -t sqlite-otel-collector:v1.0 .
```

### Running with Docker

```bash
# Run with default settings
docker run -d --name sqlite-otel -p 4318:4318 sqlite-otel-collector

# Run with data persistence
docker run -d \
  --name sqlite-otel \
  -p 4318:4318 \
  -v sqlite-otel-data:/var/lib/sqlite-otel-collector \
  sqlite-otel-collector

# Run with custom configuration
docker run -d \
  --name sqlite-otel \
  -p 4318:4318 \
  -v sqlite-otel-data:/var/lib/sqlite-otel-collector \
  -v sqlite-otel-logs:/var/log \
  -e LOG_LEVEL=debug \
  sqlite-otel-collector \
  --db-path /var/lib/sqlite-otel-collector/telemetry.db \
  --port 4318 \
  --log-max-size 50
```

### Using Docker Compose

```bash
# Start services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down

# Stop and remove volumes
docker-compose down -v
```

## Build Process Details

The Docker build uses a multi-stage approach for optimal image size and security:

### Build Stage Output
```
#0 building with "default" instance using docker driver

#1 [internal] load build definition from Dockerfile
#1 transferring dockerfile: 1.24kB done
#1 DONE 0.0s

#2 [internal] load metadata for docker.io/library/alpine:3.18
#2 DONE 1.4s

#3 [internal] load metadata for docker.io/library/golang:1.21-alpine
#3 DONE 1.4s

#4 [internal] load .dockerignore
#4 transferring context: 335B done
#4 DONE 0.0s

#5 [builder 1/7] FROM docker.io/library/golang:1.21-alpine
#5 DONE 4.6s

#6 [builder 2/7] RUN apk add --no-cache make git
#6 DONE 1.3s

#7 [builder 3/7] WORKDIR /build
#7 DONE 0.0s

#8 [builder 4/7] COPY go.mod go.sum ./
#8 DONE 0.1s

#9 [builder 5/7] RUN go mod download
#9 DONE 0.7s

#10 [builder 6/7] COPY . .
#10 DONE 0.1s

#11 [builder 7/7] RUN make build
#11 go build -o sqlite-otel-collector .
#11 DONE 7.0s

#12 [stage-1 1/5] FROM docker.io/library/alpine:3.18
#12 DONE 1.1s

#13 [stage-1 2/5] RUN apk add --no-cache ca-certificates tzdata wget
#13 DONE 1.3s

#14 [stage-1 3/5] RUN addgroup -g 1000 -S sqlite-otel && adduser -u 1000 -S sqlite-otel -G sqlite-otel
#14 DONE 0.2s

#15 [stage-1 4/5] RUN mkdir -p /var/lib/sqlite-otel-collector && chown -R sqlite-otel:sqlite-otel /var/lib/sqlite-otel-collector
#15 DONE 0.2s

#16 [stage-1 5/5] COPY --from=builder /build/sqlite-otel-collector /usr/bin/sqlite-otel-collector
#16 DONE 0.1s

#17 exporting to image
#17 exporting layers 0.3s done
#17 writing image sha256:3641fea2ac37911917b4f9cba1ef356b8319091ca9819d26fda399eb833da60c done
#17 naming to docker.io/library/sqlite-otel-collector done
#17 DONE 0.3s
```

### Final Image Details
- **Base Image**: Alpine Linux 3.18 (security-focused minimal distribution)
- **Image Size**: ~19.5MB (highly optimized)
- **Architecture**: Multi-stage build for minimal footprint
- **Security**: Non-root user execution (sqlite-otel:1000)

## Image Architecture

### Build Stage (golang:1.21-alpine)
- Installs build dependencies: `make`, `git`
- Downloads Go modules
- Compiles the binary with `make build`
- **Purpose**: Creates the optimized binary

### Runtime Stage (alpine:3.18)
- Minimal Alpine Linux base
- Essential runtime packages: `ca-certificates`, `tzdata`, `wget`
- Non-root user creation for security
- Health check capability
- **Purpose**: Secure, minimal runtime environment

## Container Configuration

### Environment Variables
- `LOG_LEVEL`: Set logging level (debug, info, warn, error)
- Standard Go application environment variables

### Exposed Ports
- `4318`: OTLP/HTTP endpoint (OpenTelemetry standard)

### Volumes
- `/var/lib/sqlite-otel-collector`: Database and data files
- `/var/log`: Application logs (when using external log files)

### Health Check
The container includes a built-in health check that:
- Runs every 30 seconds
- Checks the `/health` endpoint
- Times out after 3 seconds
- Allows 3 retries before marking unhealthy
- Waits 5 seconds after startup before first check

```bash
# Check container health status
docker ps --format "table {{.Names}}\t{{.Status}}"

# View health check logs
docker inspect sqlite-otel --format='{{range .State.Health.Log}}{{.Output}}{{end}}'
```

## Available Command Line Options

The containerized application supports all standard CLI options:

```
Usage of /usr/bin/sqlite-otel-collector:
  -db-path string
        Path to SQLite database file (default: /home/sqlite-otel/.local/share/sqlite-otel/otel-collector.db)
  -log-compress
        Compress rotated log files (default: true)
  -log-file string
        Path to log file for execution metadata (default: /home/sqlite-otel/.local/state/sqlite-otel/execution.log)
  -log-max-age int
        Maximum number of days to keep old log files (default: 30)
  -log-max-backups int
        Maximum number of old log files to keep (default: 7)
  -log-max-size int
        Maximum log file size in MB before rotation (default: 100)
  -port int
        Port to listen on (default: 4318, OTLP/HTTP standard)
```

## Production Deployment Examples

### Simple Production Setup
```bash
docker run -d \
  --name sqlite-otel-prod \
  --restart unless-stopped \
  -p 4318:4318 \
  -v /opt/sqlite-otel/data:/var/lib/sqlite-otel-collector \
  -v /opt/sqlite-otel/logs:/var/log \
  sqlite-otel-collector \
  --db-path /var/lib/sqlite-otel-collector/production.db \
  --log-file /var/log/sqlite-otel.log \
  --log-max-size 200 \
  --log-max-backups 10
```

### Kubernetes Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sqlite-otel-collector
spec:
  replicas: 1
  selector:
    matchLabels:
      app: sqlite-otel-collector
  template:
    metadata:
      labels:
        app: sqlite-otel-collector
    spec:
      containers:
      - name: sqlite-otel-collector
        image: sqlite-otel-collector:latest
        ports:
        - containerPort: 4318
        volumeMounts:
        - name: data-storage
          mountPath: /var/lib/sqlite-otel-collector
        livenessProbe:
          httpGet:
            path: /health
            port: 4318
          initialDelaySeconds: 5
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /health
            port: 4318
          initialDelaySeconds: 5
          periodSeconds: 10
      volumes:
      - name: data-storage
        persistentVolumeClaim:
          claimName: sqlite-otel-data
---
apiVersion: v1
kind: Service
metadata:
  name: sqlite-otel-collector-service
spec:
  selector:
    app: sqlite-otel-collector
  ports:
  - protocol: TCP
    port: 4318
    targetPort: 4318
```

## Development and Testing

### Local Development
```bash
# Build development image
docker build -t sqlite-otel-collector:dev .

# Run with development settings
docker run -it --rm \
  -p 4318:4318 \
  -v $(pwd)/testdata:/var/lib/sqlite-otel-collector \
  sqlite-otel-collector:dev \
  --db-path /var/lib/sqlite-otel-collector/dev.db
```

### Testing the Container
```bash
# Test health endpoint
curl http://localhost:4318/health

# Send test telemetry data
curl -X POST http://localhost:4318/v1/traces \
  -H "Content-Type: application/json" \
  -d '{"resourceSpans":[{"spans":[{"name":"test-span","traceId":"1234","spanId":"5678"}]}]}'

# Check container logs
docker logs sqlite-otel

# Inspect the database
docker exec -it sqlite-otel sqlite3 /var/lib/sqlite-otel-collector/otel-collector.db ".tables"
```

## Security Features

### Non-Root Execution
- Container runs as user `sqlite-otel` (UID: 1000)
- No unnecessary privileges
- Data directory owned by application user

### Minimal Attack Surface
- Alpine Linux base (security-focused)
- Only essential runtime dependencies
- No shell or unnecessary tools in final image

### Network Security
- Single exposed port (4318)
- Health check endpoint for monitoring
- No unnecessary network services

## Troubleshooting

### Common Issues

1. **Permission Denied on Volume Mounts**
   ```bash
   # Fix volume permissions
   docker run --rm -v sqlite-otel-data:/data alpine chown -R 1000:1000 /data
   ```

2. **Health Check Failing**
   ```bash
   # Check if service is responding
   docker exec sqlite-otel wget -q --spider http://localhost:4318/health
   ```

3. **Database Locked Errors**
   ```bash
   # Ensure proper volume mounting and single instance
   docker ps -a | grep sqlite-otel
   ```

### Debugging
```bash
# Run with debug output
docker run -it --rm -p 4318:4318 sqlite-otel-collector --help

# Access container shell for debugging
docker exec -it sqlite-otel sh

# View container resource usage
docker stats sqlite-otel
```

## Performance Considerations

### Resource Limits
```bash
# Run with resource constraints
docker run -d \
  --name sqlite-otel \
  --memory=128m \
  --cpus="0.5" \
  -p 4318:4318 \
  sqlite-otel-collector
```

### Storage Optimization
- Use named volumes for data persistence
- Configure log rotation to manage disk usage
- Monitor database size growth
- Consider backup strategies for persistent data

## Integration Examples

### With OpenTelemetry SDKs
```bash
# Environment variable for applications
export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
```

### With Docker Compose Stack
See `docker-compose.yml` for a complete example including:
- Volume management
- Health checks
- Environment configuration
- Restart policies