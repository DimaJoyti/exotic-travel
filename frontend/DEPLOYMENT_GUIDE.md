# 🚀 Deployment Guide

This guide provides step-by-step instructions for deploying the Exotic Travel Booking Platform to production environments.

## 📋 Prerequisites

### System Requirements
- **CPU**: 4+ cores (8+ recommended for production)
- **RAM**: 8GB minimum (16GB+ recommended)
- **Storage**: 100GB minimum (SSD recommended)
- **Network**: Stable internet connection with SSL certificate

### Software Requirements
- **Docker** 20.10+
- **Docker Compose** 2.0+
- **Git** 2.30+
- **OpenSSL** (for certificate generation)

### Optional (for Kubernetes deployment)
- **kubectl** 1.25+
- **Kubernetes cluster** 1.25+

## 🐳 Docker Compose Deployment (Recommended)

### 1. Server Preparation

```bash
# Update system packages
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

### 2. Application Setup

```bash
# Clone the repository
git clone https://github.com/your-org/exotic-travel-booking.git
cd exotic-travel-booking

# Create production environment file
cp .env.production.example .env.production

# Edit environment variables (see Configuration section below)
nano .env.production
```

### 3. SSL Certificate Setup

#### Option A: Let's Encrypt (Recommended)
```bash
# Install Certbot
sudo apt install certbot

# Obtain SSL certificate
sudo certbot certonly --standalone \
  -d yourdomain.com \
  -d www.yourdomain.com \
  -d api.yourdomain.com

# Certificates will be saved to:
# /etc/letsencrypt/live/yourdomain.com/fullchain.pem
# /etc/letsencrypt/live/yourdomain.com/privkey.pem

# Set up automatic renewal
echo "0 12 * * * /usr/bin/certbot renew --quiet" | sudo crontab -
```

#### Option B: Self-Signed Certificate (Development/Testing)
```bash
# Generate self-signed certificate
sudo mkdir -p /etc/ssl/certs /etc/ssl/private

sudo openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout /etc/ssl/private/privkey.pem \
  -out /etc/ssl/certs/fullchain.pem \
  -subj "/C=US/ST=State/L=City/O=Organization/CN=yourdomain.com"
```

### 4. Environment Configuration

Edit `.env.production` with your actual values:

```bash
# Critical settings to update:
DB_PASSWORD=your-secure-database-password
REDIS_PASSWORD=your-secure-redis-password
JWT_PRIVATE_KEY=your-jwt-private-key-base64
JWT_PUBLIC_KEY=your-jwt-public-key-base64
ENCRYPTION_KEY=your-32-byte-hex-encryption-key
STRIPE_SECRET_KEY=sk_live_your-stripe-secret-key
STRIPE_WEBHOOK_SECRET=whsec_your-stripe-webhook-secret
NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY=pk_live_your-stripe-publishable-key
SMTP_HOST=smtp.yourdomain.com
SMTP_USER=noreply@yourdomain.com
SMTP_PASSWORD=your-smtp-password
NEXT_PUBLIC_API_URL=https://api.yourdomain.com
NEXT_PUBLIC_APP_URL=https://yourdomain.com
GRAFANA_PASSWORD=your-secure-grafana-password
```

### 5. Generate Security Keys

```bash
# Generate JWT keys
openssl genpkey -algorithm RSA -out jwt-private.key -pkcs8 -pass pass:temp
openssl rsa -pubout -in jwt-private.key -out jwt-public.key

# Convert to base64 for environment variables
JWT_PRIVATE_KEY=$(cat jwt-private.key | base64 -w 0)
JWT_PUBLIC_KEY=$(cat jwt-public.key | base64 -w 0)

# Generate encryption key
ENCRYPTION_KEY=$(openssl rand -hex 32)

# Update .env.production with these values
echo "JWT_PRIVATE_KEY=$JWT_PRIVATE_KEY" >> .env.production
echo "JWT_PUBLIC_KEY=$JWT_PUBLIC_KEY" >> .env.production
echo "ENCRYPTION_KEY=$ENCRYPTION_KEY" >> .env.production

