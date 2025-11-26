# 🧪 Testing Guide - Full Game Flow

## Quick Start Testing

### 1. Start the Backend
```bash
cd backend
go run cmd/main.go
```

Expected output:
```
✅ Connected to Redis
✅ Database connected
🚀 Server running on :8080
```

### 2. Start the Frontend
```bash
cd frontend
npm run dev
```

Expected output:
```
✓ Ready in 2.5s
○ Local: http://localhost:3000
```

### 3. Test Full Game Flow

#### Step 1: Register/Login
1. Navigate to `http://localhost:3000`
2. Click "Register" or "Login"
3. Create account or login
4. You should see lobby with your chips

#### Step 2: Create/Join Room
1. Click "Create Room" button
2. Set buy-in amount (e.g., 1000)
3. Set blinds (e.g., 10/20)
4. Click "Create"
5. You'll be redirected to game table

#### Step 3: WebSocket Connection
✅ Check browser console for:
```
✅ WebSocket connected
📨 Received: join
```

✅ Check connection status badge (top center):
- Should show green "Connected" indicator

#### Step 4: Wait for Players
1. Open another browser/incognito window
2. Login with different account
3. Join the same room
4. Both players should see each other

#### Step 5: Game Starts
When 2+ players join:
- ✅ Cards are dealt (you see your 2 hole cards)
- ✅ Blinds are posted
- ✅ Pot shows initial amount
- ✅ First player's turn indicator glows

#### Step 6: Make Actions
When it's your turn:
- ✅ Action bar appears at bottom
- ✅ Fold button (red)
- ✅ Check/Call button (white)
- ✅ Raise slider with presets (2x, 3x, Pot, All-in)
- ✅ Raise button (green)

Test each action:
1. **Fold**: Click fold → you're out of hand
2. **Check**: Click check → action passes
3. **Call**: Click call → matches current bet
4. **Raise**: Adjust slider → click raise → bet increases

#### Step 7: Game Phases
Watch the game progress:
1. **Preflop**: 2 hole cards dealt
2. **Flop**: 3 community cards revealed (animated flip)
3. **Turn**: 4th community card revealed
4. **River**: 5th community card revealed
5. **Showdown**: All active players show cards

#### Step 8: Winner Announcement
When hand ends:
- ✅ Winner modal appears with trophy
- ✅ Confetti animation
- ✅ Chip rain (50 falling chips)
- ✅ Win fanfare sound (if enabled)
- ✅ Toast notification: "🎉 You won $XXX with [hand]!"
- ✅ Chips added to winner's stack

#### Step 9: Chat Testing
1. Type message in chat sidebar (right side)
2. Press Enter or click Send
3. Message appears in chat
4. Other players see your message
5. System messages show joins/leaves/actions

#### Step 10: Sound Effects
Click speaker icon (top-right) to toggle:
- 🎴 Card deal sound (when cards revealed)
- 🪙 Chip stack sound (on bets)
- 🔘 Button click (on actions)
- 🎺 Win fanfare (on win)

#### Step 11: Reconnection Testing
1. Stop backend server
2. Frontend shows "Connection Lost" overlay
3. Reconnection attempts start (2s, 4s, 8s, 10s, 10s)
4. Restart backend
5. Connection restores automatically
6. Game state syncs

#### Step 12: Leave Game
1. Click "Back to Lobby" button
2. WebSocket disconnects cleanly
3. Other players see "[Player] left the table"
4. You return to lobby

## Testing Checklist

### Connection
- [ ] WebSocket connects on page load
- [ ] Green "Connected" badge appears
- [ ] Console shows connection messages
- [ ] Heartbeat keeps connection alive (30s)
- [ ] Reconnects after network drop
- [ ] Shows reconnection overlay
- [ ] Manual retry button works

### Game Flow
- [ ] Join room successfully
- [ ] See other players join
- [ ] Receive 2 hole cards (only yours visible)
- [ ] Other players show card backs
- [ ] Community cards flip in sequence
- [ ] Pot updates on each bet
- [ ] Turn indicator glows for active player
- [ ] Timer counts down (if implemented)
- [ ] Actions process correctly
- [ ] Winner announced properly
- [ ] New hand starts after showdown

### Actions
- [ ] Fold removes you from hand
- [ ] Check passes action (when no bet)
- [ ] Call matches current bet
- [ ] Raise increases bet
- [ ] Slider adjusts raise amount
- [ ] Presets work (2x, 3x, Pot, All-in)
- [ ] Invalid actions rejected
- [ ] Actions disabled when not your turn

