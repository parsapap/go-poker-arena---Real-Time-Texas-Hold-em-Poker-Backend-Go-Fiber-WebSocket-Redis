# Security Features

## Authentication & Authorization

### JWT Authentication
- **Signup**: `/auth/signup` - Create new account with bcrypt password hashing
- **Login**: `/auth/login` - Authenticate and receive JWT token
- **Token Expiration**: 24 hours
- **Algorithm**: HS256

### Password Security
- Bcrypt hashing with default cost (10)
- Passwords never stored in plain text
- Passwords never returned in API responses

### User Model
```go
type User struct {
    ID        uint
    Username  string  // Unique
    Email     string  // Unique
    Password  string  // Bcrypt hashed
    Chips     int64
    Wins      int64
    Losses    int64
    IsAdmin   bool
    IsBanned  bool
    LastIP    string
}
```

## Anti-Cheat Measures

### 1. Latency Checks
- Maximum latency: 5000ms
- Detects network manipulation
- Validates action timestamps

### 2. Action Rate Limiting
- Minimum interval between actions: 100ms
- Maximum actions per minute: 60
- Detects bot behavior

### 3. Bet Validation
- Validates bet amounts against player chips
- Checks minimum raise requirements
- Prevents invalid actions (check when bet required)
- Validates player state (not folded, not all-in)

### 4. Card Hiding
- Opponent cards hidden until showdown
- Partial game state sent to clients
- Only shows card count, not actual cards
- Full reveal only at showdown phase

### 5. Collusion Detection
- Tracks player action patterns
- Flags suspicious behavior
- Can be extended with ML models

## Game History

### Database Storage
All games and actions saved to PostgreSQL:

```sql
game_histories
- id, game_id, room_id, winner_id
- pot, players (jsonb), actions (text)
- duration, final_hands (jsonb)

player_actions
- id, game_id, user_id
- action, amount, phase
- timestamp, latency
```

### Endpoints
- `GET /api/users/:id/history` - User's game history
- `GET /api/users/:id/stats` - Player statistics
  - Total games, wins, losses
  - Win rate percentage
  - Total actions

## Admin Features

### Admin Middleware
- Requires JWT authentication
- Checks `is_admin` flag on user
- Returns 403 for non-admin users

### Admin Endpoints
- `GET /admin/rooms` - View all rooms
- `POST /admin/ban` - Ban user
  ```json
  {
    "user_id": 123,
    "reason": "Cheating detected",
    "permanent": true
  }
  ```
- `POST /admin/unban` - Unban user

### Ban System
```go
type BanRecord struct {
    UserID    uint
    AdminID   uint
    Reason    string
    ExpiresAt *time.Time
    Permanent bool
}
```

## CORS & HTTPS

### CORS Configuration
```go
cors.Config{
    AllowOrigins:     "https://yourdomain.com",
    AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
    AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
    AllowCredentials: true,
}
```

### HTTPS Ready
- Multi-stage Docker build
- CA certificates included
- Supports TLS termination at load balancer
- Environment variable for allowed origins

## Rate Limiting

### API Rate Limits
- 100 requests per minute per user
- Redis-based tracking
- Returns 429 with retry_after

### WebSocket Limits
- Maximum 5 concurrent connections per user
- Automatic cleanup on disconnect
- Connection count tracked in Redis

## Security Best Practices

### 1. Input Validation
- All request bodies validated
- Parameter type checking
- SQL injection prevention (GORM)

### 2. Error Handling
- Generic error messages to clients
- Detailed logs server-side
- No stack traces exposed

### 3. Database Security
- Prepared statements (GORM)
- Connection pooling
- Soft deletes for audit trail

### 4. Redis Security
- Password authentication
- Connection encryption support
- Key expiration for temporary data

## Deployment Security

### Multi-Stage Docker Build
```dockerfile
# Stage 1: Build
FROM golang:1.21-alpine AS builder
# ... build process

# Stage 2: Runtime
FROM scratch
# Minimal attack surface
# No shell, no package manager
# Only application binary
```

### Environment Variables
- Never commit `.env` file
- Use secrets management in production
- Rotate JWT secret regularly

### Monitoring
- Prometheus metrics for suspicious activity
- Track failed login attempts
- Monitor action latency patterns
- Alert on rate limit violations

## Vulnerability Reporting

If you discover a security vulnerability, please email: security@poker-arena.com

Do not create public GitHub issues for security vulnerabilities.

## Security Checklist

- [x] Password hashing (bcrypt)
- [x] JWT authentication
- [x] Rate limiting
- [x] Anti-cheat validation
- [x] Card hiding
- [x] Game history logging
- [x] Admin controls
- [x] Ban system
- [x] CORS configuration
- [x] HTTPS support
- [x] Input validation
- [x] SQL injection prevention
- [x] Multi-stage Docker build
- [x] Minimal container image
- [x] Environment variable security

## Future Enhancements

- [ ] 2FA authentication
- [ ] IP-based geolocation
- [ ] Advanced ML-based cheat detection
- [ ] Automated ban appeals system
- [ ] Audit log for admin actions
- [ ] Session management
- [ ] Password reset flow
- [ ] Email verification