# Clean up key files
rm jwt-private.key jwt-public.key
```

### 6. Deploy Application

```bash
# Create necessary directories
mkdir -p logs/nginx backups uploads

# Deploy using the deployment script
./scripts/deploy.sh production docker-compose

# Or manually deploy
docker-compose -f docker-compose.prod.yml up -d --build

# Check deployment status
./scripts/deploy.sh production docker-compose --status
```

### 7. Verify Deployment

```bash
# Check service health
curl -f https://api.yourdomain.com/health
curl -f https://yourdomain.com/api/health

# Check logs
docker-compose -f docker-compose.prod.yml logs -f

# Monitor services
docker-compose -f docker-compose.prod.yml ps
```

## ☸️ Kubernetes Deployment

### 1. Cluster Preparation

```bash
# Verify cluster connection
kubectl cluster-info

# Create namespace
kubectl apply -f k8s/namespace.yaml

# Apply ConfigMaps
kubectl apply -f k8s/configmap.yaml
```

### 2. Create Secrets

```bash
# Create application secrets
kubectl create secret generic app-secrets \
  --from-literal=DB_PASSWORD=your-db-password \
  --from-literal=REDIS_PASSWORD=your-redis-password \
  --from-literal=JWT_PRIVATE_KEY="$(cat jwt-private.key | base64 -w 0)" \
  --from-literal=JWT_PUBLIC_KEY="$(cat jwt-public.key | base64 -w 0)" \
  --from-literal=ENCRYPTION_KEY="$(openssl rand -hex 32)" \
  --from-literal=STRIPE_SECRET_KEY=sk_live_... \
  --from-literal=STRIPE_WEBHOOK_SECRET=whsec_... \
  --from-literal=NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY=pk_live_... \
  --from-literal=SMTP_HOST=smtp.yourdomain.com \
  --from-literal=SMTP_USER=noreply@yourdomain.com \
  --from-literal=SMTP_PASSWORD=your-smtp-password \
  --namespace=exotic-travel

# Create TLS secret
kubectl create secret tls tls-secret \
  --cert=fullchain.pem \
  --key=privkey.pem \
  --namespace=exotic-travel

# Create image pull secret (if using private registry)
kubectl create secret docker-registry ghcr-secret \
  --docker-server=ghcr.io \
  --docker-username=your-username \
  --docker-password=your-token \
  --namespace=exotic-travel
```

### 3. Deploy Services

```bash
# Deploy database services
kubectl apply -f k8s/postgres.yaml
kubectl apply -f k8s/redis.yaml

# Wait for database services
kubectl wait --for=condition=ready pod -l app=postgres -n exotic-travel --timeout=300s
kubectl wait --for=condition=ready pod -l app=redis -n exotic-travel --timeout=300s

# Deploy application services
kubectl apply -f k8s/backend.yaml
kubectl apply -f k8s/frontend.yaml

# Wait for application services
kubectl wait --for=condition=ready pod -l app=backend -n exotic-travel --timeout=300s
kubectl wait --for=condition=ready pod -l app=frontend -n exotic-travel --timeout=300s

# Deploy ingress
kubectl apply -f k8s/ingress.yaml
```

### 4. Run Database Migrations

```bash
# Run migrations
kubectl exec -n exotic-travel deployment/backend -- go run cmd/migrate/main.go up
```

### 5. Verify Deployment

```bash
# Check pods
kubectl get pods -n exotic-travel

# Check services
kubectl get services -n exotic-travel

# Check ingress
kubectl get ingress -n exotic-travel

