# Deployment Guide

This guide covers deploying the Exotic Travel Booking Platform to production environments.

## Deployment Options

1. **Docker Compose** (Recommended for small to medium deployments)
2. **Kubernetes** (Recommended for large-scale deployments)
3. **Traditional VPS/Server** (Manual deployment)
4. **Cloud Platforms** (AWS, GCP, Azure)

## Prerequisites

### System Requirements
- **CPU**: 2+ cores (4+ recommended)
- **RAM**: 4GB minimum (8GB+ recommended)
- **Storage**: 50GB minimum (SSD recommended)
- **Network**: Stable internet connection with SSL certificate

### Software Requirements
- **Docker** 20.10+
- **Docker Compose** 2.0+
- **PostgreSQL** 15+ (if not using Docker)
- **Redis** 7+ (if not using Docker)
- **Nginx** (for reverse proxy)
- **SSL Certificate** (Let's Encrypt recommended)

## Docker Compose Deployment

### 1. Server Setup

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Verify installation
docker --version
docker-compose --version
```

### 2. Application Deployment

```bash
# Clone repository
git clone https://github.com/your-org/exotic-travel-booking.git
cd exotic-travel-booking

# Create production environment file
cp .env.example .env.production

# Edit environment variables
nano .env.production
```

### 3. Production Environment Configuration

```bash
# .env.production
ENVIRONMENT=production

# Database Configuration
DB_HOST=postgres
DB_PORT=5432
DB_NAME=exotic_travel_prod
DB_USER=postgres
DB_PASSWORD=your-secure-password

# Redis Configuration
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=your-redis-password

# JWT Configuration
JWT_ISSUER=exotic-travel-booking
JWT_AUDIENCE=exotic-travel-api
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=168h

# Generate secure keys (run these commands)
JWT_PRIVATE_KEY="$(openssl genpkey -algorithm RSA -out - -pkcs8 -pass pass:temp | base64 -w 0)"
JWT_PUBLIC_KEY="$(openssl rsa -pubout -in <(echo "$JWT_PRIVATE_KEY" | base64 -d) -out - | base64 -w 0)"
ENCRYPTION_KEY="$(openssl rand -hex 32)"

# Security Configuration
RATE_LIMIT_RPS=10
RATE_LIMIT_BURST=20
MAX_REQUEST_SIZE=10485760

# TLS Configuration
TLS_ENABLED=true
TLS_CERT_FILE=/etc/ssl/certs/fullchain.pem
TLS_KEY_FILE=/etc/ssl/private/privkey.pem

# Frontend Configuration
NEXT_PUBLIC_API_URL=https://api.yourdomain.com
NEXT_PUBLIC_APP_URL=https://yourdomain.com

# Stripe Configuration
STRIPE_SECRET_KEY=sk_live_...
STRIPE_WEBHOOK_SECRET=whsec_...
NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY=pk_live_...

# Email Configuration
SMTP_HOST=smtp.yourdomain.com
SMTP_PORT=587
SMTP_USER=noreply@yourdomain.com
SMTP_PASSWORD=your-smtp-password
```

### 4. Production Docker Compose

Create `docker-compose.prod.yml`:

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: ${DB_NAME}
      POSTGRES_USER: ${DB_USER}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./backups:/backups
    restart: unless-stopped
    networks:
      - app-network

  redis:
    image: redis:7-alpine
    command: redis-server --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis_data:/data
    restart: unless-stopped
    networks:
      - app-network

  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile.prod
    environment:
      - DB_HOST=postgres
      - REDIS_HOST=redis
    env_file:
      - .env.production
    depends_on:
      - postgres
      - redis
    restart: unless-stopped
    networks:
      - app-network
    volumes:
      - /etc/ssl/certs:/etc/ssl/certs:ro
      - /etc/ssl/private:/etc/ssl/private:ro

  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile.prod
    env_file:
      - .env.production
    depends_on:
      - backend
    restart: unless-stopped
    networks:
      - app-network

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro
      - /etc/ssl/certs:/etc/ssl/certs:ro
      - /etc/ssl/private:/etc/ssl/private:ro
    depends_on:
      - frontend
      - backend
    restart: unless-stopped
    networks:
      - app-network

volumes:
  postgres_data:
  redis_data:

networks:
  app-network:
    driver: bridge
```

### 5. Nginx Configuration

Create `nginx/nginx.conf`:

```nginx
events {
    worker_connections 1024;
}

http {
    upstream backend {
        server backend:8080;
    }

    upstream frontend {
        server frontend:3000;
    }

    # Rate limiting
    limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
    limit_req_zone $binary_remote_addr zone=web:10m rate=30r/s;

    # SSL Configuration
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512:ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;

    # Security Headers
    add_header X-Frame-Options DENY always;
    add_header X-Content-Type-Options nosniff always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains; preload" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;

    # API Server
    server {
        listen 443 ssl http2;
        server_name api.yourdomain.com;

        ssl_certificate /etc/ssl/certs/fullchain.pem;
        ssl_certificate_key /etc/ssl/private/privkey.pem;

        location / {
            limit_req zone=api burst=20 nodelay;
            proxy_pass http://backend;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
    }

    # Frontend Server
    server {
        listen 443 ssl http2;
        server_name yourdomain.com www.yourdomain.com;

        ssl_certificate /etc/ssl/certs/fullchain.pem;
        ssl_certificate_key /etc/ssl/private/privkey.pem;

        location / {
            limit_req zone=web burst=50 nodelay;
            proxy_pass http://frontend;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
    }

    # HTTP to HTTPS redirect
    server {
        listen 80;
        server_name yourdomain.com www.yourdomain.com api.yourdomain.com;
        return 301 https://$server_name$request_uri;
    }
}
```

### 6. SSL Certificate Setup

```bash
# Install Certbot
sudo apt install certbot python3-certbot-nginx

# Obtain SSL certificate
sudo certbot certonly --standalone -d yourdomain.com -d www.yourdomain.com -d api.yourdomain.com

# Set up automatic renewal
sudo crontab -e
# Add: 0 12 * * * /usr/bin/certbot renew --quiet
```

### 7. Deploy Application

```bash
# Build and start services
docker-compose -f docker-compose.prod.yml up -d --build

# Run database migrations
docker-compose -f docker-compose.prod.yml exec backend go run cmd/migrate/main.go up

# Check service status
docker-compose -f docker-compose.prod.yml ps

# View logs
docker-compose -f docker-compose.prod.yml logs -f
```

## Kubernetes Deployment

### 1. Kubernetes Manifests

Create `k8s/` directory with the following files:

#### Namespace
```yaml
# k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: exotic-travel
```

#### ConfigMap
```yaml
# k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
  namespace: exotic-travel
data:
  DB_HOST: "postgres-service"
  REDIS_HOST: "redis-service"
  ENVIRONMENT: "production"
```

#### Secrets
```yaml
# k8s/secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: app-secrets
  namespace: exotic-travel
type: Opaque
data:
  DB_PASSWORD: <base64-encoded-password>
  REDIS_PASSWORD: <base64-encoded-password>
  JWT_PRIVATE_KEY: <base64-encoded-key>
  ENCRYPTION_KEY: <base64-encoded-key>
```

#### PostgreSQL Deployment
```yaml
# k8s/postgres.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: postgres
  namespace: exotic-travel
spec:
  replicas: 1
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
      - name: postgres
        image: postgres:15-alpine
        env:
        - name: POSTGRES_DB
          value: "exotic_travel_prod"
        - name: POSTGRES_USER
          value: "postgres"
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: app-secrets
              key: DB_PASSWORD
        volumeMounts:
        - name: postgres-storage
          mountPath: /var/lib/postgresql/data
      volumes:
      - name: postgres-storage
        persistentVolumeClaim:
          claimName: postgres-pvc
---
apiVersion: v1
kind: Service
metadata:
  name: postgres-service
  namespace: exotic-travel
spec:
  selector:
    app: postgres
  ports:
  - port: 5432
    targetPort: 5432
```

### 2. Deploy to Kubernetes

```bash
# Apply all manifests
kubectl apply -f k8s/

# Check deployment status
kubectl get pods -n exotic-travel

# Check services
kubectl get services -n exotic-travel

# View logs
kubectl logs -f deployment/backend -n exotic-travel
```

## Monitoring and Logging

### 1. Prometheus and Grafana

```yaml
# monitoring/docker-compose.yml
version: '3.8'

services:
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3001:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana_data:/var/lib/grafana

volumes:
  grafana_data:
```

### 2. Log Aggregation

```bash
# Install and configure log aggregation
docker run -d \
  --name loki \
  -p 3100:3100 \
  grafana/loki:latest

# Configure log shipping from application containers
```

## Backup and Recovery

### 1. Database Backup Script

```bash
#!/bin/bash
# backup.sh

BACKUP_DIR="/backups"
DATE=$(date +%Y%m%d_%H%M%S)
DB_NAME="exotic_travel_prod"

# Create backup
docker-compose exec -T postgres pg_dump -U postgres $DB_NAME > $BACKUP_DIR/backup_$DATE.sql

# Compress backup
gzip $BACKUP_DIR/backup_$DATE.sql

# Remove backups older than 30 days
find $BACKUP_DIR -name "backup_*.sql.gz" -mtime +30 -delete

# Upload to cloud storage (optional)
# aws s3 cp $BACKUP_DIR/backup_$DATE.sql.gz s3://your-backup-bucket/
```

### 2. Automated Backup Cron Job

```bash
# Add to crontab
0 2 * * * /path/to/backup.sh
```

## Health Checks and Monitoring

### 1. Health Check Endpoints

The application provides several health check endpoints:

- `/health` - Basic health check
- `/health/detailed` - Detailed health information
- `/metrics` - Prometheus metrics

### 2. Monitoring Alerts

Set up alerts for:
- High error rates
- Database connection issues
- High response times
- Memory/CPU usage
- Disk space

## Security Considerations

### 1. Firewall Configuration

```bash
# UFW firewall rules
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 80/tcp    # HTTP
sudo ufw allow 443/tcp   # HTTPS
sudo ufw enable
```

### 2. Security Updates

```bash
# Automated security updates
sudo apt install unattended-upgrades
sudo dpkg-reconfigure -plow unattended-upgrades
```

### 3. Container Security

- Use non-root users in containers
- Scan images for vulnerabilities
- Keep base images updated
- Use secrets management
- Enable container security scanning

## Troubleshooting

### Common Issues

1. **Database Connection Issues**
   ```bash
   # Check database logs
   docker-compose logs postgres
   
   # Test connection
   docker-compose exec backend go run cmd/health/main.go
   ```

2. **SSL Certificate Issues**
   ```bash
   # Check certificate validity
   openssl x509 -in /etc/ssl/certs/fullchain.pem -text -noout
   
   # Renew certificate
   sudo certbot renew
   ```

3. **Performance Issues**
   ```bash
   # Check resource usage
   docker stats
   
   # Check application metrics
   curl https://api.yourdomain.com/metrics
   ```

### Log Analysis

```bash
# View application logs
docker-compose logs -f backend

# Search for errors
docker-compose logs backend | grep ERROR

# Monitor real-time logs
tail -f /var/log/nginx/access.log
```

This deployment guide provides a comprehensive approach to deploying the Exotic Travel Booking Platform in production environments with proper security, monitoring, and backup procedures.
