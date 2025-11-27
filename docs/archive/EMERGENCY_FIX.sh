#!/bin/bash

echo "🚨 EMERGENCY FIX - Restarting Frontend in Development Mode"
echo "=========================================================="
echo ""

# This will run the frontend in development mode which picks up config changes immediately

echo "1️⃣  Stopping old frontend..."
sudo docker stop poker-frontend
sudo docker rm poker-frontend
echo "✅ Old frontend removed"
echo ""

echo "2️⃣  Starting frontend in DEVELOPMENT mode..."
echo "   (This mode picks up config changes without rebuild)"
echo ""

sudo docker run -d \
  --name poker-frontend \
  --network go-poker-arena---real-time-texas-hold-em-poker-backend-go-fiber-websocket-redis_poker-network \
  -p 3001:3000 \
  -v "$(pwd)/frontend:/app" \
  -w /app \
  -e API_URL=http://backend:8080 \
  -e NEXT_PUBLIC_API_URL=http://localhost:8080 \
  -e NEXT_PUBLIC_WS_URL=ws://localhost:8080 \
  node:20-alpine \
  sh -c "npm install && npm run dev"

echo "✅ Frontend starting in development mode..."
echo ""

echo "3️⃣  Waiting for frontend to start (30 seconds)..."
sleep 30

echo ""
echo "4️⃣  Checking logs..."
sudo docker logs poker-frontend --tail 15
echo ""

echo "=========================================================="
echo "🎉 Frontend is now running in DEVELOPMENT mode!"
echo "=========================================================="
echo ""
echo "✅ Config changes are now active"
echo "✅ No rebuild needed for future changes"
echo ""
echo "🧪 Test now:"
echo "   1. Open http://localhost:3001"
echo "   2. Press Ctrl+Shift+R (hard refresh)"
echo "   3. Login: parsa / parsa10"
echo "   4. Click 'Create Room' - Should work! ✅"
echo ""
echo "📊 Monitor logs:"
echo "   sudo docker logs poker-frontend -f"
echo ""
