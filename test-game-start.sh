#!/bin/bash

echo "🧪 Testing Auto-Start Game Logic"
echo "================================"
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Check backend
echo "1. Checking backend..."
if ! curl -s http://localhost:8080/healthz > /dev/null; then
    echo -e "${RED}✗ Backend is not running${NC}"
    echo "   Start it with: cd backend && go run cmd/server/main.go"
    exit 1
fi
echo -e "${GREEN}✓ Backend is running${NC}"
echo ""

# Create/login two users
echo "2. Creating test users..."

# User 1
USER1_RESPONSE=$(curl -s -X POST http://localhost:8080/auth/signup \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser1","email":"test1@example.com","password":"password123"}')

TOKEN1=$(echo "$USER1_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$TOKEN1" ]; then
    # Try login instead
    USER1_RESPONSE=$(curl -s -X POST http://localhost:8080/auth/login \
      -H "Content-Type: application/json" \
      -d '{"username":"testuser1","password":"password123"}')
    TOKEN1=$(echo "$USER1_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
fi

if [ -z "$TOKEN1" ]; then
    echo -e "${RED}✗ Failed to authenticate user1${NC}"
    exit 1
fi

USER1_ID=$(echo "$USER1_RESPONSE" | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
echo -e "${GREEN}✓ User1 authenticated (ID: $USER1_ID)${NC}"

# User 2
USER2_RESPONSE=$(curl -s -X POST http://localhost:8080/auth/signup \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser2","email":"test2@example.com","password":"password123"}')

TOKEN2=$(echo "$USER2_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$TOKEN2" ]; then
    # Try login instead
    USER2_RESPONSE=$(curl -s -X POST http://localhost:8080/auth/login \
      -H "Content-Type: application/json" \
      -d '{"username":"testuser2","password":"password123"}')
    TOKEN2=$(echo "$USER2_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
fi

if [ -z "$TOKEN2" ]; then
    echo -e "${RED}✗ Failed to authenticate user2${NC}"
    exit 1
fi

USER2_ID=$(echo "$USER2_RESPONSE" | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
echo -e "${GREEN}✓ User2 authenticated (ID: $USER2_ID)${NC}"
echo ""

# Create a room
echo "3. Creating test room..."
CREATE_ROOM=$(curl -s -X POST http://localhost:8080/api/rooms \
  -H "Authorization: Bearer $TOKEN1" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Room","max_players":6,"small_blind":10,"big_blind":20}')

ROOM_ID=$(echo "$CREATE_ROOM" | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)

if [ -z "$ROOM_ID" ]; then
    echo -e "${RED}✗ Failed to create room${NC}"
    echo "   Response: $CREATE_ROOM"
    exit 1
fi

echo -e "${GREEN}✓ Room created (ID: $ROOM_ID)${NC}"
echo ""

# Check room status
echo "4. Checking room status..."
ROOM_STATUS=$(curl -s -H "Authorization: Bearer $TOKEN1" "http://localhost:8080/api/rooms/$ROOM_ID")
echo "   Room: $(echo $ROOM_STATUS | grep -o '"name":"[^"]*"' | cut -d'"' -f4)"
echo "   Status: $(echo $ROOM_STATUS | grep -o '"status":"[^"]*"' | cut -d'"' -f4)"
echo ""

# Test Redis connection
echo "5. Testing Redis player tracking..."
REDIS_KEY="room:${ROOM_ID}:players"
echo "   Redis key: $REDIS_KEY"

# Check initial player count
INITIAL_COUNT=$(redis-cli SCARD "$REDIS_KEY" 2>/dev/null || echo "0")
echo "   Initial player count: $INITIAL_COUNT"
echo ""

# Simulate User1 joining via API
echo "6. User1 joining room..."
JOIN1_RESPONSE=$(curl -s -X POST "http://localhost:8080/api/rooms/$ROOM_ID/join" \
  -H "Authorization: Bearer $TOKEN1")

echo "   Response: $JOIN1_RESPONSE"

# Check player count after user1 joins
sleep 1
COUNT_AFTER_USER1=$(redis-cli SCARD "$REDIS_KEY" 2>/dev/null || echo "0")
echo "   Player count after user1: $COUNT_AFTER_USER1"

if [ "$COUNT_AFTER_USER1" -eq "1" ]; then
    echo -e "${GREEN}✓ User1 joined successfully${NC}"
else
    echo -e "${YELLOW}⚠ Expected 1 player, got $COUNT_AFTER_USER1${NC}"
fi
echo ""

# Simulate User2 joining via API
echo "7. User2 joining room..."
JOIN2_RESPONSE=$(curl -s -X POST "http://localhost:8080/api/rooms/$ROOM_ID/join" \
  -H "Authorization: Bearer $TOKEN2")

echo "   Response: $JOIN2_RESPONSE"

# Check player count after user2 joins
sleep 1
COUNT_AFTER_USER2=$(redis-cli SCARD "$REDIS_KEY" 2>/dev/null || echo "0")
echo "   Player count after user2: $COUNT_AFTER_USER2"

if [ "$COUNT_AFTER_USER2" -eq "2" ]; then
    echo -e "${GREEN}✓ User2 joined successfully${NC}"
else
    echo -e "${YELLOW}⚠ Expected 2 players, got $COUNT_AFTER_USER2${NC}"
fi
echo ""

# Check if game started
echo "8. Checking if game auto-started..."
sleep 4  # Wait for 3-second countdown + 1 second buffer

# Check room status again
ROOM_STATUS_AFTER=$(curl -s -H "Authorization: Bearer $TOKEN1" "http://localhost:8080/api/rooms/$ROOM_ID")
ROOM_STATUS_VALUE=$(echo $ROOM_STATUS_AFTER | grep -o '"status":"[^"]*"' | cut -d'"' -f4)

echo "   Room status: $ROOM_STATUS_VALUE"

if [ "$ROOM_STATUS_VALUE" = "playing" ]; then
    echo -e "${GREEN}✓ Game started automatically!${NC}"
else
    echo -e "${RED}✗ Game did not start (status: $ROOM_STATUS_VALUE)${NC}"
fi
echo ""

# Check Redis for game starting flag
echo "9. Checking Redis flags..."
STARTING_FLAG=$(redis-cli GET "room:${ROOM_ID}:starting" 2>/dev/null || echo "")
if [ -n "$STARTING_FLAG" ]; then
    echo -e "${GREEN}✓ Game starting flag was set${NC}"
else
    echo "   No starting flag found (may have expired)"
fi
echo ""

# List players in Redis
echo "10. Players in room (Redis)..."
PLAYERS=$(redis-cli SMEMBERS "$REDIS_KEY" 2>/dev/null || echo "")
if [ -n "$PLAYERS" ]; then
    echo "$PLAYERS" | while read -r player; do
        echo "   - Player ID: $player"
    done
else
    echo "   No players found in Redis"
fi
echo ""

# Summary
echo "================================"
echo "📊 Test Summary"
echo "================================"
echo "Room ID: $ROOM_ID"
echo "User1 ID: $USER1_ID"
echo "User2 ID: $USER2_ID"
echo "Players in room: $COUNT_AFTER_USER2"
echo "Room status: $ROOM_STATUS_VALUE"
echo ""

if [ "$COUNT_AFTER_USER2" -eq "2" ] && [ "$ROOM_STATUS_VALUE" = "playing" ]; then
    echo -e "${GREEN}✅ ALL TESTS PASSED!${NC}"
    echo ""
    echo "🎮 Test in browser:"
    echo "   1. Open: http://localhost:3000/game/$ROOM_ID"
    echo "   2. Login as testuser1 / password123"
    echo "   3. Open another tab/browser"
    echo "   4. Login as testuser2 / password123"
    echo "   5. Join same room"
    echo "   6. Game should start with countdown!"
else
    echo -e "${RED}❌ TESTS FAILED${NC}"
    echo ""
    echo "🔍 Check backend logs for errors:"
    echo "   Look for messages like:"
    echo "   [ROOM $ROOM_ID] Player X joined → Y players total"
    echo "   [ROOM $ROOM_ID] 2 players → starting game in 3s"
fi
echo ""
