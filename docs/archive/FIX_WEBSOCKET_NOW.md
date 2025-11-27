# Fix WebSocket Disconnection - Simple Steps

## The Problem
WebSocket connects then immediately disconnects repeatedly.

## The Solution (90% of cases)

### You're trying to join a room that doesn't exist!

## How to Fix - 3 Steps:

### 1. Go to the Lobby
http://localhost:3000

### 2. Create a Room
- Click the "Create Room" button
- Fill in:
  - Name: "My Room"
  - Max Players: 6
  - Small Blind: 10
  - Big Blind: 20
- Click "Create"

### 3. Join the Room
- Click "Join" on the room you just created
- WebSocket should now stay connected! ✅

---

## Alternative: Create Room via Command

```bash
# Get your token (check browser localStorage or login again)
./test-auth.sh

# Use the token from above
TOKEN="your_token_here"

# Create a room
curl -X POST http://localhost:8080/api/rooms \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Room","max_players":6,"small_blind":10,"big_blind":20}'

# You'll get back a room with an ID, use that ID to join
```

---

## Test the Fix

1. Open test-websocket.html
2. Enter your user ID and username
3. Enter the room ID you just created
4. Click "Connect"
5. Should stay connected! ✅

---

## Still Not Working?

### Check Backend Logs
When you join a room, backend should show:
```
WebSocket connected user_id=X username=Y room_id=Z
```

If you see errors → Share the error message

### Check Frontend Console (F12)
Should see:
```
Connecting to WebSocket: ws://localhost:8080/ws?...
WebSocket connected
Join message sent
```

If you see "WebSocket disconnected" immediately → Room doesn't exist

---

## Summary

**The issue:** Trying to join a non-existent room  
**The fix:** Create a room first, then join it  
**Result:** WebSocket stays connected ✅

That's it! Create a room first, then everything will work.
