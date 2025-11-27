# 🔄 How to Rebuild and Fix the Buttons

## The Issue
The "Create Room" and "Quick Play" buttons aren't working because the Next.js proxy configuration needs to be rebuilt.

## ✅ The Fix is Ready
I've already updated `frontend/next.config.js` with the correct proxy configuration. Now you just need to rebuild the frontend container.

---

## 🚀 Option 1: Use the Rebuild Script (Easiest)

I've created a script that does everything for you:

```bash
sudo ./rebuild.sh
```

This will:
1. ✅ Stop all containers
2. ✅ Remove old frontend
3. ✅ Rebuild frontend with fix
4. ✅ Start everything again
5. ✅ Show status

**Time:** About 2-3 minutes

---

## 🚀 Option 2: Manual Commands

If you prefer to run commands manually:

```bash
# 1. Stop containers
sudo docker stop poker-frontend poker-backend poker-postgres poker-redis

# 2. Remove old frontend
sudo docker rm poker-frontend

# 3. Rebuild frontend
cd frontend
sudo docker build -t poker-frontend .
cd ..

# 4. Start database and cache
sudo docker start poker-postgres
sudo docker start poker-redis

# 5. Start backend
sudo docker start poker-backend

# 6. Start frontend with new build
sudo docker run -d \
  --name poker-frontend \
  --network go-poker-arena---real-time-texas-hold-em-poker-backend-go-fiber-websocket-redis_poker-network \
  -p 3001:3000 \
  -e API_URL=http://backend:8080 \
  -e NEXT_PUBLIC_API_URL=http://localhost:8080 \
  -e NEXT_PUBLIC_WS_URL=ws://localhost:8080 \
  poker-frontend

# 7. Check status
sudo docker ps
```

---

## 🚀 Option 3: Docker Compose (If Working)

If your docker-compose works:

```bash
sudo docker-compose down
sudo docker-compose up --build -d
```

---

## ✅ Verify the Fix

After rebuilding, test:

### 1. Check Containers Running
```bash
docker ps

# Should see 4 containers:
# - poker-frontend (port 3001)
# - poker-backend (port 8080)
# - poker-postgres (port 5432)
# - poker-redis (port 6379)
```

### 2. Check Frontend Logs
```bash
docker logs poker-frontend --tail 20

# Should see:
# ✓ Ready in XXXms
# - Local: http://localhost:3000
```

### 3. Test in Browser
1. Open http://localhost:3001
2. Login: `parsa` / `parsa10`
3. Click "Create Room" button
4. ✅ Modal should open!
5. Fill form and submit
6. ✅ Room should appear in list!

### 4. Test Quick Play
1. Click "Quick Play - Auto Matchmaking"
2. ✅ Should show alert: "Joined matchmaking queue!"

---

## 🐛 Troubleshooting

### Issue: Permission Denied
```bash
# Solution: Use sudo
sudo ./rebuild.sh
```

### Issue: Frontend Won't Start
```bash
# Check logs
docker logs poker-frontend

# If build failed, check for errors:
cd frontend
docker build -t poker-frontend . 2>&1 | grep -i error
```

### Issue: Still Getting 404
```bash
# 1. Make sure frontend was rebuilt
docker images | grep poker-frontend

# 2. Check if new config is loaded
docker exec poker-frontend cat /app/next.config.js

# 3. Hard refresh browser
Press: Ctrl + Shift + R
```

### Issue: Containers Won't Stop
```bash
# Force stop
sudo docker kill poker-frontend poker-backend poker-postgres poker-redis

# Then rebuild
sudo ./rebuild.sh
```

---

## 📊 Quick Test After Rebuild

Run this to verify everything works:

```bash
# Get token
TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"parsa","password":"parsa10"}' | \
  grep -o '"token":"[^"]*' | cut -d'"' -f4)

# Test create room through proxy
curl -X POST http://localhost:3001/api/rooms \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Room","max_players":6,"small_blind":10,"big_blind":20}'

# Should return: {"id":1,"name":"Test Room",...}
# If you see this, the fix worked! ✅
```

---

## 🎉 Success Criteria

After rebuild, you should have:

- ✅ All 4 containers running
- ✅ Frontend shows "Ready in XXXms"
- ✅ No 404 errors in backend logs
- ✅ "Create Room" button opens modal
- ✅ Can create rooms successfully
- ✅ "Quick Play" button shows alert
- ✅ Rooms appear in list

---

## 🚀 Ready to Rebuild?

Just run:

```bash
sudo ./rebuild.sh
```

Then test at http://localhost:3001 🎰

---

## 📝 What Changed?

**File:** `frontend/next.config.js`

**Before:**
```javascript
{
  source: '/api/:path*',
  destination: `${apiUrl}/:path*`,  // ❌ Strips /api
}
```

**After:**
```javascript
{
  source: '/api/:path*',
  destination: `${apiUrl}/api/:path*`,  // ✅ Keeps /api
}
```

This ensures the proxy forwards `/api/rooms` to `http://backend:8080/api/rooms` instead of `http://backend:8080/rooms`.

---

**The fix is ready! Just rebuild and test! 🎉**
