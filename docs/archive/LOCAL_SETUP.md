# Local Development Setup Guide

This guide will help you run Redis and PostgreSQL in Docker, while running the backend and frontend locally.

## Prerequisites

- Docker and Docker Compose installed
- Go 1.21+ installed
- Node.js 18+ and npm installed

## Step 1: Install Docker (if not already installed)

### Ubuntu/Debian
```bash
# Update package index
sudo apt-get update

# Install Docker
sudo apt-get install -y docker.io docker-compose

# Start Docker service
sudo systemctl start docker
sudo systemctl enable docker

# Add your user to docker group (to run without sudo)
sudo usermod -aG docker $USER

# Log out and log back in for group changes to take effect
```

### Check Docker installation
```bash
docker --version
docker-compose --version
```

## Step 2: Start Redis and PostgreSQL with Docker

```bash
# Start only Redis and PostgreSQL
docker-compose -f docker-compose.local.yml up -d

# Check if containers are running
docker ps

# Check logs if needed
docker-compose -f docker-compose.local.yml logs -f
```

## Step 3: Setup Backend (Go)

```bash
# Navigate to backend directory
cd backend

# Install Go dependencies
go mod download

# Verify .env file exists with correct settings
cat .env

# Run the backend server
go run cmd/server/main.go
```

The backend will start on `http://localhost:8080`

### Alternative: Build and run binary
```bash
cd backend
go build -o bin/poker-server ./cmd/server
./bin/poker-server
```

## Step 4: Setup Frontend (Next.js)

Open a new terminal:

```bash
# Navigate to frontend directory
cd frontend

# Install dependencies
npm install

# Create .env.local file if it doesn't exist
cat > .env.local << EOF
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_WS_URL=ws://localhost:8080
NEXT_PUBLIC_FRONTEND_URL=http://localhost:3000
EOF

# Run development server
npm run dev
```

The frontend will start on `http://localhost:3000`

## Step 5: Verify Everything is Running

1. **PostgreSQL**: `docker exec -it poker-postgres psql -U poker -d poker_arena -c "SELECT version();"`
2. **Redis**: `docker exec -it poker-redis redis-cli ping`
3. **Backend**: Open `http://localhost:8080/healthz`
4. **Frontend**: Open `http://localhost:3000`

## Useful Commands

### Docker Management
```bash
# Stop containers
docker-compose -f docker-compose.local.yml down

# Stop and remove volumes (clean slate)
docker-compose -f docker-compose.local.yml down -v

# View logs
docker-compose -f docker-compose.local.yml logs -f postgres
docker-compose -f docker-compose.local.yml logs -f redis

# Restart containers
docker-compose -f docker-compose.local.yml restart
```

### Backend Development
```bash
# Run tests
cd backend
go test ./... -v

# Run with hot reload (install air first: go install github.com/cosmtrek/air@latest)
air

# Check for errors
go vet ./...
```

### Frontend Development
```bash
# Type check
cd frontend
npm run type-check

# Lint
npm run lint

# Build for production
npm run build
```

## Troubleshooting

### Port Already in Use
If ports 5432 or 6379 are already in use:
```bash
# Check what's using the port
sudo lsof -i :5432
sudo lsof -i :6379

# Stop the service or change ports in docker-compose.local.yml
```

### Database Connection Issues
```bash
# Check if PostgreSQL is accepting connections
docker exec -it poker-postgres pg_isready -U poker

# Connect to database manually
docker exec -it poker-postgres psql -U poker -d poker_arena
```

### Redis Connection Issues
```bash
# Test Redis connection
docker exec -it poker-redis redis-cli ping

# Check Redis info
docker exec -it poker-redis redis-cli info
```

### Backend Won't Start
- Verify `.env` file exists in `backend/` directory
- Check if ports 8080 is available
- Ensure PostgreSQL and Redis are running

### Frontend Won't Start
- Delete `node_modules` and reinstall: `rm -rf node_modules && npm install`
- Clear Next.js cache: `rm -rf .next`
- Check if port 3000 is available

## Clean Restart

If you need to start fresh:

```bash
# Stop all containers and remove volumes
docker-compose -f docker-compose.local.yml down -v

# Remove backend binary
rm -f backend/bin/poker-server

# Clean frontend build
cd frontend
rm -rf .next node_modules
npm install

# Start Docker containers again
docker-compose -f docker-compose.local.yml up -d

# Start backend
cd backend
go run cmd/server/main.go

# Start frontend (in another terminal)
cd frontend
npm run dev
```

## Development Workflow

1. Start Docker containers once: `docker-compose -f docker-compose.local.yml up -d`
2. Run backend: `cd backend && go run cmd/server/main.go`
3. Run frontend: `cd frontend && npm run dev`
4. Make changes and test
5. Stop services when done (Docker containers can keep running)

## Notes

- Docker containers will persist data in volumes, so your database won't be lost when you stop them
- Backend auto-migrates database on startup (AUTO_MIGRATE=true in .env)
- Frontend has hot-reload enabled in development mode
- Use `docker-compose -f docker-compose.local.yml down -v` to completely reset databases
