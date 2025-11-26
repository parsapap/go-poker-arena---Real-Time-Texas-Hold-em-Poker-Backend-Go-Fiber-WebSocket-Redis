# Go Poker Arena 🃏

Production-ready Texas Hold'em Poker backend with real-time features, matchmaking, leaderboards, comprehensive security, and anti-cheat measures.

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Security](https://img.shields.io/badge/Security-Hardened-green.svg)](SECURITY.md)

## 🚀 Features

### Core Poker Engine
- ✅ Complete Texas Hold'em implementation
- ✅ All game phases: Pre-flop, Flop, Turn, River, Showdown
- ✅ Betting actions: Fold, Check, Call, Raise, All-in
- ✅ Pot management with automatic side pots
- ✅ Fast hand evaluator using bitmasks
- ✅ Cryptographically secure deck shuffling

### Security & Anti-Cheat
- ✅ **JWT Authentication** with bcrypt password hashing
- ✅ **Latency validation** (max 5s, detects network manipulation)
- ✅ **Action rate limiting** (prevents bot behavior)
- ✅ **Bet validation** (prevents invalid actions)
- ✅ **Card hiding** (opponents' cards hidden until showdown)
- ✅ **Collusion detection** (pattern analysis)
- ✅ **Game history logging** (all actions saved to PostgreSQL)

### Real-Time Features
- ✅ WebSocket connections for live gameplay
- ✅ Redis pub/sub for room updates
- ✅ Broadcast player actions to room
- ✅ Live pot and community card updates
- ✅ Partial game state (security-focused)

### Matchmaking System
- ✅ Auto-matchmaking queue by skill/chips
- ✅ Redis sorted sets for efficient matching
- ✅ Automatic room creation and game start
- ✅ Queue position tracking

### Leaderboard
- ✅ Redis sorted sets for rankings
- ✅ Top players by wins
- ✅ Top players by chips
- ✅ Individual player stats and rank

### Admin Features
- ✅ Admin-only endpoints
- ✅ User ban/unban system
- ✅ Ban records with reasons
- ✅ Room monitoring

### Performance & Monitoring
- ✅ Prometheus metrics endpoint
- ✅ Rate limiting (100 req/min per user)
- ✅ WebSocket connection limits (5 per user)
- ✅ Stress tested with 1000+ concurrent connections

## 🛠 Tech Stack

- **Go 1.21+** - Backend language
- **Fiber v2** - Web framework
- **WebSocket** - Real-time communication
- **GORM** - ORM for PostgreSQL
- **Redis** - Pub/sub, caching, leaderboards
- **PostgreSQL** - Database
- **Prometheus** - Metrics
- **JWT** - Authentication
- **bcrypt** - Password hashing
- **Docker** - Containerization

## 📦 Quick Start

### Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose

### Installation

1. Clone the repository
```bash
git clone https://github.com/parsapap/go-poker-arena.git
cd go-poker-arena
```

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

## 🔐 Security Features

### Authentication
- Signup with email verification
- Login with bcrypt password verification
- JWT tokens with 24-hour expiration
- Automatic IP tracking

### Anti-Cheat
- **Latency Checks**: Max 5000ms, detects manipulation
- **Rate Limiting**: Max 60 actions/min, min 100ms interval
- **Bet Validation**: Validates amounts and player state
- **Card Hiding**: Partial game state sent to clients
- **Action Logging**: All actions saved with timestamps

### Admin Controls
- Ban/unban users
- View all rooms
- Access to game history
- Audit trail

See [SECURITY.md](SECURITY.md) for complete security documentation.

## 📚 API Documentation

See [API.md](API.md) for complete API reference.

### Quick Examples

#### Signup
```bash
curl -X POST http://localhost:8080/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "username": "player1",
    "email": "player1@example.com",
    "password": "securepass123"
  }'
```

#### Login
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "player1",
    "password": "securepass123"
  }'
```

#### Create Room
```bash
curl -X POST http://localhost:8080/api/rooms \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "High Stakes",
    "max_players": 9,
    "small_blind": 10,
    "big_blind": 20
  }'
