# 🚀 Deployment Checklist

## Pre-Deployment

### Code Quality
- [x] All tests passing
- [x] Code coverage > 80%
- [x] No linting errors
- [x] Documentation complete
- [x] CHANGELOG updated

### Security
- [x] JWT secret configured
- [x] Database credentials secured
- [x] CORS origins configured
- [x] Rate limiting enabled
- [x] HTTPS ready

### Configuration
- [x] Environment variables set
- [x] Database migrations ready
- [x] Redis connection configured
- [x] Logging configured
- [x] Metrics endpoint enabled

## Railway Deployment

### Setup
```bash
# 1. Install Railway CLI
npm install -g @railway/cli

# 2. Login
railway login

# 3. Initialize project
railway init

# 4. Add PostgreSQL
railway add --plugin postgresql

# 5. Add Redis
railway add --plugin redis

# 6. Set environment variables
railway variables set JWT_SECRET=your-secret-key
railway variables set ALLOWED_ORIGINS=https://yourdomain.com

# 7. Deploy
railway up
```

### Environment Variables
```
PORT=8080
ENV=production
LOG_LEVEL=info
JWT_SECRET=<generate-secure-secret>
ALLOWED_ORIGINS=https://yourdomain.com
POSTGRES_HOST=<railway-postgres-host>
POSTGRES_PORT=5432
POSTGRES_USER=<railway-postgres-user>
POSTGRES_PASSWORD=<railway-postgres-password>
POSTGRES_DB=<railway-postgres-db>
REDIS_HOST=<railway-redis-host>
REDIS_PORT=6379
```

## Post-Deployment

### Verification
- [ ] Health check: `curl https://your-app.railway.app/healthz`
- [ ] Metrics: `curl https://your-app.railway.app/metrics`
- [ ] WebSocket: Test connection
- [ ] API endpoints: Test authentication
- [ ] Database: Verify migrations

### Monitoring
- [ ] Set up Prometheus alerts
- [ ] Configure log aggregation
- [ ] Monitor error rates
- [ ] Track response times
- [ ] Watch memory usage

### Performance
- [ ] Run load tests
- [ ] Check latency (p99 < 100ms)
- [ ] Verify connection limits
- [ ] Test auto-scaling
- [ ] Monitor Redis memory

## GitHub Actions

### Secrets to Configure
```
DOCKER_USERNAME=<your-docker-username>
DOCKER_PASSWORD=<your-docker-password>
RAILWAY_TOKEN=<your-railway-token>
```

### Workflows
- [x] CI/CD pipeline configured
- [x] Automated testing
- [x] Docker image build
- [x] Coverage reporting
- [x] Linting

## Domain & SSL

### Custom Domain
```bash
# Add custom domain in Railway dashboard
railway domain add yourdomain.com

# Update DNS records
# A record: @ -> Railway IP
# CNAME: www -> your-app.railway.app
```

### SSL Certificate
- [x] Automatic SSL via Railway
- [x] Force HTTPS redirect
- [x] HSTS headers configured

## Scaling

### Horizontal Scaling
- [ ] Enable auto-scaling in Railway
- [ ] Configure min/max instances
- [ ] Set up load balancer
- [ ] Test failover

### Database
- [ ] Connection pooling configured
- [ ] Read replicas (if needed)
- [ ] Backup strategy
- [ ] Migration rollback plan

### Redis
- [ ] Persistence enabled
- [ ] Memory limits set
- [ ] Eviction policy configured
- [ ] Backup strategy

## Monitoring & Alerts

### Metrics to Track
- HTTP request rate
- WebSocket connections
- Active games
- Queue size
- Error rate
- Response time (p50, p95, p99)
- Memory usage
- CPU usage

### Alerts
- [ ] Error rate > 1%
- [ ] Response time > 500ms
- [ ] Memory usage > 80%
- [ ] Database connection errors
- [ ] Redis connection errors

## Backup & Recovery

### Database Backups
- [ ] Daily automated backups
- [ ] Point-in-time recovery
- [ ] Backup retention (30 days)
- [ ] Test restore procedure

### Disaster Recovery
- [ ] Documented recovery steps
- [ ] RTO: < 1 hour
- [ ] RPO: < 5 minutes
- [ ] Tested recovery plan

## Security Hardening

### Production Checklist
- [x] JWT secret rotated
- [x] Database credentials secured
- [x] API rate limiting enabled
- [x] CORS properly configured
- [x] Security headers set
- [x] Input validation
- [x] SQL injection prevention
- [x] XSS protection

### Monitoring
- [ ] Failed login attempts
- [ ] Suspicious activity
- [ ] Rate limit violations
- [ ] Ban events
- [ ] Admin actions

## Documentation

### User-Facing
- [x] API documentation
- [x] WebSocket protocol
- [x] Authentication guide
- [x] Rate limits documented

### Internal
- [x] Architecture diagram
- [x] Deployment guide
- [x] Runbook
- [x] Troubleshooting guide

## Launch

### Pre-Launch
- [ ] Final security audit
- [ ] Load testing complete
- [ ] Monitoring configured
- [ ] Backup verified
- [ ] Team trained

### Launch Day
- [ ] Deploy to production
- [ ] Verify all services
- [ ] Monitor metrics
- [ ] Be ready for hotfixes
- [ ] Announce launch 🎉

### Post-Launch
- [ ] Monitor for 24 hours
- [ ] Gather user feedback
- [ ] Fix critical bugs
- [ ] Optimize performance
- [ ] Plan next iteration

---

## Quick Commands

```bash
# Deploy
railway up

# View logs
railway logs

# Check status
railway status

# Run migrations
railway run go run cmd/migrate/main.go

# Scale up
railway scale --replicas 3

# Rollback
railway rollback
```

## Support

- 📧 Email: support@poker-arena.com
- 💬 Discord: [Join server](https://discord.gg/poker-arena)
- 🐛 Issues: [GitHub](https://github.com/parsapap/go-poker-arena/issues)

---

**Last Updated:** 2024-01-15
**Version:** 1.0.0
