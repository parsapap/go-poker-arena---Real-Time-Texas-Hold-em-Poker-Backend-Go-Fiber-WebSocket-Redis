# Go Poker Arena - Backend

Production-ready Texas Hold'em Poker backend built with Go, Fiber, WebSocket, PostgreSQL, and Redis.

## Tech Stack

- **Go 1.21+**
- **Fiber v2** - Web framework
- **WebSocket** - Real-time communication
- **GORM** - ORM for PostgreSQL
- **Redis** - Pub/sub, caching, leaderboards
- **PostgreSQL** - Database
- **Prometheus** - Metrics
- **JWT** - Authentication
- **bcrypt** - Password hashing

## Quick Start

```bash
# Install dependencies
go mod download

# Run server
go run cmd/server/main.go

# Run tests
go test ./... -v

# Build
go build -o bin/poker-server ./cmd/server
```

## API Endpoints

See [../API.md](../API.md) for complete API documentation.

## Environment Variables

```env
PORT=8080
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
REDIS_HOST=localhost
REDIS_PORT=6379
JWT_SECRET=your-secret-key
```

## Project Structure

```
backend/
├── cmd/server/          # Application entry point
├── internal/
│   ├── anticheat/      # Anti-cheat validation
│   ├── auth/           # Authentication
│   ├── database/       # Database connection
│   ├── history/        # Game history
│   ├── leaderboard/    # Leaderboards
│   ├── logger/         # Structured logging
│   ├── matchmaking/    # Matchmaking queue
│   ├── metrics/        # Prometheus metrics
│   ├── middleware/     # JWT, rate limiting
│   ├── models/         # Data models
│   ├── poker/          # Game engine
│   ├── rooms/          # Room management
│   └── websocket/      # WebSocket hub
└── test/               # Load & stress tests
```

## Development

```bash
# Run with hot reload
air

# Run tests with coverage
go test ./... -cover

# Run linter
golangci-lint run
```

## Docker

```bash
# Build
docker build -t poker-backend .

# Run
docker run -p 8080:8080 --env-file .env poker-backend
```
