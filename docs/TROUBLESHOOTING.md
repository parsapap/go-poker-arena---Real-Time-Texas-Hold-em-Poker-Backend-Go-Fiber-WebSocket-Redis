# Troubleshooting Guide

## Common Issues

### 1. Login Error: "invalid credentials"
**Problem:** User doesn't exist in database

**Solution:**
- Register a new account first at http://localhost:3000/register
- Or use test credentials from `./test-auth.sh`

### 2. WebSocket Disconnecting
**Problem:** WebSocket connects then immediately disconnects

**Solution:**
- Create a room first in the lobby
- Then join the room you created
- The room must exist in the database

**Debug:**
```bash
# Test WebSocket connection
open test-websocket.html
```

### 3. NaN Error in Game Page
**Problem:** "The specified value 'NaN' cannot be parsed"

**Solution:** ✅ Fixed in latest version
- Refresh the page
- Clear browser cache if needed

### 4. Backend Not Starting
**Problem:** Backend fails to start

**Check:**
```bash
# Verify Docker services
docker-compose ps

# Should show postgres and redis as "Up"
# If not, start them:
docker-compose up -d postgres redis

# Check backend logs
cd backend
go run cmd/server/main.go
```

### 5. Frontend Not Connecting
**Problem:** "Failed to fetch" errors

**Check:**
```bash
# Verify backend is running
curl http://localhost:8080/healthz

# Should return: {"status":"ok",...}
```

**Fix:**
- Start backend: `cd backend && go run cmd/server/main.go`
- Check `.env.local` has correct API_URL

### 6. CORS Errors
**Problem:** CORS policy blocking requests

**Solution:**
- Check backend CORS config in `backend/cmd/server/main.go`
- Ensure `ALLOWED_ORIGINS` includes your frontend URL
- Default: `http://localhost:3000`

### 7. Database Connection Failed
**Problem:** Backend can't connect to PostgreSQL

**Check:**
```bash
# Verify PostgreSQL is running
docker-compose ps postgres

# Check connection
docker exec -it poker-postgres psql -U poker -d poker_arena -c "SELECT 1;"
```

**Fix:**
```bash
# Restart PostgreSQL
docker-compose restart postgres

# Or recreate
docker-compose down
docker-compose up -d postgres redis
```

### 8. Redis Connection Failed
**Problem:** Backend can't connect to Redis

**Check:**
```bash
# Verify Redis is running
docker-compose ps redis

# Test connection
docker exec -it poker-redis redis-cli ping
# Should return: PONG
```

### 9. Port Already in Use
**Problem:** "address already in use"

**Solution:**
```bash
# Find process using port 8080
lsof -i :8080

# Kill it
kill -9 <PID>

# Or use different port in .env
PORT=8081
```

### 10. Room Not Found
**Problem:** Trying to join non-existent room

**Solution:**
1. Go to lobby: http://localhost:3000
2. Click "Create Room"
3. Fill in details and create
4. Then join the room

## Debug Tools

### Test Authentication
```bash
./test-auth.sh
```
Creates a test user and returns credentials

### Test WebSocket
```bash
open test-websocket.html
```
Interactive WebSocket testing tool

### Check Backend Health
```bash
curl http://localhost:8080/healthz
```

### Check Database
```bash
docker exec -it poker-postgres psql -U poker -d poker_arena

# List users
SELECT id, username, chips FROM users;

# List rooms
SELECT id, name, status FROM rooms;

# Exit
\q
```

### Check Redis
```bash
docker exec -it poker-redis redis-cli

# Check keys
KEYS *

# Exit
exit
```

## Reset Everything

If nothing works, complete reset:

```bash
# 1. Stop all services
docker-compose down

# 2. Remove volumes (WARNING: deletes all data)
docker-compose down -v

# 3. Start fresh
docker-compose up -d postgres redis

# 4. Restart backend
cd backend
go run cmd/server/main.go

# 5. Restart frontend
cd frontend
npm run dev

# 6. Clear browser data
# In browser: Ctrl+Shift+Delete
# Clear cache and localStorage
```

## Getting Help

If issues persist:

1. Check backend console output
2. Check browser console (F12)
3. Check Network tab for failed requests
4. Run test scripts to isolate the issue
5. Check Docker logs: `docker-compose logs`

## Logs Location

- Backend: Console output
- Frontend: Browser console (F12)
- PostgreSQL: `docker-compose logs postgres`
- Redis: `docker-compose logs redis`
