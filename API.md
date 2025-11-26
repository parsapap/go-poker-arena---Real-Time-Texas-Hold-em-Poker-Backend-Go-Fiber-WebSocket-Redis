# API Documentation

## Base URL
```
http://localhost:8080
```

## Authentication

All protected endpoints require JWT token in Authorization header:
```
Authorization: Bearer <token>
```

---

## Public Endpoints

### Health Check
```http
GET /healthz
```

**Response:**
```json
{
  "status": "ok",
  "service": "go-poker-arena"
}
```

### Metrics
```http
GET /metrics
```
Returns Prometheus metrics in text format.

---

## Authentication Endpoints

### Signup
```http
POST /auth/signup
Content-Type: application/json

{
  "username": "player1",
  "email": "player1@example.com",
  "password": "securepassword123"
}
```

**Response:**
```json
{
  "user": {
    "id": 1,
    "username": "player1",
    "email": "player1@example.com",
    "chips": 1000,
    "wins": 0,
    "losses": 0,
    "is_admin": false
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Login
```http
POST /auth/login
Content-Type: application/json

{
  "username": "player1",
  "password": "securepassword123"
}
```

**Response:**
```json
{
  "user": {
    "id": 1,
    "username": "player1",
    "chips": 1000,
    "wins": 5,
    "losses": 3
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

---

## Room Endpoints

### List Rooms
```http
GET /api/rooms
Authorization: Bearer <token>
```

**Response:**
```json
[
  {
    "id": 1,
    "name": "High Stakes",
    "max_players": 9,
    "small_blind": 10,
    "big_blind": 20,
    "status": "waiting"
  }
]
```

### Create Room
```http
POST /api/rooms
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "My Room",
  "max_players": 6,
  "small_blind": 5,
  "big_blind": 10
}
```

**Response:**
```json
{
  "id": 2,
  "name": "My Room",
  "max_players": 6,
  "small_blind": 5,
  "big_blind": 10,
  "status": "waiting"
}
```

### Start Game
```http
POST /api/rooms/:id/start
Authorization: Bearer <token>
```

**Response:**
```json
{
  "id": 1,
  "room_id": 1,
  "phase": "preflop",
  "players": [...],
  "pots": [{"amount": 30, "players": [1,2,3]}],
  "current_bet": 20,
  "min_raise": 20
}
```

### Process Action
```http
POST /api/rooms/:id/action
Authorization: Bearer <token>
Content-Type: application/json

{
  "player_id": 1,
  "action": "raise",
  "amount": 50,
  "latency": 150
}
```

**Actions:** `fold`, `check`, `call`, `raise`, `allin`

**Response:**
```json
{
  "id": 1,
  "phase": "flop",
  "community_cards": [
    {"suit": "hearts", "rank": "A"},
    {"suit": "diamonds", "rank": "K"},
    {"suit": "clubs", "rank": "Q"}
  ],
  "players": [
    {
      "id": 1,
      "username": "player1",
      "chips": 950,
      "bet": 50,
      "hole_cards": [...]  // Only for this player
    },
    {
      "id": 2,
      "username": "player2",
      "chips": 1000,
      "bet": 0,
      "card_count": 2  // Hidden for opponents
    }
  ],
  "current_bet": 50
}
```

---

## User History Endpoints

### Get User History
```http
GET /api/users/:id/history?limit=20
Authorization: Bearer <token>
```

**Response:**
```json
[
  {
    "id": 1,
    "game_id": 123,
    "room_id": 1,
    "winner_id": 1,
    "pot": 500,
    "duration": 180,
    "created_at": "2024-01-15T10:30:00Z"
  }
]
```

### Get User Stats
```http
GET /api/users/:id/stats
Authorization: Bearer <token>
```

**Response:**
```json
{
  "user_id": 1,
  "username": "player1",
  "chips": 1500,
  "wins": 15,
  "losses": 10,
  "total_games": 25,
  "win_rate": 60.0,
  "total_actions": 450
}
```

---

## Leaderboard Endpoints

### Top Players by Wins
```http
GET /api/leaderboard/wins?limit=10
Authorization: Bearer <token>
```

**Response:**
```json
[
  {
    "user_id": 1,
    "username": "player1",
    "score": 50,
    "rank": 1
  }
]
```

### Top Players by Chips
```http
GET /api/leaderboard/chips?limit=10
Authorization: Bearer <token>
```

**Response:**
```json
[
  {
    "user_id": 2,
    "username": "player2",
    "score": 5000,
    "rank": 1
  }
]
```

---

## Matchmaking Endpoints

### Join Queue
```http
POST /api/matchmaking/join
Authorization: Bearer <token>
Content-Type: application/json

{
  "chips": 1000,
  "skill_rank": 1500
}
```

**Response:**
```json
{
  "status": "joined",
  "queue_size": 5
}
```

### Leave Queue
```http
POST /api/matchmaking/leave
Authorization: Bearer <token>
```

**Response:**
```json
{
  "status": "left"
}
```

---

## Admin Endpoints

### List All Rooms (Admin)
```http
GET /admin/rooms
Authorization: Bearer <admin_token>
```

**Response:**
```json
[
  {
    "id": 1,
    "name": "Room 1",
    "status": "playing",
    "players": 6
  }
]
```

### Ban User
```http
POST /admin/ban
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "user_id": 123,
  "reason": "Cheating detected",
  "permanent": true
}
```

**Response:**
```json
{
  "status": "user banned"
}
```

### Unban User
```http
POST /admin/unban
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "user_id": 123
}
```

**Response:**
```json
{
  "status": "user unbanned"
}
```

---

## WebSocket Connection

### Connect
```
ws://localhost:8080/ws?user_id=1&username=player1&room_id=room1
```

### Client → Server Messages

#### Player Action
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

### Server → Client Messages

#### Game Start
```json
{
  "type": "game_start",
  "room_id": 1,
  "game": {...}
}
```

#### Deal Cards
```json
{
  "type": "deal",
  "room_id": 1,
  "phase": "flop",
  "community_cards": [...]
}
```

#### Player Action
```json
{
  "type": "player_action",
  "room_id": 1,
  "player_id": 1,
  "username": "player1",
  "action": "raise",
  "amount": 100
}
```

#### Game End
```json
{
  "type": "game_end",
  "room_id": 1,
  "winners": [
    {
      "player_id": 1,
      "username": "player1",
      "hand": "Full House",
      "cards": [...]
    }
  ]
}
```

---

## Error Responses

### 400 Bad Request
```json
{
  "error": "Invalid request"
}
```

### 401 Unauthorized
```json
{
  "error": "unauthorized"
}
```

### 403 Forbidden
```json
{
  "error": "user is banned"
}
```

### 429 Too Many Requests
```json
{
  "error": "rate limit exceeded",
  "retry_after": 45
}
```

### 500 Internal Server Error
```json
{
  "error": "internal server error"
}
```

---

## Rate Limits

- **API Endpoints**: 100 requests per minute per user
- **WebSocket**: Maximum 5 concurrent connections per user

## Anti-Cheat Validation

All actions are validated for:
- Latency (max 5000ms)
- Action rate (max 60/min, min 100ms interval)
- Bet amounts (sufficient chips, valid raises)
- Player state (not folded, not all-in)

Invalid actions return 400 with error message.
