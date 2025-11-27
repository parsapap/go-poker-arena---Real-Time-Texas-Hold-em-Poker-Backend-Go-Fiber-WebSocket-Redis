#!/bin/bash

echo "🎨 Starting Go Poker Arena Frontend..."
echo ""

cd frontend

# Check if node_modules exists
if [ ! -d "node_modules" ]; then
    echo "📦 Installing npm dependencies..."
    npm install
fi

# Create .env.local if it doesn't exist
if [ ! -f ".env.local" ]; then
    echo "📝 Creating .env.local file..."
    cat > .env.local << EOF
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_WS_URL=ws://localhost:8080
NEXT_PUBLIC_FRONTEND_URL=http://localhost:3000
EOF
fi

echo "✅ Starting frontend server on http://localhost:3000"
echo ""
npm run dev
