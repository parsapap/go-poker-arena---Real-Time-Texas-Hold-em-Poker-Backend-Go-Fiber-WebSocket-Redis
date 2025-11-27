# Go Poker Arena - Setup Guide

## Quick Start

### Prerequisites
- Go 1.21+
- Node.js 18+
- Docker & Docker Compose
- PostgreSQL (via Docker)
- Redis (via Docker)

### 1. Start Database Services
```bash
docker-compose up -d postgres redis
```

### 2. Start Backend
```bash
cd backend
cp .env.example .env
# Edit .env if needed
go run cmd/server/main.go
```

Backend will run on: http://localhost:8080

### 3. Start Frontend
```bash
cd frontend
cp .env.local.example .env.local
# Edit .env.local if needed
npm install
npm run dev
```

Frontend will run on: http://localhost:3000

### 4. Create Account & Play
1. Go to http://localhost:3000/register
2. Create an account (you'll get 1,000 chips)
3. Create a room in the lobby
4. Join your room and play!

## Environment Variables

### Backend (.env)
```bash
PORT=8080
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=poker
POSTGRES_PASSWORD=poker123
POSTGRES_DB=poker_arena
REDIS_HOST=localhost
REDIS_PORT=6379
JWT_SECRET=your-secret-key
ALLOWED_ORIGINS=http://localhost:3000
AUTO_MIGRATE=true
```

### Frontend (.env.local)
```bash
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_WS_URL=ws://localhost:8080
NEXT_PUBLIC_FRONTEND_URL=http://localhost:3000
```

## Testing

### Test Authentication
```bash
./test-auth.sh
```

### Test WebSocket
Open `test-websocket.html` in your browser

## Troubleshooting

### Backend not starting
- Check if PostgreSQL is running: `docker-compose ps`
- Check if Redis is running: `docker-compose ps`
- Check backend logs for errors

### Frontend not connecting
- Verify backend is running: `curl http://localhost:8080/healthz`
- Check browser console for errors
- Verify environment variables are set

### WebSocket disconnecting
- Create a room first before joining
- Check backend logs when connecting
- Use test-websocket.html to debug

## Project Structure

```
.
├── backend/              # Go backend
│   ├── cmd/server/      # Main entry point
│   ├── internal/        # Internal packages
│   └── test/            # Tests
├── frontend/            # Next.js frontend
│   ├── app/            # Pages (App Router)
│   ├── components/     # React components
│   ├── hooks/          # Custom hooks
│   ├── lib/            # Utilities
│   ├── store/          # Zustand store
│   └── types/          # TypeScript types
├── docs/               # Documentation
├── docker-compose.yml  # Docker services
└── test-*.sh          # Test scripts
```

## Features

- Real-time multiplayer poker
- WebSocket-based gameplay
- User authentication (JWT)
- Room management
- Matchmaking system
- Leaderboards
- Game history
- Anti-cheat measures
- Admin controls

## Tech Stack

**Backend:**
- Go 1.21
- Fiber (web framework)
- GORM (ORM)
- PostgreSQL (database)
- Redis (caching/pub-sub)
- WebSocket (real-time)

**Frontend:**
- Next.js 14
- React 18
- TypeScript
- Tailwind CSS
- Framer Motion
- Zustand (state)

## Support

For issues or questions, check:
- `docs/TROUBLESHOOTING.md`
- `docs/API.md`
- Backend logs
- Browser console
