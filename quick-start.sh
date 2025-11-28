#!/bin/bash

# Go Poker Arena - Quick Start Script
# This script sets up and runs the entire application with one command

set -e

echo "🎰 Go Poker Arena - Quick Start"
echo "================================"
echo ""

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker first:"
    echo "   https://docs.docker.com/get-docker/"
    exit 1
fi

# Check if Docker Compose is installed
if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose is not installed. Please install Docker Compose first:"
    echo "   https://docs.docker.com/compose/install/"
    exit 1
fi

echo "✅ Docker and Docker Compose are installed"
echo ""

# Create .env file if it doesn't exist
if [ ! -f .env ]; then
    echo "📝 Creating .env file from .env.example..."
    cp .env.example .env
    echo "✅ .env file created"
else
    echo "✅ .env file already exists"
fi

echo ""
echo "🚀 Starting Go Poker Arena..."
echo ""

# Start services
docker-compose -f docker-compose.prod.yml up -d

echo ""
echo "⏳ Waiting for services to be ready..."
sleep 10

# Check if services are running
if docker-compose -f docker-compose.prod.yml ps | grep -q "Up"; then
    echo ""
    echo "✅ All services are running!"
    echo ""
    echo "🎉 Go Poker Arena is ready!"
    echo ""
    echo "📍 Access the application:"
    echo "   Frontend:  http://localhost:3000"
    echo "   Backend:   http://localhost:8080"
    echo "   API Docs:  http://localhost:8080/healthz"
    echo ""
    echo "🎮 Quick Actions:"
    echo "   • Register a new account"
    echo "   • Create or join a poker room"
    echo "   • Start playing!"
    echo ""
    echo "📊 View logs:"
    echo "   docker-compose -f docker-compose.prod.yml logs -f"
    echo ""
    echo "🛑 Stop services:"
    echo "   docker-compose -f docker-compose.prod.yml down"
    echo ""
    echo "🔄 Restart services:"
    echo "   docker-compose -f docker-compose.prod.yml restart"
    echo ""
else
    echo ""
    echo "❌ Some services failed to start. Check logs:"
    echo "   docker-compose -f docker-compose.prod.yml logs"
    exit 1
fi
