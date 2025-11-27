#!/bin/bash

echo "🔄 Restarting Frontend Container..."
echo "===================================="
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo "⚠️  This script needs sudo permissions."
    echo "   Please run: sudo ./restart-frontend.sh"
    echo ""
    exit 1
fi

echo "1️⃣  Killing old frontend container..."
docker kill poker-frontend 2>/dev/null || echo "   (Container not running)"
echo "✅ Done"
echo ""

echo "2️⃣  Removing old frontend container..."
docker rm poker-frontend 2>/dev/null || echo "   (Container already removed)"
echo "✅ Done"
echo ""

echo "3️⃣  Starting NEW frontend in development mode..."
docker run -d \
  --name poker-frontend \
  --network go-poker-arena---real-time-texas-hold-em-poker-backend-go-fiber-websocket-redis_poker-network \
  -p 3001:3000 \
  -v "$(pwd)/frontend:/app" \
  -w /app \
  -e API_URL=http://backend:8080 \
  -e NEXT_PUBLIC_API_URL=http://localhost:8080 \
  -e NEXT_PUBLIC_WS_URL=ws://localhost:8080 \
  node:20-alpine \
  sh -c 'npm install && npm run dev'

echo "✅ Frontend container started"
echo ""

echo "4️⃣  Waiting for frontend to initialize (60 seconds)..."
for i in {60..1}; do
    echo -ne "   $i seconds remaining...\r"
    sleep 1
done
echo ""
echo "✅ Wait complete"
echo ""

echo "5️⃣  Checking frontend logs..."
docker logs poker-frontend --tail 20
echo ""

echo "===================================="
echo "🎉 Frontend Restart Complete!"
echo "===================================="
echo ""
echo "✅ New frontend is running in development mode"
echo "✅ Proxy configuration is now active"
echo ""
echo "🧪 Test now:"
echo "   1. Open INCOGNITO window (Ctrl+Shift+N)"
echo "   2. Go to: http://localhost:3001"
echo "   3. Login: parsa / parsa10"
echo "   4. Click 'Create Room'"
echo "   5. Should work! ✅"
echo ""
echo "📊 If still not working:"
echo "   - Clear browser cache completely"
echo "   - Try different browser"
echo "   - Check logs: docker logs poker-frontend -f"
echo ""
