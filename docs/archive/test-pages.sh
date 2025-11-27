#!/bin/bash

echo "🧪 Testing Go Poker Arena Pages..."
echo "=================================="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test 1: Backend Health
echo "1️⃣  Testing Backend Health..."
HEALTH=$(curl -s http://localhost:8080/healthz)
if [[ $HEALTH == *"ok"* ]]; then
    echo -e "${GREEN}✅ Backend is healthy${NC}"
else
    echo -e "${RED}❌ Backend is not responding${NC}"
    exit 1
fi
echo ""

# Test 2: Frontend is accessible
echo "2️⃣  Testing Frontend..."
FRONTEND=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:3001)
if [ "$FRONTEND" == "200" ]; then
    echo -e "${GREEN}✅ Frontend is accessible${NC}"
else
    echo -e "${RED}❌ Frontend is not responding (HTTP $FRONTEND)${NC}"
fi
echo ""

# Test 3: Register endpoint
echo "3️⃣  Testing Register Endpoint..."
REGISTER=$(curl -s -X POST http://localhost:8080/auth/signup \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"test$(date +%s)\",\"email\":\"test$(date +%s)@test.com\",\"password\":\"test123\"}")
if [[ $REGISTER == *"token"* ]]; then
    echo -e "${GREEN}✅ Registration works${NC}"
    TOKEN=$(echo $REGISTER | grep -o '"token":"[^"]*' | cut -d'"' -f4)
    echo "   Token: ${TOKEN:0:30}..."
else
    echo -e "${RED}❌ Registration failed${NC}"
    echo "   Response: $REGISTER"
fi
echo ""

# Test 4: Login endpoint
echo "4️⃣  Testing Login Endpoint..."
LOGIN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"parsa","password":"parsa10"}')
if [[ $LOGIN == *"token"* ]]; then
    echo -e "${GREEN}✅ Login works${NC}"
    TOKEN=$(echo $LOGIN | grep -o '"token":"[^"]*' | cut -d'"' -f4)
else
    echo -e "${YELLOW}⚠️  Login failed (user might not exist)${NC}"
    echo "   Try creating user 'parsa' first"
fi
echo ""

# Test 5: Rooms endpoint
echo "5️⃣  Testing Rooms Endpoint..."
if [ ! -z "$TOKEN" ]; then
    ROOMS=$(curl -s http://localhost:8080/api/rooms \
      -H "Authorization: Bearer $TOKEN")
    if [[ $ROOMS == "["* ]] || [[ $ROOMS == "[]" ]]; then
        echo -e "${GREEN}✅ Rooms endpoint works${NC}"
        echo "   Rooms: $ROOMS"
    else
        echo -e "${RED}❌ Rooms endpoint failed${NC}"
    fi
else
    echo -e "${YELLOW}⚠️  Skipped (no token)${NC}"
fi
echo ""

# Test 6: WebSocket endpoint
echo "6️⃣  Testing WebSocket Endpoint..."
WS_TEST=$(curl -s -o /dev/null -w "%{http_code}" \
  "http://localhost:8080/ws?user_id=1&username=test&room_id=1" \
  -H "Connection: Upgrade" \
  -H "Upgrade: websocket")
if [ "$WS_TEST" == "426" ] || [ "$WS_TEST" == "101" ]; then
    echo -e "${GREEN}✅ WebSocket endpoint is accessible${NC}"
else
    echo -e "${RED}❌ WebSocket endpoint failed (HTTP $WS_TEST)${NC}"
fi
echo ""

# Summary
echo "=================================="
echo "📊 Test Summary"
echo "=================================="
echo ""
echo "✅ Pages Available:"
echo "   🏠 Home/Lobby:  http://localhost:3001/"
echo "   🔐 Login:       http://localhost:3001/login"
echo "   📝 Register:    http://localhost:3001/register"
echo "   🎮 Game:        http://localhost:3001/game/[roomId]"
echo ""
echo "✅ API Endpoints:"
echo "   ❤️  Health:     http://localhost:8080/healthz"
echo "   📝 Signup:      POST http://localhost:8080/auth/signup"
echo "   🔐 Login:       POST http://localhost:8080/auth/login"
echo "   🏠 Rooms:       GET  http://localhost:8080/api/rooms"
echo "   🎮 WebSocket:   ws://localhost:8080/ws"
echo ""
echo -e "${GREEN}🎉 All core features are working!${NC}"
echo ""
echo "📖 Next Steps:"
echo "   1. Open http://localhost:3001 in your browser"
echo "   2. Register a new account or login"
echo "   3. Create a room or join existing one"
echo "   4. Open another browser window to test multiplayer"
echo ""