### UI/UX
- [ ] Toast notifications appear
- [ ] Toasts auto-dismiss after 5s
- [ ] Chip rain on win
- [ ] Confetti animation
- [ ] Sound effects play
- [ ] Sound toggle works
- [ ] Chat messages send/receive
- [ ] Player chips update
- [ ] Bet amounts display
- [ ] Last action shows (fold, raise, etc)

### Edge Cases
- [ ] Handle 2 players
- [ ] Handle 8 players (max)
- [ ] Player disconnects mid-hand
- [ ] Player reconnects to ongoing game
- [ ] All players fold except one
- [ ] All-in scenarios
- [ ] Side pots (if implemented)
- [ ] Multiple winners (split pot)
- [ ] Player runs out of chips

### Mobile
- [ ] Responsive layout
- [ ] Touch controls work
- [ ] Action buttons accessible
- [ ] Chat hidden on mobile (or collapsible)
- [ ] Cards readable on small screen
- [ ] Swipe to fold (if implemented)

## Common Issues

### "WebSocket connection failed"
**Solution**: 
- Check backend is running: `curl http://localhost:8080/health`
- Verify port 8080 is not blocked
- Check `.env.local` has correct WS_URL

### "No hole cards showing"
**Solution**:
- Check console for WebSocket messages
- Verify you're in the players array
- Check user ID matches
- Ensure game has started (2+ players)

### "Actions not working"
**Solution**:
- Verify it's your turn (glowing border)
- Check myTurn state in console
- Ensure WebSocket is connected
- Check backend logs for errors

### "Sound not playing"
**Solution**:
- Click anywhere on page first (browser policy)
- Check sound toggle is enabled
- Verify browser allows audio
- Check console for audio errors

### "Reconnection not working"
**Solution**:
- Check max attempts not exceeded (5)
- Verify backend is running
- Check network connectivity
- Try manual retry button

## Performance Testing

### Load Test
1. Open 8 browser tabs
2. Login with different accounts
3. All join same room
4. Play multiple hands
5. Monitor:
   - CPU usage (<10%)
   - Memory usage (<100MB per tab)
   - Network latency (<100ms)
   - Frame rate (60fps)

### Stress Test
1. Rapid actions (fold/raise repeatedly)
2. Spam chat messages
3. Disconnect/reconnect rapidly
4. Multiple rooms simultaneously
5. Check for:
   - Memory leaks
   - Connection drops
   - UI freezes
   - Message loss

## Browser Testing

Test in multiple browsers:
- [ ] Chrome (latest)
- [ ] Firefox (latest)
- [ ] Safari (latest)
- [ ] Edge (latest)
- [ ] Mobile Safari (iOS)
- [ ] Chrome Mobile (Android)

## Docker Testing

```bash
# Build and start
docker-compose up --build

# Test
curl http://localhost:3000
curl http://localhost:8080/health

# Check logs
docker-compose logs -f frontend
docker-compose logs -f backend

# Stop
docker-compose down
```

## Production Testing

Before deploying:
- [ ] Update environment variables
- [ ] Test with production URLs
- [ ] Verify SSL/TLS (wss://)
- [ ] Test with real domain
- [ ] Check CORS settings
- [ ] Monitor error rates
- [ ] Load test with real traffic
- [ ] Backup database
- [ ] Test rollback procedure

## Debugging Tools

### Browser Console
```javascript
// Check WebSocket state
console.log(window.WebSocket)

// Monitor store
import { useGameStore } from '@/store/gameStore'
console.log(useGameStore.getState())

// Enable debug logging
localStorage.setItem('debug', 'websocket:*')
```

### Network Tab
- Filter by "WS" to see WebSocket traffic
- Check message payloads
- Monitor connection status
- Verify heartbeat pings

### React DevTools
- Inspect component state
- Check hook values
- Monitor re-renders
- Profile performance

## Success Criteria

✅ **Connection**: Connects in <1s, reconnects automatically
✅ **Latency**: Actions process in <100ms
✅ **Stability**: No disconnects during normal play
✅ **UI**: Smooth 60fps animations
✅ **Sound**: All effects play correctly
✅ **Chat**: Messages deliver instantly
✅ **Game**: Full flow works end-to-end
✅ **Mobile**: Fully responsive and playable

## Next Steps After Testing

1. ✅ Fix any bugs found
2. ✅ Optimize performance bottlenecks
3. ✅ Add missing features
4. ✅ Improve error messages
5. ✅ Add analytics/monitoring
6. ✅ Deploy to staging
7. ✅ User acceptance testing
8. ✅ Deploy to production

---

**Happy Testing!** 🎰🧪

Report issues with:
- Browser/OS version
- Steps to reproduce
- Console errors
- Network logs
- Expected vs actual behavior
