# Deployment Guide

## Production Deployment Checklist

### 1. Environment Configuration

Update `.env` with production values:

```env
PORT=8080
POSTGRES_HOST=your-db-host
POSTGRES_PORT=5432
POSTGRES_USER=poker_prod
POSTGRES_PASSWORD=<strong-password>
POSTGRES_DB=poker_arena_prod
REDIS_HOST=your-redis-host
REDIS_PORT=6379
REDIS_PASSWORD=<redis-password>
JWT_SECRET=<generate-strong-secret>
ALLOWED_ORIGINS=https://yourdomain.com,https://www.yourdomain.com
```

### 2. Generate Strong JWT Secret

```bash
openssl rand -base64 64
```

### 3. Database Setup

```bash
# Create database
psql -U postgres -c "CREATE DATABASE poker_arena_prod;"

# Run migrations (automatic on startup)
# Migrations include:
# - users table
# - rooms table
# - games table
# - game_histories table
# - player_actions table
# - ban_records table
```

### 4. Redis Setup

```bash
# Install Redis
sudo apt-get install redis-server

# Configure Redis password
sudo nano /etc/redis/redis.conf
# Add: requirepass <your-password>

# Restart Redis
sudo systemctl restart redis
```

### 5. Docker Deployment

#### Build Image
```bash
docker build -t poker-arena:latest .
```

#### Run Container
```bash
docker run -d \
  --name poker-arena \
  -p 8080:8080 \
  --env-file .env \
  --restart unless-stopped \
  poker-arena:latest
```

#### Docker Compose (Recommended)
```bash
docker-compose -f docker-compose.prod.yml up -d
```

### 6. HTTPS Setup with Nginx

```nginx
server {
    listen 80;
    server_name yourdomain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name yourdomain.com;

    ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;

    # WebSocket support
    location /ws {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 86400;
    }

    # API endpoints
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### 7. SSL Certificate (Let's Encrypt)

```bash
sudo apt-get install certbot python3-certbot-nginx
sudo certbot --nginx -d yourdomain.com -d www.yourdomain.com
```

### 8. Monitoring Setup

#### Prometheus Configuration
```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'poker-arena'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
```

#### Grafana Dashboard
Import dashboard for:
- Active connections
- Request rate
- Error rate
- Game metrics
- Queue size

### 9. Logging

```bash
# View logs
docker logs -f poker-arena

# Or with Docker Compose
docker-compose logs -f app
```

### 10. Backup Strategy

#### Database Backup
```bash
# Daily backup script
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
pg_dump -U poker_prod poker_arena_prod > backup_$DATE.sql
gzip backup_$DATE.sql

# Upload to S3 or backup storage
aws s3 cp backup_$DATE.sql.gz s3://your-backup-bucket/
```

#### Redis Backup
```bash
# Redis automatically saves to dump.rdb
# Copy to backup location
cp /var/lib/redis/dump.rdb /backup/redis_$(date +%Y%m%d).rdb
```

### 11. Health Checks

```bash
# Application health
curl https://yourdomain.com/healthz

# Database connection
psql -U poker_prod -h your-db-host -d poker_arena_prod -c "SELECT 1;"

# Redis connection
redis-cli -h your-redis-host -a <password> ping
```

### 12. Performance Tuning

#### PostgreSQL
```sql
-- Increase connection pool
ALTER SYSTEM SET max_connections = 200;

-- Optimize for performance
ALTER SYSTEM SET shared_buffers = '256MB';
ALTER SYSTEM SET effective_cache_size = '1GB';
ALTER SYSTEM SET maintenance_work_mem = '64MB';
```

#### Redis
```conf
# redis.conf
maxmemory 512mb
maxmemory-policy allkeys-lru
```

### 13. Security Hardening

```bash
# Firewall rules
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw allow 22/tcp
sudo ufw enable

# Fail2ban for SSH
sudo apt-get install fail2ban
sudo systemctl enable fail2ban
```

### 14. Create Admin User

```bash
# Connect to database
psql -U poker_prod -d poker_arena_prod

# Update user to admin
UPDATE users SET is_admin = true WHERE username = 'admin';
```

### 15. Monitoring Alerts

Set up alerts for:
- High error rate (>5%)
- High latency (>1s)
- Database connection failures
- Redis connection failures
- High memory usage (>80%)
- High CPU usage (>80%)

## Kubernetes Deployment (Optional)

### Deployment YAML
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: poker-arena
spec:
  replicas: 3
  selector:
    matchLabels:
      app: poker-arena
  template:
    metadata:
      labels:
        app: poker-arena
    spec:
      containers:
      - name: poker-arena
        image: poker-arena:latest
        ports:
        - containerPort: 8080
        env:
        - name: POSTGRES_HOST
          valueFrom:
            secretKeyRef:
              name: poker-secrets
              key: postgres-host
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: poker-secrets
              key: jwt-secret
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /healthz
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

### Service YAML
```yaml
apiVersion: v1
kind: Service
metadata:
  name: poker-arena-service
spec:
  selector:
    app: poker-arena
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: LoadBalancer
```

## Scaling Considerations

### Horizontal Scaling
- Use Redis pub/sub for cross-instance communication
- Session affinity for WebSocket connections
- Load balancer with sticky sessions

### Vertical Scaling
- Increase container resources
- Optimize database queries
- Add database read replicas

## Rollback Procedure

```bash
# Docker
docker stop poker-arena
docker rm poker-arena
docker run -d --name poker-arena poker-arena:previous-version

# Kubernetes
kubectl rollout undo deployment/poker-arena
```

## Post-Deployment Verification

1. ✅ Health check returns 200
2. ✅ Can signup new user
3. ✅ Can login
4. ✅ Can create room
5. ✅ Can start game
6. ✅ WebSocket connection works
7. ✅ Metrics endpoint accessible
8. ✅ Leaderboard updates
9. ✅ Game history saves
10. ✅ Admin endpoints work

## Support

For deployment issues, contact: devops@poker-arena.com
