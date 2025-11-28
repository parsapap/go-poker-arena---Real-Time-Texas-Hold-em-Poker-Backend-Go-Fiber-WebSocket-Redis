#!/bin/bash

echo "🧪 Quick WebSocket Test"
echo ""

# Get token
echo "1. Getting auth token..."
TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser_1764254810","password":"password123"}' \
  | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
    echo "❌ Login failed"
    exit 1
fi

echo "✅ Logged in"
echo ""

# Get user info
USER_INFO=$(curl -s -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/user)
USER_ID=$(echo "$USER_INFO" | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
USERNAME=$(echo "$USER_INFO" | grep -o '"username":"[^"]*"' | cut -d'"' -f4)

echo "2. User info:"
echo "   ID: $USER_ID"
echo "   Username: $USERNAME"
echo ""

# Get or create room
echo "3. Getting room..."
ROOMS=$(curl -s -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/rooms)
ROOM_ID=$(echo "$ROOMS" | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)

if [ -z "$ROOM_ID" ]; then
    echo "   Creating room..."
    CREATE_RESPONSE=$(curl -s -X POST http://localhost:8080/api/rooms \
      -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json" \
      -d '{"name":"Test Room","max_players":6,"small_blind":10,"big_blind":20}')
    ROOM_ID=$(echo "$CREATE_RESPONSE" | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
fi

echo "✅ Room ID: $ROOM_ID"
echo ""

echo "4. Testing WebSocket connection..."
echo "   URL: ws://localhost:8080/ws?user_id=$USER_ID&username=$USERNAME&room_id=$ROOM_ID"
echo ""

# Use websocat if available, otherwise provide manual test
if command -v websocat &> /dev/null; then
    echo "   Connecting with websocat..."
    (echo '{"type":"join","room_id":"'$ROOM_ID'","user_id":'$USER_ID',"username":"'$USERNAME'"}'; sleep 2) | websocat "ws://localhost:8080/ws?user_id=$USER_ID&username=$USERNAME&room_id=$ROOM_ID" &
    WS_PID=$!
    sleep 3
    kill $WS_PID 2>/dev/null
else
    echo "   ℹ️  Install websocat for automated testing: cargo install websocat"
fi

echo ""
echo "✅ Test complete!"
echo ""
echo "📝 To test in browser:"
echo "   1. Open http://localhost:3000/game/$ROOM_ID"
echo "   2. Check browser console for connection logs"
echo "   3. Look for 'Client joined room' in backend logs"
echo ""
echo "📊 Backend logs:"
echo "   tail -f backend/logs/app.log"
echo "   Or check the process output"
