# Deployment Guide

Production deployment strategies for the SQLite OTEL Collector.

## Overview

The SQLite OTEL Collector can be deployed in various ways depending on your infrastructure and requirements:

- **Standalone Binary** - Direct installation on servers
- **Docker** - Containerized deployment
- **Systemd Service** - Linux service with package management
- **Kubernetes** - Container orchestration
- **Edge Deployment** - IoT and edge computing scenarios

## Standalone Binary Deployment

### System Requirements

**Minimum:**
- 64MB RAM
- 10MB disk space + storage for telemetry data
- Network connectivity on chosen port

**Recommended:**
- 256MB RAM
- 1GB+ disk space for telemetry data
- Dedicated user account

### Installation Steps

1. **Download the binary:**
```bash
# Download for your platform
wget https://github.com/RedShiftVelocity/sqlite-otel/releases/latest/download/sqlite-otel-linux-amd64
chmod +x sqlite-otel-linux-amd64
sudo mv sqlite-otel-linux-amd64 /usr/local/bin/sqlite-otel
```

2. **Create dedicated user:**
```bash
sudo useradd --system --shell /bin/false --home /var/lib/sqlite-otel sqlite-otel
sudo mkdir -p /var/lib/sqlite-otel
sudo chown sqlite-otel:sqlite-otel /var/lib/sqlite-otel
```

3. **Create systemd service:**
```bash
sudo tee /etc/systemd/system/sqlite-otel.service << EOF
[Unit]
Description=SQLite OpenTelemetry Collector
After=network.target

[Service]
Type=simple
User=sqlite-otel
Group=sqlite-otel
ExecStart=/usr/local/bin/sqlite-otel -port 4318 -db-path /var/lib/sqlite-otel/collector.db
Restart=always
RestartSec=5

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/sqlite-otel /var/log

[Install]
WantedBy=multi-user.target
EOF
```

4. **Start the service:**
```bash
sudo systemctl daemon-reload
sudo systemctl enable sqlite-otel
sudo systemctl start sqlite-otel
```

## Docker Deployment

### Basic Docker Deployment

```bash
# Create persistent volume
docker volume create sqlite-otel-data

# Run container
docker run -d \
  --name sqlite-otel \
  --restart unless-stopped \
  -p 4318:4318 \
  -v sqlite-otel-data:/var/lib/sqlite-otel-collector \
  ghcr.io/redshiftvelocity/sqlite-otel:latest
```

### Docker Compose

```yaml
# docker-compose.yml
version: '3.8'

services:
  sqlite-otel:
    image: ghcr.io/redshiftvelocity/sqlite-otel:latest
    container_name: sqlite-otel
    restart: unless-stopped
    ports:
      - "4318:4318"
    volumes:
      - sqlite-otel-data:/var/lib/sqlite-otel-collector
      - ./logs:/var/log/sqlite-otel
    environment:
      - TZ=UTC
    command: [
      "-port", "4318",
      "-log-max-size", "50",
      "-log-max-backups", "5"
    ]
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:4318/health"]
      interval: 30s
      timeout: 10s
      retries: 3

volumes:
  sqlite-otel-data:
```

Start with:
```bash
docker-compose up -d
```

### Production Docker Configuration

```bash
# Production deployment with resource limits
docker run -d \
  --name sqlite-otel-prod \
  --restart unless-stopped \
  --memory=512m \
  --cpus=1.0 \
  -p 4318:4318 \
  -v /opt/telemetry/data:/var/lib/sqlite-otel-collector \
  -v /opt/telemetry/logs:/var/log/sqlite-otel \
  --user 1000:1000 \
  ghcr.io/redshiftvelocity/sqlite-otel:latest \
  -port 4318 \
  -log-max-size 100 \
  -log-max-backups 10 \
  -log-max-age 30
```

## Package Management Deployment

### Debian/Ubuntu (.deb)

```bash
# Download and install
wget https://github.com/RedShiftVelocity/sqlite-otel/releases/latest/download/sqlite-otel-collector_amd64.deb
sudo dpkg -i sqlite-otel-collector_amd64.deb

# Service is automatically installed and configured
sudo systemctl status sqlite-otel-collector
```

### RHEL/CentOS/Fedora (.rpm)

