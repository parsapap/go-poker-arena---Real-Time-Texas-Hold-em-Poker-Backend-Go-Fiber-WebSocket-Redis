# Go Poker Arena

Production-ready Texas Hold'em Poker backend built with Go, Fiber, WebSockets, PostgreSQL, and Redis.

## Features

- Real-time WebSocket communication
- Room management (create, join, leave)
- Redis pub/sub for distributed messaging
- PostgreSQL for persistent storage
- Docker support with docker-compose
- Health check endpoint
- Load testing support (100+ concurrent connections)

## Tech Stack

- **Go 1.21+**
- **Fiber v2** - Web framework
- **WebSocket** - Real-time communication
- **GORM** - ORM for PostgreSQL
- **Redis** - Pub/sub and caching
- **PostgreSQL** - Database
- **Docker** - Containerization

## Quick Start

### Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose

### Installation

1. Clone the repository
2. Copy environment variables:
```bash
cp .env.example .env
```

3. Start infrastructure services:
```bash
make docker-up
```

4. Install dependencies:
```bash
go mod download
```

5. Run the server:
```bash
make run
```

The server will start on `http://localhost:8080`

## API Endpoints

### HTTP Endpoints

- `GET /healthz` - Health check
- `GET /rooms` - List all rooms
- `POST /rooms` - Create a new room

### WebSocket Endpoint

- `GET /ws?user_id=1&username=player1&room_id=room1` - WebSocket connection

## WebSocket Message Format

```json
{
  "type": "action_type",
  "room_id": "room1",
  "user_id": 1,
  "username": "player1",
  "payload": {}
}
```

## Load Testing

Run load test with 100 concurrent WebSocket connections:

```bash
make load-test
```

## Docker Commands

```bash
# Start services
make docker-up

# Stop services
make docker-down

# Clean everything
make clean
```

## Project Structure

```
go-poker-arena/
├── cmd/
│   └── server/
│       └── main.go           # Application entry point
├── internal/
│   ├── database/
│   │   └── database.go       # Database connection
│   ├── models/
│   │   ├── user.go          # User model
│   │   ├── room.go          # Room model
│   │   └── game.go          # Game model
│   ├── rooms/
│   │   └── manager.go       # Room management
│   └── websocket/
│       ├── hub.go           # WebSocket hub (broadcast, register, unregister)
│       └── client.go        # WebSocket client
├── test/
│   └── ws_load_test.go      # Load testing
├── docker-compose.yml        # Docker services
├── Dockerfile               # Application container
├── Makefile                 # Build commands
└── .env                     # Environment variables
```

## Environment Variables

See `.env.example` for all available configuration options.

## Development

Build the application:
```bash
make build
```

Run locally:
```bash
make run
```

## License

MIT
