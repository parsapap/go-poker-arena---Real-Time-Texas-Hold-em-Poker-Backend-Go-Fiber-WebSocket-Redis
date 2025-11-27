# 🎰 START HERE - Go Poker Arena

## ✅ Setup Complete!

Your development environment is ready:

| Service | Status | Port | Container |
|---------|--------|------|-----------|
| PostgreSQL | ✅ Running | 5432 | poker-postgres |
| Redis | ✅ Running | 6379 | poker-redis |
| Backend | ⏳ Ready to start | 8080 | - |
| Frontend | ⏳ Ready to start | 3000 | - |

## 🚀 Start Your Application

### Terminal 1 - Backend
```bash
./run-backend.sh
```

### Terminal 2 - Frontend  
```bash
./run-frontend.sh
```

### Browser
```
http://localhost:3000
```

## 📚 Documentation

- **[QUICK_START_LOCAL.md](QUICK_START_LOCAL.md)** - Quick commands
- **[SETUP_COMPLETE.md](SETUP_COMPLETE.md)** - Full setup guide
- **[LOCAL_SETUP.md](LOCAL_SETUP.md)** - Detailed instructions
- **[README.md](README.md)** - Project overview

## 🎮 First Time Setup

1. Start backend and frontend (see above)
2. Go to http://localhost:3000/register
3. Create an account
4. Login and start playing!

## 🔧 Common Commands

```bash
# View Docker containers
sudo docker ps

# Stop Docker services
sudo docker-compose -f docker-compose.local.yml down

# Start Docker services
sudo docker-compose -f docker-compose.local.yml up -d

# View logs
sudo docker-compose -f docker-compose.local.yml logs -f
```

## 💡 Tips

- Keep Docker containers running (they use minimal resources)
- Stop/start backend and frontend as needed
- Backend auto-migrates database on startup
- Frontend has hot-reload enabled

---

**Ready to play poker? Run the commands above! 🎰♠️♥️♣️♦️**