```bash
# Download and install
wget https://github.com/RedShiftVelocity/sqlite-otel/releases/latest/download/sqlite-otel-collector-amd64.rpm
sudo rpm -ivh sqlite-otel-collector-amd64.rpm

# Service is automatically installed and configured
sudo systemctl status sqlite-otel-collector
```

### Service Configuration

Override default settings:
```bash
# Create override directory
sudo mkdir -p /etc/systemd/system/sqlite-otel-collector.service.d

# Create override configuration
sudo tee /etc/systemd/system/sqlite-otel-collector.service.d/override.conf << EOF
[Service]
ExecStart=
ExecStart=/usr/bin/sqlite-otel-collector -port 4318 -log-max-size 50
EOF

# Reload and restart
sudo systemctl daemon-reload
sudo systemctl restart sqlite-otel-collector
```

## Kubernetes Deployment

### Basic Kubernetes Deployment

```yaml
# k8s-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sqlite-otel
  labels:
    app: sqlite-otel
spec:
  replicas: 1
  selector:
    matchLabels:
      app: sqlite-otel
  template:
    metadata:
      labels:
        app: sqlite-otel
    spec:
      containers:
      - name: sqlite-otel
        image: ghcr.io/redshiftvelocity/sqlite-otel:latest
        ports:
        - containerPort: 4318
        args:
        - "-port"
        - "4318"
        - "-log-max-size"
        - "25"
        volumeMounts:
        - name: data-volume
          mountPath: /var/lib/sqlite-otel-collector
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 4318
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 4318
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: data-volume
        persistentVolumeClaim:
          claimName: sqlite-otel-pvc

---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: sqlite-otel-pvc
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi

---
apiVersion: v1
kind: Service
metadata:
  name: sqlite-otel-service
spec:
  selector:
    app: sqlite-otel
  ports:
  - port: 4318
    targetPort: 4318
  type: ClusterIP
```

Deploy:
```bash
kubectl apply -f k8s-deployment.yaml
```

### Production Kubernetes with ConfigMap

```yaml
# k8s-production.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: sqlite-otel-config
data:
  collector.args: |
    -port
    4318
    -log-max-size
    50
    -log-max-backups
    7
    -log-max-age
    30

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sqlite-otel-prod
  labels:
    app: sqlite-otel
    environment: production
spec:
  replicas: 1
  selector:
    matchLabels:
      app: sqlite-otel
  template:
    metadata:
      labels:
        app: sqlite-otel
    spec:
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        fsGroup: 1000
      containers:
      - name: sqlite-otel
        image: ghcr.io/redshiftvelocity/sqlite-otel:latest
        ports:
        - containerPort: 4318
        command: ["/usr/bin/sqlite-otel"]
        args:
        - "-port"
        - "4318"
        - "-log-max-size"
        - "50"
        volumeMounts:
        - name: data-volume
          mountPath: /var/lib/sqlite-otel-collector
        - name: log-volume
          mountPath: /var/log/sqlite-otel
        resources:
          requests:
            memory: "256Mi"
            cpu: "200m"
          limits:
            memory: "1Gi"
            cpu: "1000m"
        livenessProbe:
          httpGet:
            path: /health
            port: 4318
          initialDelaySeconds: 30
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /health
            port: 4318
          initialDelaySeconds: 10
          periodSeconds: 10
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            drop:
            - ALL
          readOnlyRootFilesystem: true
      volumes:
      - name: data-volume
        persistentVolumeClaim:
          claimName: sqlite-otel-data-pvc
      - name: log-volume
        emptyDir: {}
```

## Edge Deployment

### IoT and Edge Computing

For resource-constrained edge environments:

```bash
# Minimal configuration for edge deployment
sqlite-otel \
  -port 4318 \
  -db-path /data/telemetry.db \
  -log-max-size 5 \
  -log-max-backups 1 \
  -log-max-age 3
```

### Docker for Edge

```bash
# Edge deployment with minimal resources
docker run -d \
  --name sqlite-otel-edge \
  --restart unless-stopped \
  --memory=64m \
  --cpus=0.5 \
  -p 4318:4318 \
  -v edge-data:/var/lib/sqlite-otel-collector \
  ghcr.io/redshiftvelocity/sqlite-otel:latest \
  -log-max-size 5 \
  -log-max-backups 1
```

## High Availability

### Load Balancer Configuration

While SQLite OTEL Collector instances can't share a database, you can load balance across multiple instances:

