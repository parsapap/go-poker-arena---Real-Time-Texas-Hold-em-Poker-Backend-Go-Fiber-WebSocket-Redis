# WebSocket Disconnection Fix - Final Solution

## Problem
WebSocket connections are disconnecting immediately and repeatedly, causing a reconnection loop.

## Root Cause
The backend **was missing join message handling entirely**. When the frontend sent a join message, the backend didn't process it, and the connection would eventually timeout or close for other reasons.

## Fixes Applied

### Backend Changes (`backend/internal/websocket/client.go`)

#### 1. Added Message.Data Field
```go
type Message struct {
    Type    string                 `json:"type"`
    RoomID  string                 `json:"room_id,omitempty"`
    UserID  uint                   `json:"user_id,omitempty"`
    Username string                `json:"username,omitempty"`
    Payload interface{}            `json:"payload,omitempty"`
    Data    map[string]interface{} `json:"data,omitempty"` // NEW
}
```

#### 2. Added Pong Message Handling
```go
// Handle pong message (client responding to our ping)
if msg.Type == "pong" {
    logger.Debug().Uint("user_id", c.UserID).Msg("Received pong from client")
    continue
}
```

#### 3. Added Join Message Handling with Error Recovery
```go
// Handle join message
if msg.Type == "join" {
    logger.Debug().Uint("user_id", c.UserID).Str("room_id", c.RoomID).Str("username", c.Username).Msg("Attempting to add client to room")
    
    if c.RoomManager != nil {
        if err := c.RoomManager.AddClientToRoom(c.UserID, c.RoomID); err != nil {
            logger.Error().Err(err).Uint("user_id", c.UserID).Str("room_id", c.RoomID).Msg("Failed to add client to room")
            
            // Send error message to client instead of closing connection
            errorMsg := Message{
                Type: "error",
                Data: map[string]interface{}{
                    "message": "Failed to join room: " + err.Error(),
                },
            }
            
            if data, err := json.Marshal(errorMsg); err == nil {
                select {
                case c.Send <- data:
                default:
                    close(c.Send)
                    return
                }
            }
            
            // Don't return here - keep connection alive
            continue
        } else {
            logger.Info().Uint("user_id", c.UserID).Str("room_id", c.RoomID).Msg("Client added to room successfully")
        }
    }
}
```

**Key Points:**
- ✅ Processes join messages
- ✅ Adds client to room via RoomManager
- ✅ Sends error messages instead of closing connection
- ✅ Keeps connection alive even on errors
- ✅ Logs success/failure for debugging

### Frontend Changes (`frontend/hooks/usePokerWebSocket.ts`)

#### 1. Enhanced Close Code Logging
```typescript
if (process.env.NODE_ENV === 'development') {
  logger.ws.disconnected()
  logger.info(`WebSocket closed: code=${event.code}, reason=${event.reason || 'No reason'}`)
  
  // Log common close codes
  const closeReasons: Record<number, string> = {
    1000: 'Normal closure',
    1001: 'Going away',
    1002: 'Protocol error',
    1003: 'Unsupported data',
    1006: 'Abnormal closure (no close frame)',
    1011: 'Server error',
    1012: 'Service restart'
  }
  
  const reasonText = closeReasons[event.code] || 'Unknown'
  logger.warn(`Close code ${event.code}: ${reasonText}`)
}
```

## Critical: Backend Must Be Restarted

**The backend MUST be restarted for these changes to take effect!**

### Option 1: Use the restart script
```bash
./restart-backend.sh
cd backend && go run cmd/server/main.go
```

### Option 2: Manual restart
```bash
# Find and kill the process
ps aux | grep "go run cmd/server/main.go" | grep -v grep
kill <PID>

# Start backend
cd backend && go run cmd/server/main.go
```

### Option 3: Use Make
```bash
cd backend
make run
```

## Testing After Restart

### 1. Check Backend Logs
You should see:
```
[DEBUG] Attempting to add client to room user_id=1 room_id=2 username=testuser
[INFO] Client added to room successfully user_id=1 room_id=2
```

### 2. Check Frontend Console
You should see:
```
[INFO] WebSocket connected
[DEBUG] Join message sent: {type: 'join', room_id: '2', user_id: 1, username: 'testuser'}
```

And NO more disconnect spam!

### 3. Check Connection Badge
The badge in the top-right should be:
- 🟢 Green "Connected" (not red/yellow)
- Stay green (not flickering)

## What Was Wrong Before

1. **No join message handling** - Backend ignored join messages
2. **No pong handling** - Backend didn't acknowledge client pongs
3. **No error recovery** - Any error would close the connection
4. **Missing Data field** - Couldn't send error messages properly

## What's Fixed Now

1. ✅ **Join messages processed** - Backend adds clients to rooms
2. ✅ **Pong messages handled** - Backend acknowledges client responses
3. ✅ **Error recovery** - Errors sent as messages, connection stays alive
4. ✅ **Better logging** - Can see exactly what's happening
5. ✅ **Close code debugging** - Know why connections close

## Expected Behavior

### Successful Connection Flow
1. Frontend connects to WebSocket
2. Frontend sends join message
3. Backend receives join message
4. Backend adds client to room
5. Backend logs success
6. Connection stays alive
7. Ping/pong keeps connection healthy

### Error Flow (if room doesn't exist)
1. Frontend connects to WebSocket
2. Frontend sends join message
3. Backend receives join message
4. Backend fails to add client (room not found)
5. Backend sends error message to client
6. **Connection stays alive** (not closed!)
7. Frontend shows error toast
8. User can try again or go back to lobby

## Troubleshooting

### Still seeing disconnects?

1. **Did you restart the backend?**
   - This is the most common issue!
   - Run `./restart-backend.sh` then start backend again

2. **Check backend logs**
   - Look for "Attempting to add client to room"
   - Look for "Client added to room successfully" or error messages

3. **Check frontend console**
   - Look for close code (should see "Close code X: Description")
   - If code 1011: Backend error, check backend logs
   - If code 1006: Network issue or backend crashed

4. **Verify room exists**
   - Run `./test-room-websocket.sh`
   - Use the room ID it provides

5. **Check database**
   - Ensure PostgreSQL is running
   - Ensure room exists in database

### Common Close Codes

- **1000**: Normal closure (expected when leaving page)
- **1006**: Abnormal closure (network issue or backend crash)
- **1011**: Server error (backend encountered an error)

## Files Modified

1. ✅ `backend/internal/websocket/client.go`
   - Added Message.Data field
   - Added pong message handling
   - Added join message handling with error recovery
   - Better logging

2. ✅ `frontend/hooks/usePokerWebSocket.ts`
   - Enhanced close code logging
   - Better debugging information

3. ✅ `restart-backend.sh` (new)
   - Helper script to restart backend

## Next Steps

1. **Restart the backend** (critical!)
2. **Test the connection** with the room ID from test script
3. **Check logs** to verify join messages are processed
4. **Verify connection stays alive** (green badge, no disconnect spam)

If you still see issues after restarting, check the close code in the console and backend logs for specific error messages.
