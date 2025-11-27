# WebSocket Disconnect Solution

## Problem
WebSocket connects but immediately disconnects, creating a reconnection loop.

## Quick Diagnosis

Open `test-websocket.html` in your browser and test the connection:
```bash
open test-websocket.html
```

1. Enter your user ID (check localStorage in browser)
2. Enter your username
3. Enter room ID: `1`
4. Click "Connect"

**If it stays connected:** Frontend issue  
**If it disconnects:** Backend issue

## Most Common Causes & Solutions

### 1. Room Doesn't Exist ⭐ MOST LIKELY

**Problem:** You're trying to join a room that doesn't exist in the database.

**Solution:** Create a room first!

1. Go to lobby: http://localhost:3000
2. Click "Create Room" button
3. Fill in room details:
   - Name: "Test Room"
   - Max Players: 6
   - Small Blind: 10
   - Big Blind: 20
4. Click "Create"
5. Then click "Join" on the room you just created

**Or create via API:**
```bash
# Get your token from localStorage
TOKEN="your_token_here"

# Create a room
curl -X POST http://localhost:8080/api/rooms \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Room","max_players":6,"small_blind":10,"big_blind":20}'
```

### 2. Invalid Room ID

**Problem:** Room ID in URL doesn't match any room in database.

**Check existing rooms:**
```bash
TOKEN="your_token_here"
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/rooms
```

**Solution:** Use a valid room ID from the list above.

### 3. Backend Not Handling Join Message

**Problem:** Backend receives join message but doesn't know how to handle it.

**Check backend logs** when you connect. Should see:
```
WebSocket connected user_id=1 username=testuser room_id=1
```

If you see errors or panics → Backend issue

### 4. Hub Not Processing Messages

**Problem:** Hub is running but not broadcasting messages.

**Check:** Backend should have this in main.go:
```go
hub := websocket.NewHub(redisClient)
go hub.Run()  // ← Must be here
```

## Step-by-Step Fix

### Step 1: Verify Backend is Running
```bash
curl http://localhost:8080/healthz
# Should return: {"status":"ok","service":"go-poker-arena","version":"1.0.0"}
```

### Step 2: Check if Rooms Exist
```bash
# Login first to get token
TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser_1764254810","password":"password123"}' \
  | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

# List rooms
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/rooms
```

If no rooms exist, create one:
```bash
curl -X POST http://localhost:8080/api/rooms \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Room","max_players":6,"small_blind":10,"big_blind":20}'
```

### Step 3: Test WebSocket Connection

Use the test tool:
```bash
open test-websocket.html
```

Or test in browser console:
```javascript
// Get user from localStorage
const user = JSON.parse(localStorage.getItem('user'))

// Connect to WebSocket
const ws = new WebSocket(`ws://localhost:8080/ws?user_id=${user.id}&username=${user.username}&room_id=1`)

ws.onopen = () => {
  console.log('✅ Connected!')
  // Send join message
  ws.send(JSON.stringify({
    type: 'join',
    room_id: '1',
    user_id: user.id,
    username: user.username
  }))
}

ws.onclose = (e) => {
  console.log('❌ Closed:', e.code, e.reason)
}

ws.onerror = (e) => {
  console.error('❌ Error:', e)
}

ws.onmessage = (e) => {
  console.log('📨 Message:', e.data)
}
```

### Step 4: Check Backend Logs

When WebSocket connects, backend should log:
```
WebSocket connected user_id=1 username=testuser room_id=1
```

When it disconnects:
```
WebSocket disconnected user_id=1
```

Look for any errors or panics between these messages.

### Step 5: Check Frontend Console

Should see:
```
Connecting to WebSocket: ws://localhost:8080/ws?user_id=1&username=testuser&room_id=1
WebSocket connected
Join message sent
```

If you see "WebSocket disconnected" immediately → Backend closing connection

## Debugging Checklist

- [ ] Backend is running (curl healthz works)
- [ ] User is logged in (check localStorage)
- [ ] Room exists (check /api/rooms)
- [ ] Room ID in URL matches existing room
- [ ] Backend logs show "WebSocket connected"
- [ ] No errors or panics in backend logs
- [ ] test-websocket.html connects successfully
- [ ] Frontend console shows "WebSocket connected"

## Expected Behavior

1. Navigate to room → Loading screen
2. WebSocket connects → "Connected to game" toast
3. Join message sent → Backend registers client
4. Game state received → Table renders
5. Connection stays open → Can see other players

## If Still Not Working

### Collect Debug Info

1. **Backend logs** (when WebSocket connects):
   ```
   Copy the last 20 lines from backend console
   ```

2. **Frontend console** (F12 → Console):
   ```
   Copy all WebSocket-related messages
   ```

3. **Network tab** (F12 → Network → WS):
   - Click on WebSocket connection
   - Check Status (should be 101)
   - Check Messages tab
   - Screenshot if possible

4. **test-websocket.html results**:
   - Does it connect?
   - Does it stay connected?
   - Any error messages?

### Common Error Messages

**"WebSocket disconnected" immediately**
→ Backend closing connection, check backend logs

**"Failed to construct 'WebSocket'"**
→ Invalid WebSocket URL, check NEXT_PUBLIC_WS_URL

**"WebSocket connection failed"**
→ Backend not running or wrong port

**"Cannot connect: user data not loaded"**
→ User not logged in, login first

## Quick Test Script

```bash
#!/bin/bash

# Test everything
echo "Testing WebSocket connection..."

# 1. Check backend
curl -s http://localhost:8080/healthz && echo "✅ Backend running" || echo "❌ Backend not running"

# 2. Login and get token
TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser_1764254810","password":"password123"}' \
  | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -n "$TOKEN" ]; then
  echo "✅ Login successful"
  
  # 3. Check rooms
  ROOMS=$(curl -s -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/rooms)
  ROOM_COUNT=$(echo "$ROOMS" | grep -o '"id"' | wc -l)
  echo "✅ Found $ROOM_COUNT rooms"
  
  if [ "$ROOM_COUNT" -eq 0 ]; then
    echo "⚠️  No rooms exist, creating one..."
    curl -s -X POST http://localhost:8080/api/rooms \
      -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json" \
      -d '{"name":"Test Room","max_players":6,"small_blind":10,"big_blind":20}'
    echo "✅ Room created"
  fi
else
  echo "❌ Login failed"
fi
```

## Summary

The most common issue is trying to join a room that doesn't exist. **Create a room first in the lobby**, then join it. The WebSocket will stay connected once you're in a valid room.
