# Go Poker Arena — Backend

Real-time Texas Hold'em poker backend built with Go, Fiber, WebSocket, PostgreSQL, and Redis.

## Tech Stack

- **Go 1.21+**
- **Fiber v2** — HTTP framework
- **gorilla/fasthttp WebSocket** — real-time gameplay
- **GORM + PostgreSQL** — persistence
- **Redis** — pub/sub, caching, leaderboards, matchmaking
- **Prometheus** — metrics
- **zerolog** — structured logging
- **JWT (HS256)** + **bcrypt** — auth

## Run locally

Prerequisites: Go 1.21+, PostgreSQL, and Redis (or use Docker Compose below).

```bash
# 1. Copy and edit environment
cp ../.env.example .env        # then set at least JWT_SECRET

# 2. Install dependencies
go mod download

# 3. Run the server
go run ./cmd/server
# or with the Makefile
make run
```

Or bring up the whole stack (Postgres + Redis + backend + frontend):

```bash
docker compose up --build        # from the repository root
```

## Testing

```bash
make test            # go test ./...  (unit tests, race detector in CI)
make test-coverage   # HTML coverage report
make test-integration  # end-to-end; requires a running server + Postgres + Redis
```

Integration tests live in `test/integration` behind the `integration` build tag,
so they never run during the default `go test ./...`.

## Environment variables

| Variable | Required | Default | Notes |
|---|---|---|---|
| `ENV` | no | `development` | Non-`development` = production (JSON logs, strict CORS, strong-secret enforcement) |
| `PORT` | no | `8080` | |
| `LOG_LEVEL` | no | `info` | `debug`/`info`/`warn`/`error` |
| `JWT_SECRET` | **yes** | — | Must be ≥16 chars and non-placeholder in production |
| `ALLOWED_ORIGINS` | prod | `http://localhost:3000` | Comma-separated; must not be `*` in production |
| `BODY_LIMIT_BYTES` | no | `1048576` | Max request body (1 MiB) |
| `REQUEST_TIMEOUT_SECONDS` | no | `30` | HTTP read/write timeout |
| `API_RATE_LIMIT` | no | `100` | Requests/min per user or IP on `/api` |
| `WS_MAX_CONNECTIONS` | no | `5` | Max concurrent WS connections per user |
| `POSTGRES_HOST` | **yes** | — | |
| `POSTGRES_PORT` | no | `5432` | |
| `POSTGRES_USER` | **yes** | — | |
| `POSTGRES_PASSWORD` | no | — | |
| `POSTGRES_DB` | **yes** | — | |
| `POSTGRES_SSLMODE` | no | `disable` | Use `require`/`verify-full` in production |
| `AUTO_MIGRATE` | no | `false` | Auto-run GORM migrations (dev only) |
| `DB_MAX_OPEN_CONNS` | no | `25` | Connection pool |
| `DB_MAX_IDLE_CONNS` | no | `10` | Connection pool |
| `DB_CONN_MAX_LIFETIME_MINUTES` | no | `30` | Connection pool |
| `REDIS_HOST` | **yes** | — | |
| `REDIS_PORT` | no | `6379` | |
| `REDIS_PASSWORD` | no | — | |
| `REDIS_DB` | no | `0` | |
| `REDIS_POOL_SIZE` | no | `20` | |

Configuration is loaded and validated at startup (`internal/config`). The server
**refuses to start** if a critical value is missing or weak.

## ⚠️ Breaking change: WebSocket now requires JWT

WebSocket connections are authenticated. The old query params `user_id`/`username`
are **ignored** — identity comes from the verified token. Clients must connect as:

```
ws://<host>/ws?token=<JWT>&room_id=<id>
```

(The `Authorization: Bearer <JWT>` header also works for non-browser clients.)

## HTTP endpoints

| Path | Purpose |
|---|---|
| `GET /health`, `GET /healthz` | Liveness probe |
| `GET /readyz` | Readiness probe (checks Postgres + Redis) |
| `GET /metrics` | Prometheus metrics |
| `POST /auth/signup`, `POST /auth/login` | Auth (public) |
| `/api/*` | Authenticated API (JWT + rate limit + ban check) |
| `GET /ws` | Authenticated WebSocket |

See [../API.md](../API.md) for the full API reference.

## Architecture overview

```
              ┌─────────────┐        ┌──────────────┐
  Clients ───▶│  Fiber HTTP │───────▶│  PostgreSQL  │  (users, rooms, history)
   (WS+REST)  │  + WS Hub   │        └──────────────┘
              │             │        ┌──────────────┐
              │             │───────▶│    Redis     │  (pub/sub, cache,
              └─────────────┘        └──────────────┘   leaderboard, queue)
```

- **Game engine** (`internal/poker`) is pure and deterministic (hand eval, betting,
  side pots, showdown).
- **Room manager** (`internal/rooms`) holds live games in memory and broadcasts
  state changes over Redis pub/sub so multiple instances can fan out to clients.
- **WebSocket hub** (`internal/websocket`) subscribes to `room:*` and pushes events
  to connected clients.
- **Matchmaking** (`internal/matchmaking`) uses a Redis sorted set keyed by skill,
  with a members hash for O(log N) leave.

```
backend/
├── cmd/server/         # entry point, route wiring, graceful shutdown
├── internal/
│   ├── anticheat/      # action/latency validation
│   ├── auth/           # signup/login, bcrypt
│   ├── config/         # env loading + validation
│   ├── database/       # connection + pooling + migrations
│   ├── history/        # game history persistence
│   ├── leaderboard/    # Redis sorted-set leaderboards
│   ├── logger/         # zerolog setup
│   ├── matchmaking/    # Redis-backed queue
│   ├── metrics/        # Prometheus collectors + middleware
│   ├── middleware/     # JWT, WS auth, rate limit, admin/ban, request ID, security headers
│   ├── models/         # GORM models
│   ├── poker/          # game engine
│   ├── rooms/          # room/game lifecycle
│   └── websocket/      # hub + client pumps
└── test/
    ├── integration/    # end-to-end (build tag: integration)
    ├── loadtest/       # WebSocket load generator
    └── stress/         # stress generator
```

## Deployment

- Multi-stage `Dockerfile` builds a static binary into a `scratch` image running
  as a non-root user (uid 65532).
- The runtime image has no shell/curl — use **external** HTTP probes against
  `/healthz` (liveness) and `/readyz` (readiness), e.g. Kubernetes probes.
- Production compose: `docker compose -f docker-compose.prod.yml up --build`.
  Required env vars (`JWT_SECRET`, `ALLOWED_ORIGINS`, `POSTGRES_PASSWORD`) are
  enforced by compose and will fail the deploy if unset.

```bash
docker build -t poker-backend ./backend
docker run -p 8080:8080 --env-file .env poker-backend
```

## Observability

Prometheus metrics include HTTP request counts/latency/errors (labeled by route
pattern), WebSocket connections/messages, active games/rooms, hands dealt,
matchmaking queue size, and a structured `errors_total{component,type}` counter.
Pair with Grafana for dashboards and Alertmanager for alerting on error rates.
