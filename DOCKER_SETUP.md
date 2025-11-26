# 🐳 Docker Setup Complete!

## ✅ Services Running

All services are now running successfully on Docker:

| Service | Status | Port | URL |
|---------|--------|------|-----|
| **Backend** | ✅ Running | 8080 | http://localhost:8080 |
| **Frontend** | ✅ Running | 3000 | http://localhost:3000 |
| **PostgreSQL** | ✅ Healthy | 5432 | localhost:5432 |
| **Redis** | ✅ Healthy | 6379 | localhost:6379 |

## 🚀 Quick Commands

```bash
# View all running containers
docker compose ps

# View logs
docker compose logs -f

# View specific service logs
docker compose logs backend -f
docker compose logs frontend -f

# Stop all services
docker compose down

# Start all services
docker compose up -d

# Restart a service
docker compose restart backend
docker compose restart frontend

# Rebuild and restart
docker compose up -d --build
```

## 🔗 Access Points

### Frontend
- **URL**: http://localhost:3000
- **Login Page**: http://localhost:3000/login
- **Register Page**: http://localhost:3000/register

### Backend API
- **Health Check**: http://localhost:8080/healthz
- **Metrics**: http://localhost:8080/metrics
- **API Base**: http://localhost:8080/api
- **WebSocket**: ws://localhost:8080/ws

### Database
- **PostgreSQL**: localhost:5432
  - Database: `poker_arena`
  - User: `poker`
  - Password: `poker123`

- **Redis**: localhost:6379

## 📝 Testing the Application

### 1. Register a New User
```bash
curl -X POST http://localhost:8080/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "username": "player1",
    "email": "player1@example.com",
    "password": "password123"
  }'
```

### 2. Login
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "player1",
    "password": "password123"
  }'
```

### 3. Create a Room (with token)
```bash
curl -X POST http://localhost:8080/api/rooms \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -d '{
    "name": "High Stakes",
    "max_players": 6,
    "small_blind": 10,
    "big_blind": 20
  }'
```

### 4. List Rooms
```bash
curl http://localhost:8080/api/rooms \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

## 🎮 Using the Frontend

1. Open http://localhost:3000 in your browser
2. You'll be redirected to the login page
3. Click "Register" to create a new account
4. After registration, you'll be logged in automatically
5. You'll see the lobby with available rooms
6. Create a new room or join an existing one

## 🔧 Configuration Changes Made

### Backend Dockerfile
- Removed `.env.example` copy (not needed in production)
- Using environment variables from docker-compose

### Frontend Dockerfile
- Removed public folder copy (created empty folder)
- Using standalone output for production

### Frontend next.config.js
- Added `output: 'standalone'` for Docker
- Made API URL configurable via environment variable

### docker-compose.yml
- Removed volume mounts for frontend (using production build)
- Removed `npm run dev` command override
- Using production Next.js server

## 🐛 Troubleshooting

### Frontend not loading?
```bash
docker compose logs frontend
docker compose restart frontend
```

### Backend connection issues?
```bash
docker compose logs backend
# Check if PostgreSQL and Redis are healthy
docker compose ps
```

### Database connection failed?
```bash
# Check PostgreSQL logs
docker compose logs postgres

# Restart PostgreSQL
docker compose restart postgres
```

### Clear everything and start fresh
```bash
# Stop and remove all containers, networks, and volumes
docker compose down -v

# Rebuild and start
docker compose up -d --build
```

## 📊 Monitoring

### View Prometheus Metrics
```bash
curl http://localhost:8080/metrics
```

### Check Container Stats
```bash
docker stats
```

### Check Container Health
```bash
docker compose ps
```

## 🎉 Success!

Your Go Poker Arena is now running on Docker with:
- ✅ Production-ready backend (Go + Fiber)
- ✅ Production-ready frontend (Next.js 14)
- ✅ PostgreSQL database with migrations
- ✅ Redis for caching and pub/sub
- ✅ All services networked together
- ✅ Health checks enabled
- ✅ Auto-restart on failure

**Ready to play poker!** 🃏

---

**Note**: For production deployment, remember to:
1. Change JWT_SECRET to a strong random value
2. Update database passwords
3. Configure ALLOWED_ORIGINS for your domain
4. Enable HTTPS
5. Set up proper monitoring and logging