# Check logs
kubectl logs -n exotic-travel deployment/backend
kubectl logs -n exotic-travel deployment/frontend
```

## 🔧 Configuration

### Environment Variables

Key environment variables that must be configured:

| Variable | Description | Required |
|----------|-------------|----------|
| `DB_PASSWORD` | PostgreSQL password | Yes |
| `REDIS_PASSWORD` | Redis password | Yes |
| `JWT_PRIVATE_KEY` | JWT signing private key | Yes |
| `JWT_PUBLIC_KEY` | JWT verification public key | Yes |
| `ENCRYPTION_KEY` | Data encryption key | Yes |
| `STRIPE_SECRET_KEY` | Stripe secret key | Yes |
| `STRIPE_WEBHOOK_SECRET` | Stripe webhook secret | Yes |
| `SMTP_PASSWORD` | Email service password | Yes |
| `NEXT_PUBLIC_API_URL` | Backend API URL | Yes |
| `NEXT_PUBLIC_APP_URL` | Frontend app URL | Yes |

### Domain Configuration

Update these domains in your configuration:
- `yourdomain.com` - Main website
- `www.yourdomain.com` - WWW redirect
- `api.yourdomain.com` - API endpoint

### DNS Configuration

Set up DNS records:
```
A     yourdomain.com          -> your-server-ip
A     www.yourdomain.com      -> your-server-ip
A     api.yourdomain.com      -> your-server-ip
CNAME monitoring.yourdomain.com -> yourdomain.com
```

## 📊 Monitoring Setup

### Access Monitoring Dashboards

- **Grafana**: `http://your-server:3001` (admin/your-grafana-password)
- **Prometheus**: `http://your-server:9090`

### Set Up Alerts

1. Configure Slack/email notifications in Grafana
2. Set up alert rules for critical metrics
3. Test alert delivery

## 🔒 Security Checklist

- [ ] SSL certificates installed and configured
- [ ] Strong passwords for all services
- [ ] JWT keys properly generated and secured
- [ ] Firewall configured (ports 80, 443, 22 only)
- [ ] Regular security updates enabled
- [ ] Database access restricted to application only
- [ ] Monitoring and logging enabled
- [ ] Backup procedures tested

## 🔄 Maintenance

### Regular Tasks

```bash
# Update SSL certificates (automated with Let's Encrypt)
sudo certbot renew

# Update Docker images
docker-compose -f docker-compose.prod.yml pull
docker-compose -f docker-compose.prod.yml up -d

# Backup database
docker-compose -f docker-compose.prod.yml exec postgres pg_dump -U postgres exotic_travel_prod > backup_$(date +%Y%m%d).sql

# Clean up old Docker images
docker system prune -f
```

### Scaling

#### Docker Compose Scaling
```bash
# Scale backend services
docker-compose -f docker-compose.prod.yml up -d --scale backend=3

# Scale frontend services
docker-compose -f docker-compose.prod.yml up -d --scale frontend=2
```

#### Kubernetes Scaling
```bash
# Scale backend deployment
kubectl scale deployment backend --replicas=5 -n exotic-travel

# Scale frontend deployment
kubectl scale deployment frontend --replicas=3 -n exotic-travel
```

## 🆘 Troubleshooting

### Common Issues

1. **SSL Certificate Issues**
   ```bash
   # Check certificate validity
   openssl x509 -in /etc/ssl/certs/fullchain.pem -text -noout
   
   # Renew certificate
   sudo certbot renew --force-renewal
   ```

2. **Database Connection Issues**
   ```bash
   # Check database logs
   docker-compose -f docker-compose.prod.yml logs postgres
   
   # Test database connection
   docker-compose -f docker-compose.prod.yml exec backend go run cmd/health/main.go
   ```

3. **Service Health Issues**
   ```bash
   # Check service status
   ./scripts/deploy.sh production docker-compose --status
   
   # View logs
   ./scripts/deploy.sh production docker-compose --logs
   ```

### Rollback Procedure

```bash
# Rollback deployment
./scripts/deploy.sh production docker-compose --rollback

# Or manually rollback
docker-compose -f docker-compose.prod.yml down
# Restore from backup if needed
# Redeploy previous version
```

## 📞 Support

For deployment issues:
1. Check the troubleshooting section above
2. Review application logs
3. Consult the monitoring dashboards
4. Contact the development team with specific error messages and logs

---

This deployment guide ensures a secure, scalable, and maintainable production deployment of the Exotic Travel Booking Platform.
