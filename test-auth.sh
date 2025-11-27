#!/bin/bash

# Test Authentication Endpoints
# This script tests the backend auth endpoints

API_URL="http://localhost:8080"

echo "================================"
echo "Testing Poker Arena Auth API"
echo "================================"
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test 1: Health Check
echo "1. Testing Health Check..."
HEALTH=$(curl -s -w "\n%{http_code}" "$API_URL/healthz")
HTTP_CODE=$(echo "$HEALTH" | tail -n1)
RESPONSE=$(echo "$HEALTH" | head -n-1)

if [ "$HTTP_CODE" = "200" ]; then
    echo -e "${GREEN}✓ Backend is running${NC}"
    echo "   Response: $RESPONSE"
else
    echo -e "${RED}✗ Backend is not responding${NC}"
    echo "   HTTP Code: $HTTP_CODE"
    echo "   Make sure backend is running: cd backend && go run cmd/server/main.go"
    exit 1
fi
echo ""

# Test 2: Register New User
echo "2. Testing User Registration..."
TIMESTAMP=$(date +%s)
USERNAME="testuser_$TIMESTAMP"
EMAIL="test_${TIMESTAMP}@example.com"
PASSWORD="password123"

REGISTER_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_URL/auth/signup" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USERNAME\",\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")

HTTP_CODE=$(echo "$REGISTER_RESPONSE" | tail -n1)
RESPONSE=$(echo "$REGISTER_RESPONSE" | head -n-1)

if [ "$HTTP_CODE" = "201" ]; then
    echo -e "${GREEN}✓ User registered successfully${NC}"
    echo "   Username: $USERNAME"
    echo "   Email: $EMAIL"
    echo "   Password: $PASSWORD"
    
    # Extract token
    TOKEN=$(echo "$RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    USER_ID=$(echo "$RESPONSE" | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
    
    if [ -n "$TOKEN" ]; then
        echo -e "${GREEN}   Token received: ${TOKEN:0:20}...${NC}"
    fi
else
    echo -e "${RED}✗ Registration failed${NC}"
    echo "   HTTP Code: $HTTP_CODE"
    echo "   Response: $RESPONSE"
    
    # Try with a simple username
    echo ""
    echo "   Trying with simple username 'testuser'..."
    USERNAME="testuser"
    EMAIL="test@example.com"
    
    REGISTER_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_URL/auth/signup" \
      -H "Content-Type: application/json" \
      -d "{\"username\":\"$USERNAME\",\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")
    
    HTTP_CODE=$(echo "$REGISTER_RESPONSE" | tail -n1)
    RESPONSE=$(echo "$REGISTER_RESPONSE" | head -n-1)
    
    if [ "$HTTP_CODE" = "201" ]; then
        echo -e "${GREEN}   ✓ User registered with simple username${NC}"
        TOKEN=$(echo "$RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    elif [ "$HTTP_CODE" = "400" ] && echo "$RESPONSE" | grep -q "already exists"; then
        echo -e "${YELLOW}   ℹ User already exists, will test login${NC}"
    else
        echo -e "${RED}   ✗ Still failed${NC}"
        echo "   Response: $RESPONSE"
    fi
fi
echo ""

# Test 3: Login
echo "3. Testing User Login..."
LOGIN_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}")

HTTP_CODE=$(echo "$LOGIN_RESPONSE" | tail -n1)
RESPONSE=$(echo "$LOGIN_RESPONSE" | head -n-1)

if [ "$HTTP_CODE" = "200" ]; then
    echo -e "${GREEN}✓ Login successful${NC}"
    echo "   Username: $USERNAME"
    
    # Extract user info
    CHIPS=$(echo "$RESPONSE" | grep -o '"chips":[0-9]*' | cut -d':' -f2)
    WINS=$(echo "$RESPONSE" | grep -o '"wins":[0-9]*' | cut -d':' -f2)
    
    echo "   Chips: $CHIPS"
    echo "   Wins: $WINS"
    
    TOKEN=$(echo "$RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    if [ -n "$TOKEN" ]; then
        echo -e "${GREEN}   Token received: ${TOKEN:0:20}...${NC}"
    fi
else
    echo -e "${RED}✗ Login failed${NC}"
    echo "   HTTP Code: $HTTP_CODE"
    echo "   Response: $RESPONSE"
fi
echo ""

# Test 4: Test Protected Endpoint
if [ -n "$TOKEN" ]; then
    echo "4. Testing Protected Endpoint (List Rooms)..."
    ROOMS_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$API_URL/api/rooms" \
      -H "Authorization: Bearer $TOKEN")
    
    HTTP_CODE=$(echo "$ROOMS_RESPONSE" | tail -n1)
    RESPONSE=$(echo "$ROOMS_RESPONSE" | head -n-1)
    
    if [ "$HTTP_CODE" = "200" ]; then
        echo -e "${GREEN}✓ Protected endpoint accessible${NC}"
        ROOM_COUNT=$(echo "$RESPONSE" | grep -o '"id"' | wc -l)
        echo "   Active rooms: $ROOM_COUNT"
    else
        echo -e "${RED}✗ Protected endpoint failed${NC}"
        echo "   HTTP Code: $HTTP_CODE"
        echo "   Response: $RESPONSE"
    fi
    echo ""
fi

# Summary
echo "================================"
echo "Test Summary"
echo "================================"
echo ""
echo "Use these credentials to login:"
echo -e "${YELLOW}Username: $USERNAME${NC}"
echo -e "${YELLOW}Password: $PASSWORD${NC}"
echo ""
echo "Login URL: http://localhost:3000/login"
echo ""

# Check database connection
echo "Checking database..."
if docker ps | grep -q poker-postgres; then
    echo -e "${GREEN}✓ PostgreSQL container is running${NC}"
else
    echo -e "${RED}✗ PostgreSQL container is not running${NC}"
    echo "   Start it with: docker-compose up -d postgres"
fi

if docker ps | grep -q poker-redis; then
    echo -e "${GREEN}✓ Redis container is running${NC}"
else
    echo -e "${RED}✗ Redis container is not running${NC}"
    echo "   Start it with: docker-compose up -d redis"
fi
echo ""
