#!/bin/bash

echo "🔄 Rebuilding Go Poker Arena..."
echo "================================"
echo ""

# Check if running as root or with sudo
if [ "$EUID" -ne 0 ]; then 
    echo "⚠️  This script needs sudo permissions to stop Docker containers."
    echo "   Please run: sudo ./rebuild.sh"
    echo ""
    exit 1
fi

echo "1️⃣  Stopping containers..."
docker stop poker-frontend poker-backend poker-postgres poker-redis
echo "✅ Containers stopped"
echo ""

echo "2️⃣  Removing old frontend container..."
docker rm poker-frontend
echo "✅ Old container removed"
echo ""

echo "3️⃣  Rebuilding frontend with fix..."
cd frontend
docker build -t poker-frontend . 2>&1 | tail -10
cd ..
echo "✅ Frontend rebuilt"
echo ""

echo "4️⃣  Starting all containers..."
docker start poker-postgres
sleep 2
docker start poker-redis
sleep 2
docker start poker-backend
sleep 3

# Start frontend with correct configuration
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
  poker-frontend

echo "✅ All containers started"
echo ""

echo "5️⃣  Checking status..."
sleep 3
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
echo ""

echo "6️⃣  Checking frontend logs..."
docker logs poker-frontend --tail 10
echo ""

echo "================================"
echo "🎉 Rebuild Complete!"
echo "================================"
echo ""
echo "✅ Frontend has been rebuilt with the proxy fix"
echo "✅ All containers are running"
echo ""
echo "🧪 Test the fix:"
echo "   1. Open http://localhost:3001"
echo "   2. Login with: parsa / parsa10"
echo "   3. Click 'Create Room' button"
echo "   4. Should work now! ✅"
echo ""
echo "📊 Monitor logs:"
echo "   Frontend: docker logs poker-frontend -f"
echo "   Backend:  docker logs poker-backend -f"
echo ""