```nginx
# nginx.conf
upstream sqlite_otel_backend {
    server 10.1.1.10:4318;
    server 10.1.1.11:4318;
    server 10.1.1.12:4318;
}

server {
    listen 80;
    location / {
        proxy_pass http://sqlite_otel_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### Multi-Instance Deployment

Deploy multiple collectors for different services or regions:

```bash
# Service A collector
docker run -d --name sqlite-otel-service-a \
  -p 4318:4318 \
  -v service-a-data:/var/lib/sqlite-otel-collector \
  ghcr.io/redshiftvelocity/sqlite-otel:latest

# Service B collector
docker run -d --name sqlite-otel-service-b \
  -p 4319:4318 \
  -v service-b-data:/var/lib/sqlite-otel-collector \
  ghcr.io/redshiftvelocity/sqlite-otel:latest
```

## Monitoring and Observability

### Health Monitoring

```bash
# Health check script
#!/bin/bash
if curl -f http://localhost:4318/health > /dev/null 2>&1; then
    echo "SQLite OTEL Collector is healthy"
    exit 0
else
    echo "SQLite OTEL Collector is unhealthy"
    exit 1
fi
```

### Log Monitoring

```bash
# Monitor logs for errors
tail -f /var/log/sqlite-otel-collector.log | grep -i error

# Or for systemd
journalctl -u sqlite-otel-collector -f | grep -i error
```

### Resource Monitoring

```bash
# Monitor resource usage
ps aux | grep sqlite-otel
du -sh /var/lib/sqlite-otel-collector/
```

## Backup and Recovery

### Database Backup

```bash
# Create backup script
#!/bin/bash
BACKUP_DIR="/backup/sqlite-otel"
DB_PATH="/var/lib/sqlite-otel-collector/otel-collector.db"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p "$BACKUP_DIR"
sqlite3 "$DB_PATH" ".backup $BACKUP_DIR/backup_$DATE.db"
gzip "$BACKUP_DIR/backup_$DATE.db"

# Keep only last 7 days of backups
find "$BACKUP_DIR" -name "backup_*.db.gz" -mtime +7 -delete
```

### Automated Backup with Cron

```bash
# Add to crontab
0 2 * * * /opt/scripts/backup-sqlite-otel.sh
```

## Security Considerations

### Network Security

```bash
# Firewall configuration (UFW)
sudo ufw allow from 10.0.0.0/8 to any port 4318
sudo ufw deny 4318

# Or iptables
iptables -A INPUT -s 10.0.0.0/8 -p tcp --dport 4318 -j ACCEPT
iptables -A INPUT -p tcp --dport 4318 -j DROP
```

### File Permissions

```bash
# Secure file permissions
sudo chmod 600 /var/lib/sqlite-otel-collector/otel-collector.db
sudo chown sqlite-otel:sqlite-otel /var/lib/sqlite-otel-collector/otel-collector.db
```

### TLS Termination

Use a reverse proxy for TLS:

```nginx
server {
    listen 443 ssl;
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;
    
    location / {
        proxy_pass http://localhost:4318;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## Troubleshooting

### Common Issues

**Service won't start:**
```bash
# Check service status
sudo systemctl status sqlite-otel-collector
sudo journalctl -u sqlite-otel-collector

# Check file permissions
ls -la /var/lib/sqlite-otel-collector/
```

**Database corruption:**
```bash
# Check database integrity
sqlite3 /var/lib/sqlite-otel-collector/otel-collector.db "PRAGMA integrity_check;"

# Restore from backup if needed
cp /backup/sqlite-otel/backup_latest.db /var/lib/sqlite-otel-collector/otel-collector.db
```

**High resource usage:**
```bash
# Check database size
du -sh /var/lib/sqlite-otel-collector/

# Optimize database
sqlite3 /var/lib/sqlite-otel-collector/otel-collector.db "VACUUM;"
```

## Performance Optimization

### Database Optimization

```sql
-- Run periodically to optimize
PRAGMA optimize;
VACUUM;
ANALYZE;

-- Check database stats
.dbinfo
```

### Resource Limits

```bash
# Systemd resource limits
sudo systemctl edit sqlite-otel-collector

[Service]
MemoryLimit=1G
CPUQuota=50%
```

## See Also

- [Configuration Guide](configuration.md) - Detailed configuration options
- [CLI Reference](cli.md) - Command-line options
- [Quick Start](quickstart.md) - Getting started guide