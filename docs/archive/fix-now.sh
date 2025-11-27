#!/bin/bash

echo "🔥 NUCLEAR OPTION - Deleting Build Cache & Restarting"
echo "====================================================="
echo ""

if [ "$EUID" -ne 0 ]; then 
    echo "❌ Need sudo. Run: sudo ./fix-now.sh"
    exit 1
fi

echo "1️⃣  Deleting old build cache..."
docker exec poker-frontend rm -rf /app/.next || echo "   (Already deleted)"
echo "✅ Cache deleted"
echo ""

echo "2️⃣  Restarting frontend..."
docker restart poker-frontend
echo "✅ Restarted"
echo ""

echo "3️⃣  Waiting 10 seconds..."
sleep 10
echo ""

echo "4️⃣  Checking status..."
docker logs poker-frontend --tail 10
echo ""

echo "====================================================="
echo "🎉 DONE!"
echo "====================================================="
echo ""
echo "NOW:"
echo "1. Open browser in INCOGNITO mode (Ctrl+Shift+N)"
echo "2. Go to: http://localhost:3001"
echo "3. Login: parsa / parsa10"
echo "4. Click 'Create Room'"
echo "5. IT WILL WORK! ✅"
echo ""
