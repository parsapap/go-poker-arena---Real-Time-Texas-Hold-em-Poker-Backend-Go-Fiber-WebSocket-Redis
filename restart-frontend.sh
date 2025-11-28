#!/bin/bash

echo "🔄 Restarting Frontend with Clean Cache"
echo "========================================"
echo ""

# Find and kill frontend process
echo "1. Stopping frontend..."
PID=$(lsof -ti:3000 2>/dev/null)

if [ -n "$PID" ]; then
    echo "   Killing process on port 3000 (PID: $PID)"
    kill $PID 2>/dev/null
    sleep 2
    
    # Force kill if still running
    if lsof -ti:3000 > /dev/null 2>&1; then
        echo "   Force killing..."
        kill -9 $PID 2>/dev/null
    fi
    echo "   ✅ Frontend stopped"
else
    echo "   ℹ️  No process running on port 3000"
fi

echo ""
echo "2. Clearing Next.js cache..."
rm -rf frontend/.next
rm -rf frontend/node_modules/.cache
echo "   ✅ Cache cleared"

echo ""
echo "3. Starting frontend..."
echo "   Run this command in a new terminal:"
echo ""
echo "   cd frontend && npm run dev"
echo ""
echo "4. After frontend starts:"
echo "   - Open INCOGNITO/PRIVATE window"
echo "   - Navigate to: http://localhost:3000"
echo "   - Login and test"
echo ""
echo "✅ Frontend cache cleared and ready to restart!"
