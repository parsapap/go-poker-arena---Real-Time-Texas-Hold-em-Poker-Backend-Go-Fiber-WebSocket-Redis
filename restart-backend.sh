#!/bin/bash

echo "🔄 Restarting backend..."

# Find and kill the running backend process
PID=$(ps aux | grep "go run cmd/server/main.go" | grep -v grep | awk '{print $2}')

if [ -n "$PID" ]; then
    echo "   Stopping backend (PID: $PID)..."
    kill $PID
    sleep 2
    
    # Force kill if still running
    if ps -p $PID > /dev/null 2>&1; then
        echo "   Force stopping..."
        kill -9 $PID
    fi
    
    echo "   ✅ Backend stopped"
else
    echo "   ℹ️  Backend not running"
fi

echo ""
echo "📝 To start the backend, run:"
echo "   cd backend && go run cmd/server/main.go"
echo ""
echo "   Or in a new terminal:"
echo "   cd backend && make run"
