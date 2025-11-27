# WebSocket Connection Fix

## Issue
When navigating to a game room, the WebSocket was trying to connect before user data was loaded from localStorage, causing the error:
```
[WARN] Cannot connect: user data not loaded
```

## Root Cause
The `usePokerWebSocket` hook was being called immediately when the component mounted, but the user data was being loaded asynchronously in a `useEffect`. This created a race condition where:

1. Component renders
2. `usePokerWebSocket` hook initializes with `userId: 0` and `username: ''`
3. WebSocket tries to connect with invalid user data
4. `useEffect` loads user data from localStorage
5. User data updates but WebSocket already failed

## Solution

### 1. Added `enabled` prop to WebSocket hook
```typescript
interface UsePokerWebSocketProps {
  roomId: string
  userId: number
  username: string
  enabled?: boolean  // NEW: Control when to connect
  onConnect?: () => void
  onDisconnect?: () => void
  onError?: (error: Event) => void
}
```

### 2. Updated connection logic
```typescript
const connect = useCallback(() => {
  if (ws.current?.readyState === WebSocket.OPEN) return
  
  // Don't connect if disabled or user data is not loaded yet
  if (!enabled || !userId || !username) {
    if (!enabled) {
      logger.debug('WebSocket connection disabled')
    } else {
      logger.warn('Cannot connect: user data not loaded')
    }
    return
  }
  // ... rest of connection logic
}, [roomId, userId, username, enabled, onConnect, onDisconnect, onError])
```

### 3. Updated useEffect to respect enabled flag
```typescript
useEffect(() => {
  // Only connect if enabled and user data is available
  if (enabled && userId && username) {
    connect()
  }

  return () => {
    disconnect()
  }
}, [enabled, userId, username, connect, disconnect])
```

### 4. Updated game page to pass enabled flag
```typescript
const { isConnected, isReconnecting, sendAction, sendChat, reconnect } = usePokerWebSocket({
  roomId: params.roomId as string,
  userId: user?.id || 0,
  username: user?.username || '',
  enabled: !!user, // Only connect when user is loaded
  onConnect: () => {
    addToast('success', 'Connected to game')
  },
  onDisconnect: () => {
    addToast('error', 'Disconnected from game')
  }
})
```

### 5. Added loading state
```typescript
// Show loading state while user data is being loaded
if (!user) {
  return (
    <div className="fixed inset-0 bg-black flex items-center justify-center">
      <div className="text-center">
        <div className="animate-spin rounded-full h-16 w-16 border-t-2 border-b-2 border-emerald-500 mx-auto mb-4"></div>
        <p className="text-white/60">Loading game...</p>
      </div>
    </div>
  )
}
```

## Benefits

1. **No more race conditions**: WebSocket only connects after user data is confirmed loaded
2. **Better UX**: Shows loading spinner while user data loads
3. **Cleaner logs**: No more warning messages about missing user data
4. **More reliable**: Connection happens at the right time with valid data
5. **Reusable**: The `enabled` prop can be used in other scenarios where conditional connection is needed

## Testing

To verify the fix:
1. Clear localStorage
2. Login to the application
3. Create or join a room
4. Navigate to the game page
5. Check console - should see "Connected to game" without warnings
6. Verify WebSocket connection is established with correct user data

## Files Modified

- `frontend/hooks/usePokerWebSocket.ts` - Added enabled prop and updated connection logic
- `frontend/app/game/[roomId]/page.tsx` - Added enabled flag and loading state

## Related Issues

This fix also prevents:
- Multiple WebSocket connection attempts
- Invalid user_id being sent to server
- Potential authentication issues
- Race conditions on page navigation
