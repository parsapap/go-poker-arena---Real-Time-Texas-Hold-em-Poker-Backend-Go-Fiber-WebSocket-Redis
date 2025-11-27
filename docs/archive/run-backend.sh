#!/bin/bash

echo "🚀 Starting Go Poker Arena Backend..."
echo ""

# Use the correct Go version from snap
export PATH=/snap/bin:$PATH

cd backend

# Check if dependencies are installed
if [ ! -d "vendor" ] && [ ! -f "go.sum" ]; then
    echo "📦 Installing Go dependencies..."
    go mod download
fi

echo "✅ Starting backend server on http://localhost:8080"
echo ""
go run cmd/server/main.go
