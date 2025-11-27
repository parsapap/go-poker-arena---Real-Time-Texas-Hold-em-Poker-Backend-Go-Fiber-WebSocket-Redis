# 🔧 Fix: Create Room & Quick Play Buttons Not Working

## 🐛 Problem Identified

The "Create Room" and "Quick Play" buttons are not working because of a **Next.js proxy configuration issue**.

### Root Cause:
- Frontend calls: `/api/rooms`
- Next.js proxy was rewriting to: `/rooms` (removing `/api`)
- Backend expects: `/api/rooms`
- Result: **404 Not Found**

### Evidence:
```bash
Backend logs show:
2025-11-27T12:29:03Z ERR Cannot GET /rooms path=/rooms status=404
2025-11-27T12:29:16Z ERR Cannot POST /rooms path=/rooms status=404
```

---

## ✅ Solution Applied

I've fixed the `frontend/next.config.js` file:

### Before (Broken):
```javascript
{
  source: '/api/:path*',
  destination: `${apiUrl}/:path*`,  // ❌ Strips /api
}
```

### After (Fixed):
```javascript
{
  source: '/api/:path*',
  destination: `${apiUrl}/api/:path*`,  // ✅ Keeps /api
}
```

---

## 🚀 How to Apply the Fix

You need to rebuild the frontend container to apply the changes.

### Option 1: Full Restart (Recommended - 2 minutes)
```bash
# Stop all containers
docker-compose down

# Rebuild and start
docker-compose up --build

# Wait for "Ready in..." message
# Then open http://localhost:3001
```

### Option 2: Rebuild Just Frontend (Faster - 1 minute)
```bash
# Rebuild frontend only
docker-compose build frontend

# Restart frontend
docker-compose up -d frontend

# Check logs
docker logs poker-frontend -f
```

### Option 3: Manual Docker Commands
```bash
# Stop frontend
docker stop poker-frontend
docker rm poker-frontend

# Rebuild
docker build -t poker-frontend ./frontend

# Start with correct config
docker run -d \
  --name poker-frontend \
  --network go-poker-arena---real-time-texas-hold-em-poker-backend-go-fiber-websocket-redis_poker-network \
  -p 3001:3000 \
  -e API_URL=http://backend:8080 \
  -e NEXT_PUBLIC_API_URL=http://localhost:8080 \
  -e NEXT_PUBLIC_WS_URL=ws://localhost:8080 \
  poker-frontend
```

---

## 🧪 Verify the Fix

After rebuilding, test these:

### 1. Check Frontend Logs
```bash
docker logs poker-frontend --tail 20

# Should see:
# ✓ Ready in XXXms
# No errors
```

### 2. Test Create Room Button
1. Open http://localhost:3001
2. Login with: `parsa` / `parsa10`
3. Click "Create Room"
4. Fill form:
   - Name: Test Room
   - Blinds: 10/20
   - Players: 6
5. Click "Create Room"
6. ✅ Should see new room in list!

### 3. Test Quick Play Button
1. Click "Quick Play - Auto Matchmaking"
2. ✅ Should see alert: "Joined matchmaking queue!"

### 4. Check Backend Logs
```bash
docker logs poker-backend --tail 20

# Should see:
# Room created room_id=1 name="Test Room"
# No 404 errors
```

---

## 🎯 Expected Behavior After Fix

### Create Room:
1. Click button → Modal opens ✅
2. Fill form → Submit ✅
3. Modal closes → Room appears in list ✅
4. Backend logs: "Room created" ✅

### Quick Play:
1. Click button → Alert shows ✅
2. Backend logs: "Player joined matchmaking queue" ✅

---

## 🐛 If Still Not Working

### Check 1: Frontend Rebuilt?
```bash
# Check if frontend image was rebuilt
docker images | grep frontend

# Should show recent timestamp
```

### Check 2: Browser Cache
```bash
# Clear browser cache
Press: Ctrl + Shift + R (hard refresh)

# Or open incognito window
```

### Check 3: Backend API Works?
```bash
# Test backend directly
TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"parsa","password":"parsa10"}' | \
  grep -o '"token":"[^"]*' | cut -d'"' -f4)

# Create room via API
curl -X POST http://localhost:8080/api/rooms \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"API Test","max_players":6,"small_blind":10,"big_blind":20}'

# Should return: {"id":1,"name":"API Test",...}
```

### Check 4: Proxy Working?
```bash
# Test through Next.js proxy
curl -X POST http://localhost:3001/api/rooms \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Proxy Test","max_players":6,"small_blind":10,"big_blind":20}'

# Should return: {"id":2,"name":"Proxy Test",...}
```

---

## 📊 Quick Test Script

Run this to verify everything works:

```bash
#!/bin/bash

echo "🧪 Testing Poker Arena After Fix..."

# 1. Login
echo "1. Testing login..."
LOGIN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"parsa","password":"parsa10"}')
TOKEN=$(echo $LOGIN | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
    echo "❌ Login failed"
    exit 1
fi
echo "✅ Login successful"

# 2. Create room via backend
echo "2. Testing create room (backend)..."
ROOM1=$(curl -s -X POST http://localhost:8080/api/rooms \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Backend Test","max_players":6,"small_blind":10,"big_blind":20}')

if [[ $ROOM1 == *"id"* ]]; then
    echo "✅ Backend create room works"
else
    echo "❌ Backend create room failed"
fi

# 3. Create room via proxy
echo "3. Testing create room (proxy)..."
ROOM2=$(curl -s -X POST http://localhost:3001/api/rooms \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Proxy Test","max_players":6,"small_blind":10,"big_blind":20}')

if [[ $ROOM2 == *"id"* ]]; then
    echo "✅ Proxy create room works"
else
    echo "❌ Proxy create room failed"
    echo "   Response: $ROOM2"
fi

# 4. List rooms
echo "4. Testing list rooms..."
ROOMS=$(curl -s http://localhost:3001/api/rooms \
  -H "Authorization: Bearer $TOKEN")

if [[ $ROOMS == "["* ]]; then
    echo "✅ List rooms works"
    echo "   Found rooms: $ROOMS"
else
    echo "❌ List rooms failed"
fi

echo ""
echo "🎉 All tests completed!"
echo ""
echo "Now test in browser:"
echo "1. Open http://localhost:3001"
echo "2. Login"
echo "3. Click 'Create Room' button"
echo "4. Should work! ✅"
```

Save as `test-fix.sh` and run:
```bash
chmod +x test-fix.sh
./test-fix.sh
```

---

## 🎉 Summary

**Issue:** Next.js proxy was stripping `/api` prefix  
**Fix:** Updated `next.config.js` to keep `/api` in destination  
**Action Required:** Rebuild frontend container  
**Time:** 1-2 minutes  
**Result:** Buttons will work! ✅  

---

## 🚀 Quick Fix Command

Just run this:

```bash
docker-compose down && docker-compose up --build -d
```

Then open http://localhost:3001 and test the buttons! 🎰
