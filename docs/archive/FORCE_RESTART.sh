#!/bin/bash

echo "🔥 FORCE RESTARTING EVERYTHING"
echo "=============================="
echo ""

if [ "$EUID" -ne 0 ]; then 
    echo "❌ Need sudo"
    exit 1
fi

echo "Killing all containers..."
docker kill poker-frontend poker-backend poker-postgres poker-redis 2>/dev/null
docker rm poker-frontend 2>/dev/null

echo "Starting database & cache..."
docker start poker-postgres poker-redis
sleep 3

echo "Starting backend..."
docker start poker-backend
sleep 2

echo "Starting frontend (using old working image)..."
docker run -d \
  --name poker-frontend \
  --network go-poker-arena---real-time-texas-hold-em-poker-backend-go-fiber-websocket-redis_poker-network \
  -p 3001:3000 \
  -e API_URL=http://backend:8080 \
  -e NEXT_PUBLIC_API_URL=http://localhost:8080 \
  -e NEXT_PUBLIC_WS_URL=ws://localhost:8080 \
  go-poker-arena---real-time-texas-hold-em-poker-backend-go-fiber-websocket-redis-frontend

sleep 5

echo ""
echo "✅ All containers restarted"
echo ""
echo "Check status:"
docker ps
echo ""
echo "The site should work now at http://localhost:3001"
echo "But buttons still won't work because we need to rebuild with code changes"
echo ""
