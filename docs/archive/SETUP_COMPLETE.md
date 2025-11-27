# ✅ Setup Complete!

## 🎉 Docker Services Running

PostgreSQL and Redis are now running in Docker containers:

- **PostgreSQL**: `localhost:5432`
  - Database: `poker_arena`
  - User: `poker`
  - Password: `poker123`

- **Redis**: `localhost:6379`

## 🚀 Next Steps

### 1. Start Backend (Terminal 1)

```bash
./run-backend.sh
```

Or manually:
```bash
cd backend
go run cmd/server/main.go
```

The backend will start on **http://localhost:8080**

### 2. Start Frontend (Terminal 2)

Open a new terminal and run:

```bash
./run-frontend.sh
```

Or manually:
```bash
cd frontend
npm install  # First time only
npm run dev
```

The frontend will start on **http://localhost:3000**

### 3. Access the Application

Open your browser and go to:
- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8080/healthz
- **Metrics**: http://localhost:8080/metrics

## 📋 Useful Commands

### Docker Management

```bash
# View running containers
sudo docker ps

# View logs
sudo docker-compose -f docker-compose.local.yml logs -f

# Stop containers
sudo docker-compose -f docker-compose.local.yml down

# Restart containers
sudo docker-compose -f docker-compose.local.yml restart

# Clean restart (removes data)
sudo docker-compose -f docker-compose.local.yml down -v
sudo docker-compose -f docker-compose.local.yml up -d
```

### Database Access

```bash
# Connect to PostgreSQL
sudo docker exec -it poker-postgres psql -U poker -d poker_arena

# Check Redis
sudo docker exec -it poker-redis redis-cli
```

### Backend Development

```bash
cd backend

# Run tests
go test ./... -v

# Build binary
go build -o bin/poker-server ./cmd/server

# Run binary
./bin/poker-server
```

### Frontend Development

```bash
cd frontend

# Type check
npm run type-check

# Lint
npm run lint

# Build for production
npm run build
npm start
```

## 🔍 Verify Everything is Working

1. **PostgreSQL**: 
   ```bash
   sudo docker exec poker-postgres pg_isready -U poker
   ```
   Should output: `/var/run/postgresql:5432 - accepting connections`

2. **Redis**: 
   ```bash
   sudo docker exec poker-redis redis-cli ping
   ```
   Should output: `PONG`

3. **Backend**: 
   ```bash
   curl http://localhost:8080/healthz
   ```
   Should return: `{"status":"ok","service":"go-poker-arena","version":"1.0.0"}`

4. **Frontend**: 
   Open http://localhost:3000 in your browser

## 🛠️ Troubleshooting

### Backend won't start
- Check if PostgreSQL and Redis are running: `sudo docker ps`
- Verify `.env` file exists in `backend/` directory
- Check port 8080 is not in use: `sudo lsof -i :8080`

### Frontend won't start
- Delete node_modules and reinstall: `cd frontend && rm -rf node_modules && npm install`
- Clear Next.js cache: `rm -rf frontend/.next`
- Check port 3000 is not in use: `sudo lsof -i :3000`

### Database connection errors
- Restart Docker containers: `sudo docker-compose -f docker-compose.local.yml restart`
- Check logs: `sudo docker-compose -f docker-compose.local.yml logs postgres`

### Redis connection errors
- Check Redis logs: `sudo docker-compose -f docker-compose.local.yml logs redis`
- Test connection: `sudo docker exec poker-redis redis-cli ping`

## 📚 Project Structure

```
go-poker-arena/
├── backend/                    # Go backend
│   ├── cmd/server/            # Main entry point
│   ├── internal/              # Internal packages
│   ├── .env                   # Environment config
│   └── go.mod                 # Go dependencies
├── frontend/                  # Next.js frontend
│   ├── app/                   # App router pages
│   ├── components/            # React components
│   ├── .env.local             # Frontend config
│   └── package.json           # npm dependencies
├── docker-compose.local.yml   # Docker services
├── run-backend.sh             # Backend startup script
└── run-frontend.sh            # Frontend startup script
```

## 🎮 Start Playing!

1. Register a new account at http://localhost:3000/register
2. Login at http://localhost:3000/login
3. Create or join a poker room
4. Start playing Texas Hold'em!

## 📞 Need Help?

- Check the logs in your terminal
- Review the [API Documentation](API.md)
- See [TESTING_GUIDE.md](TESTING_GUIDE.md) for testing instructions

---

**Happy Coding! 🎰♠️♥️♣️♦️**
