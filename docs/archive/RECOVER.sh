#!/bin/bash

echo "🚑 RECOVERING FRONTEND..."
echo "========================"
echo ""

if [ "$EUID" -ne 0 ]; then 
    echo "❌ Need sudo. Run: sudo ./RECOVER.sh"
    exit 1
fi

echo "1️⃣  Stopping broken frontend..."
docker stop poker-frontend
docker rm poker-frontend
echo "✅ Removed"
echo ""

echo "2️⃣  Rebuilding frontend image..."
cd frontend
docker build -t poker-frontend-new . 2>&1 | tail -20
cd ..
echo "✅ Built"
echo ""

echo "3️⃣  Starting new frontend..."
docker run -d \
  --name poker-frontend \
  --network go-poker-arena---real-time-texas-hold-em-poker-backend-go-fiber-websocket-redis_poker-network \
  -p 3001:3000 \
  -e API_URL=http://backend:8080 \
  -e NEXT_PUBLIC_API_URL=http://localhost:8080 \
  -e NEXT_PUBLIC_WS_URL=ws://localhost:8080 \
  poker-frontend-new

echo "✅ Started"
echo ""

echo "4️⃣  Waiting 15 seconds..."
sleep 15
echo ""

echo "5️⃣  Checking logs..."
docker logs poker-frontend --tail 15
echo ""

echo "========================"
echo "🎉 RECOVERED!"
echo "========================"
echo ""
echo "NOW TEST:"
echo "1. Open INCOGNITO (Ctrl+Shift+N)"
echo "2. Go to: http://localhost:3001"
echo "3. Login and test buttons"
echo ""
