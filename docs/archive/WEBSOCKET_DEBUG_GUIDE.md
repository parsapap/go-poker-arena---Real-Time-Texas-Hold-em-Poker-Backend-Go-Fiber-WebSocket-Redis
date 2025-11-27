# WebSocket Connection Debug Guide

## Current Issues

Based on the error screenshot:
1. ✅ WebSocket connects successfully ("Connected to game")
2. ❌ WebSocket immediately disconnects ("Disconnected from game")
3. ❌ Multiple reconnection attempts fail
4. ❌ "WebSocket not connected, cannot send message" errors

## Root Cause

The WebSocket is connecting but then immediately closing. This could be due to:

1. **Backend rejecting the connection**
2. **CORS issues**
3. **Invalid WebSocket upgrade**
4. **Backend not handling the connection properly**

## Fixes Applied

### 1. Fixed Circular Dependency in sendMessage
**Problem:** `sendMessage` was being called in `onopen` before it was defined.

**Solution:** Send messages directly in `onopen` and `disconnect`:
```typescript
// Before (circular dependency)
ws.current.onopen = () => {
  sendMessage({ type: 'join', ... }) // sendMessage not defined yet!
}

// After (direct send)
ws.current.onopen = () => {
  if (ws.current?.readyState === WebSocket.OPEN) {
    ws.current.send(JSON.stringify({
      type: 'join',
      room_id: roomId,
      user_id: userId,
      username
    }))
  }
}
```

### 2. Added Better Logging
```typescript
logger.info(`Connecting to WebSocket: ${url}`)
logger.info(`WebSocket closed: code=${event.code}, reason=${event.reason}`)
```

### 3. Prevent Unnecessary Reconnections
```typescript
// Don't reconnect if it was a clean close
if (!enabled || event.code === 1000) {
  logger.info('Clean disconnect, not reconnecting')
  return
}
```

### 4. Added Max Reconnection Message
```typescript
if (reconnectAttempts.current >= maxReconnectAttempts) {
  addChatMessage({
    user: 'System',
    message: 'Connection lost. Please refresh the page.',
    timestamp: Date.now()
  })
}
```

## Debugging Steps

### Step 1: Check Backend is Running
```bash
# In backend directory
cd backend
go run cmd/server/main.go

# Should see:
# Server starting on port 8080
# Database connected successfully
# Redis connected successfully
```

### Step 2: Check Backend WebSocket Endpoint
The backend WebSocket endpoint should be at:
```
ws://localhost:8080/ws?user_id=1&username=testuser&room_id=1
```

Check backend logs for:
- WebSocket connection attempts
- Any error messages
- Connection close reasons

### Step 3: Test WebSocket Connection Manually

Using browser console:
```javascript
// Test WebSocket connection
const ws = new WebSocket('ws://localhost:8080/ws?user_id=1&username=test&room_id=1')

ws.onopen = () => console.log('Connected!')
ws.onclose = (e) => console.log('Closed:', e.code, e.reason)
ws.onerror = (e) => console.error('Error:', e)
ws.onmessage = (e) => console.log('Message:', e.data)

// Send a test message
ws.send(JSON.stringify({ type: 'ping' }))
```

### Step 4: Check Browser Network Tab
1. Open DevTools (F12)
2. Go to Network tab
3. Filter by "WS" (WebSocket)
4. Look for the WebSocket connection
5. Check:
   - Status (should be 101 Switching Protocols)
   - Messages sent/received
   - Close code and reason

### Step 5: Check Backend WebSocket Handler

The backend should have something like:
```go
// backend/cmd/server/main.go
app.Get("/ws", ws.New(func(c *ws.Conn) {
    userID := c.Query("user_id", "0")
    username := c.Query("username", "guest")
    roomID := c.Query("room_id", "")
    
    // Create client
    client := &websocket.Client{
        Hub:         hub,
        Conn:        c,
        Send:        make(chan []byte, 256),
        UserID:      uid,
        Username:    username,
        RoomID:      roomID,
        RoomManager: roomManager,
    }
    
    hub.Register <- client
    
    go client.WritePump()
    client.ReadPump() // This blocks until connection closes
}))
```

## Common Issues and Solutions

### Issue 1: Backend Not Running
**Symptom:** Connection refused errors
**Solution:**
```bash
cd backend
go run cmd/server/main.go
```

### Issue 2: CORS Issues
**Symptom:** Connection fails with CORS error
**Solution:** Check backend CORS configuration:
```go
app.Use(cors.New(cors.Config{
    AllowOrigins:     "http://localhost:3000",
    AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
    AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
    AllowCredentials: true,
}))
```

### Issue 3: WebSocket Upgrade Failed
**Symptom:** 400 or 426 error
**Solution:** Ensure WebSocket middleware is configured:
```go
app.Use("/ws", func(c *fiber.Ctx) error {
    if ws.IsWebSocketUpgrade(c) {
        return c.Next()
    }
    return fiber.ErrUpgradeRequired
})
```

### Issue 4: Backend Closes Connection Immediately
**Symptom:** Connects then immediately disconnects
**Possible Causes:**
1. Backend validation failing (invalid user_id, room_id)
2. Backend panic/error in WebSocket handler
3. Hub not running
4. Client not properly registered

**Solution:** Check backend logs for errors

### Issue 5: Frontend Sending Messages Before Connection Ready
**Symptom:** "WebSocket not connected" errors
**Solution:** Already fixed - messages now sent only when `readyState === WebSocket.OPEN`

## Testing Checklist

- [ ] Backend server is running on port 8080
- [ ] Frontend server is running on port 3000
- [ ] Redis is running (check with `redis-cli ping`)
- [ ] PostgreSQL is running (check with `psql -U poker -d poker_arena`)
- [ ] User is logged in (check localStorage for 'token' and 'user')
- [ ] Room exists (check database or create new room)
- [ ] Browser console shows WebSocket connection attempt
- [ ] Backend logs show WebSocket connection
- [ ] No CORS errors in browser console
- [ ] WebSocket status is 101 in Network tab

## Quick Fix Commands

### Restart Everything
```bash
# Terminal 1 - Stop and restart backend
cd backend
# Ctrl+C to stop
go run cmd/server/main.go

# Terminal 2 - Stop and restart frontend
cd frontend
# Ctrl+C to stop
npm run dev

# Terminal 3 - Check Docker services
docker-compose ps
# If not running:
docker-compose up -d postgres redis
```

### Clear Browser State
```javascript
// In browser console
localStorage.clear()
sessionStorage.clear()
// Then refresh page (Ctrl+Shift+R)
```

### Check Services
```bash
# Check if backend is responding
curl http://localhost:8080/healthz

# Check if Redis is running
redis-cli ping

# Check if PostgreSQL is running
docker-compose ps postgres
```

## Expected Behavior

1. User logs in → token and user data saved to localStorage
2. User navigates to game page → loading spinner shows
3. User data loads → WebSocket connection initiated
4. WebSocket connects → "Connected to game" toast
5. Join message sent → Backend registers client
6. Game state received → Players and table render
7. User can interact → Actions sent via WebSocket
8. User leaves → Disconnect message sent, connection closed cleanly

## Next Steps

1. **Check backend logs** - Look for WebSocket connection messages and any errors
2. **Test WebSocket manually** - Use browser console to test connection
3. **Check Network tab** - Look at WebSocket status and messages
4. **Verify services** - Ensure all services (backend, Redis, PostgreSQL) are running
5. **Check user data** - Verify user is logged in and data is in localStorage

If issues persist, please share:
- Backend console output
- Browser console output (full errors)
- Network tab WebSocket details
- Backend WebSocket handler code
