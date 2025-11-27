# Quick Fix Guide - WebSocket Connection Issues

## TL;DR - Quick Steps

### 1. Test WebSocket Connection First
```bash
# Open the test tool in your browser
open test-websocket.html
```
- Click "Connect"
- If it works → Frontend code issue (now fixed)
- If it fails → Backend issue (see below)

### 2. Restart Backend
```bash
cd backend
# Stop current process (Ctrl+C)
go run cmd/server/main.go
```

### 3. Restart Frontend
```bash
cd frontend
# Stop current process (Ctrl+C)
npm run dev
```

### 4. Clear Browser Cache
- Press `Ctrl+Shift+Delete`
- Select "Cached images and files"
- Click "Clear data"
- Or hard refresh: `Ctrl+Shift+R`

### 5. Test Application
1. Go to `http://localhost:3000`
2. Login
3. Create/join a room
4. Check browser console (F12) for errors

## What Was Fixed

✅ WebSocket circular dependency  
✅ User data loading race condition  
✅ Null safety checks  
✅ Error boundary added  
✅ Better logging  
✅ Reconnection logic improved  

## If Still Not Working

### Check Backend is Running
```bash
curl http://localhost:8080/healthz
# Should return: {"status":"ok","service":"go-poker-arena","version":"1.0.0"}
```

### Check Docker Services
```bash
docker-compose ps
# Both postgres and redis should be "Up"
```

### Check Backend Logs
Look for:
- ✅ "Server starting on port 8080"
- ✅ "Database connected successfully"
- ✅ "Redis connected successfully"
- ❌ Any error messages

### Check Browser Console
Look for:
- ✅ "Connecting to WebSocket: ws://localhost:8080/ws?..."
- ✅ "WebSocket connected"
- ❌ Any red error messages

### Check Network Tab (F12 → Network → WS)
- Status should be: `101 Switching Protocols`
- Should see messages being sent/received
- Should NOT immediately close

## Common Error Messages

### "WebSocket not connected, cannot send message"
**Status:** ✅ FIXED  
**What it was:** Trying to send before connection ready  
**Fix applied:** Now checks connection state before sending

### "Cannot connect: user data not loaded"
**Status:** ✅ FIXED  
**What it was:** WebSocket connecting too early  
**Fix applied:** Added `enabled` prop that waits for user data

### "Cannot read property 'map' of undefined"
**Status:** ✅ FIXED  
**What it was:** Arrays undefined before data loads  
**Fix applied:** Added null checks before mapping

### "WebSocket disconnected"
**If this still happens:**
1. Check backend is running
2. Check backend logs for errors
3. Use test-websocket.html to verify backend
4. Check if room exists in database

## Emergency Reset

If nothing works, do a complete reset:

```bash
# 1. Stop everything
# Press Ctrl+C in all terminals

# 2. Reset Docker
docker-compose down
docker-compose up -d postgres redis

# 3. Clear browser
# In browser console:
localStorage.clear()
sessionStorage.clear()

# 4. Restart backend
cd backend
go run cmd/server/main.go

# 5. Restart frontend
cd frontend
npm run dev

# 6. Test with tool first
open test-websocket.html
```

## Success Indicators

You'll know it's working when you see:

1. ✅ test-websocket.html connects successfully
2. ✅ Browser console shows "WebSocket connected"
3. ✅ No error messages in console
4. ✅ Network tab shows 101 status
5. ✅ Game page loads without errors
6. ✅ Can see "Connected to game" toast

## Still Having Issues?

1. Share backend console output
2. Share browser console output (full errors)
3. Share Network tab WebSocket details
4. Run test-websocket.html and share results

## Files to Check

- `frontend/hooks/usePokerWebSocket.ts` - Connection logic
- `frontend/app/game/[roomId]/page.tsx` - Game page
- `backend/cmd/server/main.go` - WebSocket endpoint
- `backend/internal/websocket/hub.go` - WebSocket hub

## Helpful Commands

```bash
# Check if ports are in use
lsof -i :3000  # Frontend
lsof -i :8080  # Backend
lsof -i :5432  # PostgreSQL
lsof -i :6379  # Redis

# Kill process on port
kill -9 $(lsof -t -i:8080)

# Check Docker logs
docker-compose logs postgres
docker-compose logs redis

# Check backend logs (if running in background)
tail -f backend.log
```

## Contact Points

If you need more help, provide:
1. Backend console output (last 50 lines)
2. Browser console output (all errors)
3. Network tab screenshot (WebSocket section)
4. test-websocket.html results
5. Output of `docker-compose ps`
6. Output of `curl http://localhost:8080/healthz`
