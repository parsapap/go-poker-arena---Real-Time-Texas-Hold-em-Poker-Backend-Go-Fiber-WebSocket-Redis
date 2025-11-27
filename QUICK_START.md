# 🚀 Quick Start Guide - Go Poker Arena (Fixed)

## Prerequisites
- Docker & Docker Compose
- Go 1.21+ (for local development)
- Node.js 20+ (for local development)

---

## 🐳 Running with Docker (Recommended)

### 1. Start Everything
```bash
docker-compose up --build
```

### 2. Access the Application
- **Frontend:** http://localhost:3001
- **Backend API:** http://localhost:8080
- **Health Check:** http://localhost:8080/healthz
- **Metrics:** http://localhost:8080/metrics

### 3. Create an Account
1. Go to http://localhost:3001
2. Click "Register"
3. Create username and password
4. You'll get 1000 chips to start

### 4. Play Poker!
- Create a room or join an existing one
- Wait for other players (or open multiple browser tabs)
- Play Texas Hold'em!

---

## 💻 Local Development

### Backend
```bash
cd backend

# Install dependencies
go mod download

# Set up environment
cp .env.example .env

# Start PostgreSQL and Redis (Docker)
docker-compose up postgres redis -d

# Run server
go run cmd/server/main.go
```

### Frontend
```bash
cd frontend

# Install dependencies
npm install

# Set up environment
cp .env.local.example .env.local

# Run development server
npm run dev
```

Access at http://localhost:3000 (or 3001 if configured)

---

## 🔧 Configuration

### Environment Variables

#### Backend (.env)
```bash
PORT=8080
ENV=development
LOG_LEVEL=info

# Database
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=poker
POSTGRES_PASSWORD=poker123
POSTGRES_DB=poker_arena
AUTO_MIGRATE=true

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# Security
JWT_SECRET=your-secret-key-change-in-production
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001
```

#### Frontend (.env.local)
```bash
API_URL=http://localhost:8080
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_WS_URL=ws://localhost:8080
NEXT_PUBLIC_FRONTEND_URL=http://localhost:3001
```

---

## 🧪 Testing

### Run Backend Tests
```bash
cd backend
go test ./... -v
```

### Run Poker Engine Tests (100% Coverage)
```bash
cd backend
go test ./internal/poker -v -cover
```

### Check TypeScript Types
```bash
cd frontend
npm run type-check
```

---

## 🐛 Troubleshooting

### Issue: "Cannot connect to WebSocket"
**Solution:** Make sure backend is running on port 8080
```bash
curl http://localhost:8080/healthz
```

### Issue: "CORS error"
**Solution:** Check ALLOWED_ORIGINS includes your frontend URL
```bash
# In backend/.env
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001
```

### Issue: "Database connection failed"
**Solution:** Ensure PostgreSQL is running
```bash
docker-compose up postgres -d
docker-compose logs postgres
```

### Issue: "Redis connection failed"
**Solution:** Ensure Redis is running
```bash
docker-compose up redis -d
docker-compose logs redis
```

### Issue: "Frontend can't reach API"
**Solution:** Check Next.js proxy configuration
```bash
# Should see rewrites in frontend/next.config.js
cat frontend/next.config.js
```

---

## 📊 Monitoring

### View Logs
```bash
# All services
docker-compose logs -f

# Backend only
docker-compose logs -f backend

# Frontend only
docker-compose logs -f frontend
```

### Check Metrics
```bash
curl http://localhost:8080/metrics
```

### Database Status
```bash
docker-compose exec postgres psql -U poker -d poker_arena -c "\dt"
```

### Redis Status
```bash
docker-compose exec redis redis-cli ping
```

---

## 🎮 Game Features

### Create a Room
1. Click "Create Room" button
2. Set blinds (e.g., 10/20)
3. Set max players (2-9)
4. Click "Create"

### Join a Room
1. Browse available rooms
2. Click "Join" on any room
3. Wait for other players
4. Game starts automatically with 2+ players

### Game Actions
- **Fold:** Give up your hand
- **Check:** Pass (when no bet)
- **Call:** Match current bet
- **Raise:** Increase the bet
- **All-in:** Bet all your chips

### Phases
1. **Pre-flop:** 2 hole cards dealt
2. **Flop:** 3 community cards
3. **Turn:** 4th community card
4. **River:** 5th community card
5. **Showdown:** Best hand wins!

---

## 🔐 Security Features

- ✅ JWT authentication (24h expiration)
- ✅ Bcrypt password hashing
- ✅ Rate limiting (100 req/min)
- ✅ Anti-cheat validation
- ✅ Card hiding (opponents can't see your cards)
- ✅ Admin ban system
- ✅ Audit trail in database

---

## 📱 Multi-Player Testing

### Option 1: Multiple Browser Tabs
1. Open http://localhost:3001 in multiple tabs
2. Register different users in each tab
3. Join the same room
4. Play!

### Option 2: Multiple Browsers
1. Chrome: User 1
2. Firefox: User 2
3. Safari: User 3
4. All join same room

### Option 3: Incognito Windows
1. Open multiple incognito windows
2. Each acts as a separate user
3. Register and play

---

## 🚀 Production Deployment

### Pre-Deployment Checklist
- [ ] Set `AUTO_MIGRATE=false`
- [ ] Change `JWT_SECRET` to strong random value
- [ ] Update `ALLOWED_ORIGINS` to your domain
- [ ] Use `wss://` for WebSocket (HTTPS)
- [ ] Set `NODE_ENV=production`
- [ ] Configure proper database credentials
- [ ] Set up SSL/TLS certificates
- [ ] Configure monitoring/alerting
- [ ] Set up backup strategy
- [ ] Test all features in staging

### Deploy to Railway
```bash
# Backend
cd backend
railway up

# Frontend
cd frontend
vercel deploy
```

### Deploy to Docker Swarm
```bash
docker stack deploy -c docker-compose.yml poker
```

### Deploy to Kubernetes
See `DEPLOYMENT.md` for Kubernetes manifests

---

## 📚 Additional Resources

- **API Documentation:** [API.md](API.md)
- **Security Guide:** [SECURITY.md](SECURITY.md)
- **Deployment Guide:** [DEPLOYMENT.md](DEPLOYMENT.md)
- **Fixes Applied:** [FIXES_APPLIED.md](FIXES_APPLIED.md)
- **Contributing:** [CONTRIBUTING.md](CONTRIBUTING.md)

---

## 💡 Tips

1. **Development:** Use `LOG_LEVEL=debug` for verbose logging
2. **Testing:** Open browser DevTools to see WebSocket messages
3. **Performance:** Monitor `/metrics` endpoint for bottlenecks
4. **Debugging:** Check `docker-compose logs` for errors
5. **Database:** Use `psql` to inspect game state

---

## 🎉 You're Ready!

Everything is configured and ready to go. Just run:

```bash
docker-compose up
```

Then visit http://localhost:3001 and start playing poker! 🃏

---

## 🆘 Need Help?

- 📧 Email: support@poker-arena.com
- 💬 Discord: [Join server](https://discord.gg/poker-arena)
- 🐛 Issues: [GitHub Issues](https://github.com/parsapap/go-poker-arena/issues)
- 📖 Docs: Check the markdown files in this repo

**Happy Playing! 🎰**
