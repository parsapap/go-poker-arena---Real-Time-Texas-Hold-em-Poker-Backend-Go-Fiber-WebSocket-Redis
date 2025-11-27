# WebSocket Connection Final Debug

## Current Status

✅ Backend WebSocket endpoint is working (returns 101 Switching Protocols)  
✅ Login works  
❌ WebSocket connects then immediately disconnects repeatedly

## The Problem

The WebSocket is connecting successfully but then closing immediately. This creates a reconnection loop.

## Root Cause Analysis

Looking at the logs, the pattern is:
1. WebSocket connects
2. Immediately disconnects
3. Reconnects (exponential backoff)
4. Disconnects again
5. Repeat...

This suggests the backend is closing the connection right after it opens, likely because:
1. The Hub is not running
2. The client registration is failing
3. There's a panic in the WebSocket handler

## Solution

### Step 1: Check Backend Logs

When you start the backend, you should see:
```
Server starting on port 8080
Database connected successfully
Redis connected successfully
Auto-matchmaking worker started
```

**Important:** Look for any error messages or panics when a WebSocket connects.

### Step 2: Verify Hub is Running

The Hub must be running for WebSocket connections to work. Check `backend/cmd/server/main.go`:

```go
hub := websocket.NewHub(redisClient)
go hub.Run()  // ← This MUST be called
```

### Step 3: Test with Simple WebSocket Client

Use the test tool:
```bash
open test-websocket.html
```

1. Enter your user ID (from login)
2. Enter your username
3. Enter room ID: `1`
4. Click "Connect"
5. Watch the log for connection status

If it connects and stays connected → Backend is fine, frontend issue  
If it connects and disconnects → Backend issue

## Quick Fix

The most likely issue is that the backend is panicking when trying to process the join message. Let's add better error handling:

### Backend Fix (if needed)

Check `backend/cmd/server/main.go` around line 280:

```go
app.Get("/ws", ws.New(func(c *ws.Conn) {
    metrics.WebSocketConnections.Inc()
    defer metrics.WebSocketConnections.Dec()

    userID := c.Query("user_id", "0")
    username := c.Query("username", "guest")
    roomID := c.Query("room_id", "")

    var uid uint
    fmt.Sscanf(userID, "%d", &uid)

    // Add validation
    if uid == 0 {
        logger.Warn().Msg("Invalid user_id in WebSocket connection")
        c.WriteMessage(websocket.CloseMessage, []byte("Invalid user_id"))
        c.Close()
        return
    }

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
    logger.Info().Uint("user_id", uid).Str("username", username).Str("room_id", roomID).Msg("WebSocket connected")

    go client.WritePump()
    client.ReadPump()  // This blocks until connection closes

    logger.Info().Uint("user_id", uid).Msg("WebSocket disconnected")
}))
```

## Debugging Steps

### 1. Check Backend Console

When you navigate to a room, the backend should log:
```
WebSocket connected user_id=1 username=testuser room_id=1
```

If you don't see this → Hub not running or connection rejected

### 2. Check Frontend Console

Should see:
```
Connecting to WebSocket: ws://localhost:8080/ws?user_id=1&username=testuser&room_id=1
WebSocket connected
Join message sent
```

If you see "WebSocket disconnected" immediately after → Backend closing connection

### 3. Check Network Tab

1. Open DevTools (F12)
2. Go to Network tab
3. Filter by "WS"
4. Click on the WebSocket connection
5. Check "Messages" tab
6. Should see messages being sent/received

If no messages → Connection closing before any communication

## Common Issues

### Issue 1: Hub Not Running
**Symptom:** Connection closes immediately, no backend logs  
**Fix:** Ensure `go hub.Run()` is called in main.go

### Issue 2: Invalid User ID
**Symptom:** Connection closes, backend logs "Invalid user_id"  
**Fix:** Ensure user is logged in and user data is in localStorage

### Issue 3: Room Doesn't Exist
**Symptom:** Connection works but no game state  
**Fix:** Create a room first in the lobby

### Issue 4: CORS Issues
**Symptom:** Connection fails with CORS error  
**Fix:** Check backend CORS configuration

## Testing Procedure

### Test 1: Backend Health
```bash
curl http://localhost:8080/healthz
# Should return: {"status":"ok",...}
```

### Test 2: WebSocket with curl
```bash
# This should stay connected (press Ctrl+C to stop)
curl -i -N -H "Connection: Upgrade" -H "Upgrade: websocket" \
  -H "Sec-WebSocket-Version: 13" -H "Sec-WebSocket-Key: test" \
  "http://localhost:8080/ws?user_id=1&username=test&room_id=1"
```

### Test 3: WebSocket Test Tool
```bash
open test-websocket.html
```

### Test 4: Check Backend Logs
Look for:
- "WebSocket connected" messages
- Any error or panic messages
- "WebSocket disconnected" messages

## Expected Flow

1. User logs in → Token saved
2. User navigates to room → Loading screen
3. User data loads → WebSocket enabled
4. WebSocket connects → Backend logs "WebSocket connected"
5. Join message sent → Backend registers client
6. Game state sent → Frontend renders table
7. Connection stays open → Ping/pong every 30s

## If Still Not Working

### Restart Everything
```bash
# Stop backend (Ctrl+C)
# Stop frontend (Ctrl+C)

# Start backend
cd backend
go run cmd/server/main.go

# Start frontend (new terminal)
cd frontend
npm run dev
```

### Check Logs Carefully

Backend should show:
```
Server starting on port 8080
Database connected successfully
Redis connected successfully
Auto-matchmaking worker started
WebSocket connected user_id=1 username=testuser room_id=1
```

Frontend console should show:
```
Connecting to WebSocket: ws://localhost:8080/ws?...
WebSocket connected
Join message sent
```

### Share Logs

If still not working, share:
1. Backend console output (when WebSocket connects)
2. Frontend console output (full errors)
3. Network tab WebSocket details
4. test-websocket.html results

## Most Likely Fix

The issue is probably that the Hub is not running. Check your backend main.go file and ensure this line exists:

```go
hub := websocket.NewHub(redisClient)
go hub.Run()  // ← Make sure this is here!
```

This starts the Hub in a goroutine, which is essential for WebSocket connections to work.
