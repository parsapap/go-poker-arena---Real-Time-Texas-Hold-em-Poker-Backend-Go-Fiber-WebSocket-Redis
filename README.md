# Go Poker Arena

Production-ready Texas Hold'em Poker backend with real-time features, matchmaking, leaderboards, and metrics.

## Features

### Core Poker Engine
- ✓ Complete Texas Hold'em implementation
- ✓ All game phases: Pre-flop, Flop, Turn, River, Showdown
- ✓ Betting actions: Fold, Check, Call, Raise, All-in
- ✓ Pot management with automatic side pots
- ✓ Fast hand evaluator using bitmasks
- ✓ Cryptographically secure deck shuffling

### Real-Time Features
- ✓ WebSocket connections for live gameplay
- ✓ Redis pub/sub for room updates
- ✓ Broadcast player actions to room
- ✓ Live pot and community card updates
- ✓ Hidden opponent cards (revealed at showdown)

### Matchmaking System
- ✓ Auto-matchmaking queue by skill/chips
- ✓ Redis sorted sets for efficient matching
- ✓ Automatic room creation and game start
- ✓ Queue position tracking

### Leaderboard
- ✓ Redis sorted sets for rankings
- ✓ Top players by wins
- ✓ Top players by chips
- ✓ Individual player stats and rank

### Security & Performance
- ✓ JWT authentication
- ✓ Rate limiting (100 req/min per user)
- ✓ WebSocket connection limits (5 per user)
- ✓ Prometheus metrics endpoint
- ✓ Stress tested with 1000+ concurrent connections

## Tech Stack

- **Go 1.21+**
- **Fiber v2** - Web framework
- **WebSocket** - Real-time communication
- **GORM** - ORM for PostgreSQL
- **Redis** - Pub/sub, caching, leaderboards
- **PostgreSQL** - Database
- **Prometheus** - Metrics
- **JWT** - Authentication
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

### Health & Metrics
- `GET /healthz` - Health check
- `GET /metrics` - Prometheus metrics

### Rooms
- `GET /api/rooms` - List all rooms
- `POST /api/rooms` - Create a new room
- `POST /api/rooms/:id/start` - Start game in room
- `POST /api/rooms/:id/action` - Process player action

### Leaderboard
- `GET /api/leaderboard/wins?limit=10` - Top players by wins
- `GET /api/leaderboard/chips?limit=10` - Top players by chips
- `GET /api/leaderboard/player/:id` - Player stats

### Matchmaking
- `POST /api/matchmaking/join` - Join matchmaking queue
- `POST /api/matchmaking/leave` - Leave matchmaking queue
- `GET /api/matchmaking/status` - Queue status

### Authentication
- `POST /api/auth/login` - Login and get JWT token

### WebSocket
- `GET /ws?user_id=1&username=player1&room_id=room1` - WebSocket connection

## Prometheus Metrics

Available at `/metrics`:

- `http_requests_total` - Total HTTP requests
- `http_request_duration_seconds` - Request duration
- `websocket_connections_active` - Active WebSocket connections
- `websocket_messages_total` - Total WebSocket messages
- `poker_games_active` - Active games
- `poker_games_total` - Total games started
- `poker_hands_dealt_total` - Total hands dealt
- `poker_rooms_active` - Active rooms
- `matchmaking_queue_size` - Players in queue

## Testing

### Run Unit Tests
```bash
make test
```

### Run with Coverage
```bash
make test-coverage
```

### Load Test (100 connections)
```bash
make load-test
```

### Stress Test (1000 connections)
```bash
make stress-test
```

## Rate Limiting

- **API Endpoints**: 100 requests per minute per user
- **WebSocket**: Maximum 5 concurrent connections per user

## WebSocket Message Format

### Client → Server
```json
{
  "type": "action",
  "room_id": "1",
  "payload": {
    "action": "raise",
    "amount": 100
  }
}
```

### Server → Client
```json
{
  "type": "player_action",
  "room_id": 1,
  "player_id": 1,
  "username": "player1",
  "action": "raise",
  "amount": 100,
  "game": {...}
}
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
│       └── main.go              # Application entry point
├── internal/
│   ├── database/
│   │   └── database.go          # Database connection
│   ├── leaderboard/
│   │   └── leaderboard.go       # Redis leaderboard
│   ├── matchmaking/
│   │   └── queue.go             # Matchmaking queue
│   ├── metrics/
│   │   └── metrics.go           # Prometheus metrics
│   ├── middleware/
│   │   ├── jwt.go               # JWT authentication
│   │   └── ratelimit.go         # Rate limiting
│   ├── models/
│   │   ├── user.go              # User model
│   │   ├── room.go              # Room model
│   │   └── game.go              # Game model
│   ├── poker/
│   │   ├── deck.go              # Deck management
│   │   ├── hand.go              # Hand evaluator
│   │   ├── game.go              # Game logic
│   │   └── hand_test.go         # Unit tests
│   ├── rooms/
│   │   └── manager.go           # Room management
│   └── websocket/
│       ├── hub.go               # WebSocket hub
│       └── client.go            # WebSocket client
├── test/
│   ├── ws_load_test.go          # Load testing
│   └── stress_test.go           # Stress testing
├── docker-compose.yml           # Docker services
├── Dockerfile                   # Application container
├── Makefile                     # Build commands
└── .env                         # Environment variables
```

## Environment Variables

See `.env.example` for all available configuration options.

## Performance

Tested with:
- ✓ 1000+ concurrent WebSocket connections
- ✓ 100 simultaneous games
- ✓ Sub-millisecond hand evaluation
- ✓ Redis pub/sub for distributed scaling

## License

MIT
