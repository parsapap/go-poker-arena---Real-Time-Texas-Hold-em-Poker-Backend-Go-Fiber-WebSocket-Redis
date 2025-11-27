#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Go Poker Arena - Local Setup${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo -e "${RED}❌ Docker is not installed${NC}"
    echo "Please install Docker first:"
    echo "  Ubuntu/Debian: sudo apt-get install docker.io docker-compose"
    exit 1
fi

# Check if Docker Compose is installed
if ! command -v docker-compose &> /dev/null; then
    echo -e "${RED}❌ Docker Compose is not installed${NC}"
    echo "Please install Docker Compose first"
    exit 1
fi

echo -e "${GREEN}✓ Docker is installed${NC}"

# Start Docker containers
echo ""
echo -e "${BLUE}Starting PostgreSQL and Redis...${NC}"
docker-compose -f docker-compose.local.yml up -d

# Wait for services to be healthy
echo ""
echo -e "${BLUE}Waiting for services to be ready...${NC}"
sleep 5

# Check PostgreSQL
if docker exec poker-postgres pg_isready -U poker &> /dev/null; then
    echo -e "${GREEN}✓ PostgreSQL is ready${NC}"
else
    echo -e "${RED}❌ PostgreSQL is not ready${NC}"
    exit 1
fi

# Check Redis
if docker exec poker-redis redis-cli ping &> /dev/null; then
    echo -e "${GREEN}✓ Redis is ready${NC}"
else
    echo -e "${RED}❌ Redis is not ready${NC}"
    exit 1
fi

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  Services are ready!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "PostgreSQL: localhost:5432"
echo "Redis:      localhost:6379"
echo ""
echo "Next steps:"
echo ""
echo "1. Start Backend (in a new terminal):"
echo -e "   ${BLUE}cd backend && go run cmd/server/main.go${NC}"
echo ""
echo "2. Start Frontend (in another terminal):"
echo -e "   ${BLUE}cd frontend && npm run dev${NC}"
echo ""
echo "3. Open browser:"
echo -e "   ${BLUE}http://localhost:3000${NC}"
echo ""
echo "To stop Docker services:"
echo -e "   ${BLUE}docker-compose -f docker-compose.local.yml down${NC}"
echo ""
