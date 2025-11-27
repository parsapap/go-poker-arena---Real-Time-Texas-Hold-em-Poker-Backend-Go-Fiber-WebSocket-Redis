#!/bin/bash

echo "🗑️  COMPLETE CLEANUP AND FRESH START"
echo "====================================="
echo ""

if [ "$EUID" -ne 0 ]; then 
    echo "❌ Need sudo. Run: sudo ./CLEAN_START.sh"
    exit 1
fi

echo "1️⃣  Stopping all containers..."
docker stop poker-frontend poker-backend poker-postgres poker-redis 2>/dev/null
echo "✅ Stopped"
echo ""

echo "2️⃣  Removing all containers..."
docker rm poker-frontend poker-backend poker-postgres poker-redis 2>/dev/null
echo "✅ Removed"
echo ""

echo "3️⃣  Removing all images..."
docker rmi go-poker-arena---real-time-texas-hold-em-poker-backend-go-fiber-websocket-redis-frontend 2>/dev/null
docker rmi go-poker-arena---real-time-texas-hold-em-poker-backend-go-fiber-websocket-redis-backend 2>/dev/null
docker rmi poker-frontend-new 2>/dev/null
echo "✅ Removed"
echo ""

echo "4️⃣  Removing volumes..."
docker volume rm go-poker-arena---real-time-texas-hold-em-poker-backend-go-fiber-websocket-redis_postgres_data 2>/dev/null
docker volume rm go-poker-arena---real-time-texas-hold-em-poker-backend-go-fiber-websocket-redis_redis_data 2>/dev/null
echo "✅ Removed"
echo ""

echo "5️⃣  Removing network..."
docker network rm go-poker-arena---real-time-texas-hold-em-poker-backend-go-fiber-websocket-redis_poker-network 2>/dev/null
echo "✅ Removed"
echo ""

echo "6️⃣  Starting fresh with docker-compose..."
docker-compose up -d --build
echo ""

echo "7️⃣  Waiting 30 seconds for everything to start..."
sleep 30
echo ""

echo "8️⃣  Checking status..."
docker ps
echo ""

echo "====================================="
echo "🎉 FRESH START COMPLETE!"
echo "====================================="
echo ""
echo "Your poker app is now running fresh:"
echo "  Frontend: http://localhost:3001"
echo "  Backend:  http://localhost:8080"
echo ""
echo "Test it now!"
echo ""
