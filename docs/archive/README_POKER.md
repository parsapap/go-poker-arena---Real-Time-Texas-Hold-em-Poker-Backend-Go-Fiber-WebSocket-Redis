# Texas Hold'em Poker Implementation

## Game Flow

### 1. Starting a Game
```bash
POST /rooms/:id/start
```

### 2. Game Phases
- **Pre-flop**: Each player receives 2 hole cards
- **Flop**: 3 community cards are dealt
- **Turn**: 1 community card is dealt
- **River**: 1 community card is dealt
- **Showdown**: Players reveal hands, winner determined

### 3. Player Actions

#### WebSocket Message Format
```json
{
  "type": "action",
  "room_id": "1",
  "payload": {
    "action": "raise",
    "amount": 100
  }
}
```

#### Available Actions
- `fold` - Fold your hand
- `check` - Check (only if no bet to call)
- `call` - Match the current bet
- `raise` - Raise the bet (must specify amount)
- `allin` - Go all-in with remaining chips

### 4. Hand Rankings (Highest to Lowest)
1. **Royal Flush** - A, K, Q, J, 10 of same suit
2. **Straight Flush** - 5 consecutive cards of same suit
3. **Four of a Kind** - 4 cards of same rank
4. **Full House** - 3 of a kind + pair
5. **Flush** - 5 cards of same suit
6. **Straight** - 5 consecutive cards
7. **Three of a Kind** - 3 cards of same rank
8. **Two Pair** - 2 different pairs
9. **One Pair** - 2 cards of same rank
10. **High Card** - Highest card wins

## WebSocket Events

### Server → Client

#### Game Start
```json
{
  "type": "game_start",
  "room_id": 1,
  "game": {
    "id": 1,
    "phase": "preflop",
    "players": [...],
    "pots": [{"amount": 30, "players": [1,2,3]}]
  }
}
```

#### Deal Cards
```json
{
  "type": "deal",
  "room_id": 1,
  "phase": "flop",
  "community_cards": [
    {"suit": "hearts", "rank": "A"},
    {"suit": "diamonds", "rank": "K"},
    {"suit": "clubs", "rank": "Q"}
  ]
}
```

#### Player Action
```json
{
  "type": "player_action",
  "room_id": 1,
  "player_id": 1,
  "action": "raise",
  "amount": 100,
  "game": {...}
}
```

#### Game End
```json
{
  "type": "game_end",
  "room_id": 1,
  "winners": [
    {
      "player_id": 1,
      "username": "player1",
      "hand": "Full House",
      "cards": [...]
    }
  ]
}
```

## API Endpoints

### Start Game
```bash
POST /rooms/:id/start
Response: Game object with initial state
```

### Process Action
```bash
POST /rooms/:id/action
Body: {
  "player_id": 1,
  "action": "raise",
  "amount": 100
}
```

## Testing

### Run Unit Tests
```bash
go test ./internal/poker/... -v
```

### Run with Coverage
```bash
go test ./internal/poker/... -cover -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Benchmark Hand Evaluator
```bash
go test ./internal/poker/... -bench=. -benchmem
```

## Implementation Details

### Deck Management
- 52 cards (4 suits × 13 ranks)
- Cryptographically secure shuffle using `crypto/rand`
- Draw operations remove cards from deck

### Hand Evaluation
- Uses bitmask operations for fast straight detection
- Evaluates all 21 possible 5-card combinations from 7 cards
- Returns best hand with rank and value for comparison

### Pot Management
- Main pot for all active players
- Side pots created automatically for all-in situations
- Each pot tracks eligible players

### Betting Rounds
- Tracks current bet and minimum raise
- Validates actions based on game state
- Automatically advances to next phase when betting complete

## Example Game Flow

1. **Create Room**
```bash
POST /rooms
{
  "name": "High Stakes",
  "max_players": 9,
  "small_blind": 10,
  "big_blind": 20
}
```

2. **Players Join via WebSocket**
```
ws://localhost:8080/ws?user_id=1&username=player1&room_id=1
```

3. **Start Game**
```bash
POST /rooms/1/start
```

4. **Players Receive Hole Cards** (via WebSocket)

5. **Betting Round** (Pre-flop)
- Player 1: Call 20
- Player 2: Raise to 60
- Player 3: Fold
- Player 1: Call 40

6. **Flop Dealt** (3 community cards)

7. **Betting Round**
- Player 1: Check
- Player 2: Bet 100
- Player 1: Call 100

8. **Turn Dealt** (1 community card)

9. **Betting Round**
- Player 1: Check
- Player 2: Bet 200
- Player 1: Raise to 500
- Player 2: Call 300

10. **River Dealt** (1 community card)

11. **Final Betting Round**
- Player 1: All-in 1000
- Player 2: Call 1000

12. **Showdown**
- Hands evaluated
- Winner determined
- Chips distributed
