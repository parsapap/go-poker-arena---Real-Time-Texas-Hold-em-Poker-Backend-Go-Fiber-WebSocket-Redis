#!/bin/bash

echo "🔄 RESTORING FRONTEND FROM BACKUP"
echo "=================================="
echo ""

if [ "$EUID" -ne 0 ]; then 
    echo "❌ Need sudo"
    exit 1
fi

echo "1️⃣  Stopping broken frontend..."
docker stop poker-frontend
docker rm poker-frontend

echo "2️⃣  Starting fresh frontend from original image..."
docker run -d \
  --name poker-frontend \
  --network go-poker-arena---real-time-texas-hold-em-poker-backend-go-fiber-websocket-redis_poker-network \
  -p 3001:3000 \
  -e API_URL=http://backend:8080 \
  -e NEXT_PUBLIC_API_URL=http://localhost:8080 \
  -e NEXT_PUBLIC_WS_URL=ws://localhost:8080 \
  -e NEXT_PUBLIC_FRONTEND_URL=http://localhost:3001 \
  -e PORT=3000 \
  -e HOSTNAME="0.0.0.0" \
  go-poker-arena---real-time-texas-hold-em-poker-backend-go-fiber-websocket-redis-frontend

echo "3️⃣  Waiting 10 seconds..."
sleep 10

echo "4️⃣  Checking status..."
docker ps | grep frontend
echo ""
docker logs poker-frontend --tail 10

echo ""
echo "=================================="
echo "✅ FRONTEND RESTORED!"
echo "=================================="
echo ""
echo "Test now: http://localhost:3001"
echo ""
echo "The site should work now!"
echo "(But Create Room button still won't work - use API)"
echo ""
