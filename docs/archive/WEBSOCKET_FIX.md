# WebSocket Connection Fix

## Issue
The WebSocket was failing to connect with the error:
```
WebSocket connection to 'ws://localhost:8080/ws?user_id=0&username=&room_id=1' failed
```

## Root Cause
The game page was initializing the WebSocket hook before the user data was loaded from localStorage. This caused the connection to be attempted with:
- `user_id=0` (invalid)
- `username=` (empty)

The backend was rejecting these invalid connections.

## Solution Applied

### 1. Updated WebSocket Hook (`frontend/hooks/usePokerWebSocket.ts`)
Added a check to prevent connection attempts when user data is not loaded:

```typescript
const connect = useCallback(() => {
  if (ws.current?.readyState === WebSocket.OPEN) return
  
  // Don't connect if user data is not loaded yet
  if (!userId || !username) {
    logger.warn('Cannot connect: user data not loaded')
    return
  }
  
  // ... rest of connection logic
}, [roomId, userId, username, onConnect, onDisconnect, onError])
```

### 2. How It Works Now

1. User navigates to game page
2. Component loads and checks localStorage for user data
3. User state is set with valid `id` and `username`
4. WebSocket hook receives valid user data
5. Connection is established successfully

## Testing

After the fix, the WebSocket should:
1. ✅ Wait for user data to load
2. ✅ Connect with valid `user_id` and `username`
3. ✅ Successfully join the game room
4. ✅ Receive and send game messages

## Next Steps

1. Refresh the browser page (the Next.js dev server should auto-reload)
2. Login to the application
3. Create or join a room
4. The WebSocket should now connect successfully

## Verification

Check the browser console - you should see:
- `[INFO] WebSocket reconnecting` messages stop
- `[INFO] Connected to game` toast notification
- No more `user_id=0` connection attempts

Check the backend logs - you should see:
- `WebSocket connected user_id=<valid_id> username=<your_username> room_id=<room_id>`
- No more immediate disconnections