```

## 🎮 Game Flow

1. **Signup/Login** → Get JWT token
2. **Join Matchmaking** → Auto-matched with players
3. **Room Created** → Game starts automatically
4. **Blinds Posted** → Small/big blinds deducted
5. **Cards Dealt** → 2 hole cards per player
6. **Betting Rounds** → Pre-flop, Flop, Turn, River
7. **Showdown** → Best hand wins
8. **Chips Distributed** → Winner gets pot
9. **History Saved** → Game logged to database
10. **Leaderboard Updated** → Rankings refreshed

## 📊 Prometheus Metrics

Available at `/metrics`:

- `http_requests_total` - Total HTTP requests
- `http_request_duration_seconds` - Request duration
- `websocket_connections_active` - Active WebSocket connections
- `poker_games_active` - Active games
- `poker_hands_dealt_total` - Total hands dealt
- `matchmaking_queue_size` - Players in queue

## 🧪 Testing

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

## 🐳 Docker Deployment

### Multi-Stage Build
```bash
docker build -t poker-arena .
docker run -p 8080:8080 --env-file .env poker-arena
```

### Docker Compose
```bash
docker-compose up -d
```

## 📁 Project Structure

```
go-poker-arena/
├── cmd/server/              # Application entry point
├── internal/
│   ├── anticheat/          # Anti-cheat validation
│   ├── auth/               # Authentication service
│   ├── database/           # Database connection
│   ├── history/            # Game history service
│   ├── leaderboard/        # Redis leaderboard
│   ├── matchmaking/        # Matchmaking queue
│   ├── metrics/            # Prometheus metrics
│   ├── middleware/         # JWT, rate limiting, admin
│   ├── models/             # Data models
│   ├── poker/              # Game engine
│   ├── rooms/              # Room management
│   └── websocket/          # WebSocket hub
├── test/                   # Load & stress tests
├── docker-compose.yml      # Infrastructure
├── Dockerfile              # Multi-stage build
├── API.md                  # API documentation
├── SECURITY.md             # Security documentation
└── FEATURES.md             # Feature list
```

## 🔧 Configuration

### Environment Variables

```env
PORT=8080
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=poker
POSTGRES_PASSWORD=poker123
POSTGRES_DB=poker_arena
REDIS_HOST=localhost
REDIS_PORT=6379
JWT_SECRET=your-secret-key-change-in-production
ALLOWED_ORIGINS=https://yourdomain.com
```

## 🚦 Rate Limits

- **API Endpoints**: 100 requests per minute per user
- **WebSocket**: Maximum 5 concurrent connections per user
- **Actions**: Max 60 per minute, min 100ms interval

## 📈 Performance

Tested with:
- ✅ 1000+ concurrent WebSocket connections
- ✅ 100 simultaneous games
- ✅ Sub-millisecond hand evaluation
- ✅ Redis pub/sub for distributed scaling

## 🤝 Contributing

Contributions are welcome! Please read our contributing guidelines first.

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

## 🔒 Security

For security vulnerabilities, please email: security@poker-arena.com

Do not create public GitHub issues for security vulnerabilities.

## 📞 Support

- Documentation: [API.md](API.md), [SECURITY.md](SECURITY.md)
- Issues: [GitHub Issues](https://github.com/parsapap/go-poker-arena/issues)
- Discussions: [GitHub Discussions](https://github.com/parsapap/go-poker-arena/discussions)

## 🎯 Roadmap

- [ ] Tournament mode
- [ ] Sit & Go tables
- [ ] Multi-table support
- [ ] Player avatars
- [ ] Chat system
- [ ] Replay system
- [ ] Mobile app support
- [ ] Cryptocurrency integration

---

**Built with ❤️ using Go and Fiber**
