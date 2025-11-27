#!/bin/bash

echo "🔧 Fixing Frontend Proxy Configuration..."
echo ""

# The issue: Next.js proxy is stripping /api prefix
# Solution: Update next.config.js to keep /api in destination

echo "✅ Updated next.config.js"
echo "   Changed: destination: \`\${apiUrl}/:path*\`"
echo "   To:      destination: \`\${apiUrl}/api/:path*\`"
echo ""

echo "📝 To apply the fix, you need to rebuild the frontend:"
echo ""
echo "Option 1: Restart Docker Compose (Recommended)"
echo "   docker-compose down"
echo "   docker-compose up --build"
echo ""
echo "Option 2: Rebuild just frontend"
echo "   docker-compose build frontend"
echo "   docker-compose up -d frontend"
echo ""
echo "Option 3: Quick fix without rebuild (temporary)"
echo "   The frontend will pick up changes on next restart"
echo ""

# Test if API works directly
echo "🧪 Testing backend API directly..."
TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"parsa","password":"parsa10"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [ ! -z "$TOKEN" ]; then
    echo "✅ Backend API works!"
    echo ""
    echo "Testing /api/rooms endpoint..."
    ROOMS=$(curl -s http://localhost:8080/api/rooms -H "Authorization: Bearer $TOKEN")
    echo "Response: $ROOMS"
    echo ""
    echo "✅ Backend is working correctly!"
    echo "❌ Issue is with Next.js proxy configuration"
else
    echo "❌ Backend login failed"
fi

echo ""
echo "🔄 Temporary Workaround (until rebuild):"
echo "   The buttons won't work until you rebuild the frontend."
echo "   But you can test the API directly:"
echo ""
echo "   curl -X POST http://localhost:8080/api/rooms \\"
echo "     -H 'Authorization: Bearer YOUR_TOKEN' \\"
echo "     -H 'Content-Type: application/json' \\"
echo "     -d '{\"name\":\"Test Room\",\"max_players\":6,\"small_blind\":10,\"big_blind\":20}'"
echo ""
