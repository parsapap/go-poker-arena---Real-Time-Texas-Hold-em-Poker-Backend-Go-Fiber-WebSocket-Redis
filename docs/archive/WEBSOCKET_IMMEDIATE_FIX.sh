#!/bin/bash

echo "================================"
echo "WebSocket Connection Quick Fix"
echo "================================"
echo ""

# Check if backend is running
if ! curl -s http://localhost:8080/healthz > /dev/null 2>&1; then
    echo "❌ Backend is not running!"
    echo "   Start it with: cd backend && go run cmd/server/main.go"
    exit 1
fi

echo "✅ Backend is running"
echo ""

# Get user info from localStorage (you'll need to provide this)
echo "To fix the WebSocket issue, we need to check a few things:"
echo ""
echo "1. Open your browser console (F12)"
echo "2. Go to the Application tab"
echo "3. Look at Local Storage → http://localhost:3000"
echo "4. Find the 'user' item"
echo "5. Copy the user ID from there"
echo ""
echo "Then test the WebSocket connection:"
echo ""
echo "Method 1: Use the test tool"
echo "  open test-websocket.html"
echo ""
echo "Method 2: Test in browser console"
echo "  const ws = new WebSocket('ws://localhost:8080/ws?user_id=YOUR_USER_ID&username=YOUR_USERNAME&room_id=1')"
echo "  ws.onopen = () => console.log('Connected!')"
echo "  ws.onclose = (e) => console.log('Closed:', e.code, e.reason)"
echo "  ws.onmessage = (e) => console.log('Message:', e.data)"
echo ""
echo "Method 3: Check backend logs"
echo "  Look for 'WebSocket connected' messages"
echo "  Look for any error or panic messages"
echo ""

# Create a test WebSocket connection
echo "Testing WebSocket connection..."
echo ""

# Use websocat if available, otherwise provide instructions
if command -v websocat &> /dev/null; then
    echo "Connecting to WebSocket..."
    timeout 5 websocat "ws://localhost:8080/ws?user_id=1&username=test&room_id=1" <<< '{"type":"ping"}' || true
else
    echo "Install websocat for better testing:"
    echo "  brew install websocat  # macOS"
    echo "  cargo install websocat  # Rust"
fi

echo ""
echo "================================"
echo "Next Steps"
echo "================================"
echo ""
echo "1. Check backend console for WebSocket logs"
echo "2. Check frontend console for connection errors"
echo "3. Use test-websocket.html to test connection"
echo "4. Share backend logs if issue persists"
echo ""
