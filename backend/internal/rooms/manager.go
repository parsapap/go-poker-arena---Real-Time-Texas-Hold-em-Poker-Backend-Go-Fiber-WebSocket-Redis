package rooms

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
	"go-poker-arena/internal/models"
	"go-poker-arena/internal/poker"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Manager struct {
	DB    *gorm.DB
	Redis *redis.Client
	// Games holds the in-memory live game state keyed by room ID.
	// It is accessed concurrently from HTTP handlers and background
	// goroutines (e.g. StartGameCountdown), so every access MUST go
	// through the setGame/getGame/deleteGame helpers, which guard it
	// with mu.
	Games map[uint]*poker.Game
	// mu guards the Games map. We use an RWMutex so that frequent
	// reads (getGame) can proceed in parallel, while writes
	// (setGame/deleteGame) take an exclusive lock.
	mu sync.RWMutex
}

// roomCacheTTL bounds how long a cached room snapshot may be stale. Room
// metadata is also invalidated explicitly on update (see cacheRoom /
// invalidateRoomCache), but a TTL provides a safety net so cache entries can
// never leak or remain stale indefinitely.
const roomCacheTTL = 10 * time.Minute

func NewManager(db *gorm.DB, redisClient *redis.Client) *Manager {
	return &Manager{
		DB:    db,
		Redis: redisClient,
		Games: make(map[uint]*poker.Game),
	}
}

// setGame stores a game in the Games map under an exclusive write lock.
func (m *Manager) setGame(roomID uint, game *poker.Game) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Games[roomID] = game
}

// getGame retrieves a game from the Games map under a shared read lock,
// allowing concurrent reads without blocking each other.
func (m *Manager) getGame(roomID uint) (*poker.Game, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	game, ok := m.Games[roomID]
	return game, ok
}

// deleteGame removes a game from the Games map under an exclusive write lock.
func (m *Manager) deleteGame(roomID uint) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.Games, roomID)
}

// roomCacheKey returns the Redis key holding the cached room snapshot.
func roomCacheKey(roomID uint) string {
	return fmt.Sprintf("room:%d", roomID)
}

// cacheRoom writes the room snapshot to Redis with a bounded TTL.
func (m *Manager) cacheRoom(ctx context.Context, room *models.Room) {
	roomData, err := json.Marshal(room)
	if err != nil {
		return
	}
	m.Redis.Set(ctx, roomCacheKey(room.ID), roomData, roomCacheTTL)
}

// invalidateRoomCache removes a stale room snapshot so the next read
// repopulates it from the database. Call this whenever room state changes.
func (m *Manager) invalidateRoomCache(ctx context.Context, roomID uint) {
	m.Redis.Del(ctx, roomCacheKey(roomID))
}

// RoomConfig holds the parameters needed to create a room. It is validated by
// ValidateRoomConfig before a room is persisted.
type RoomConfig struct {
	Name       string
	MaxPlayers int
	SmallBlind int64
	BigBlind   int64
}

func (m *Manager) CreateRoom(name string, maxPlayers int, smallBlind, bigBlind int64) (*models.Room, error) {
	cfg := RoomConfig{
		Name:       name,
		MaxPlayers: maxPlayers,
		SmallBlind: smallBlind,
		BigBlind:   bigBlind,
	}
	if err := ValidateRoomConfig(cfg); err != nil {
		return nil, err
	}

	room := &models.Room{
		Name:       cfg.Name,
		MaxPlayers: cfg.MaxPlayers,
		SmallBlind: cfg.SmallBlind,
		BigBlind:   cfg.BigBlind,
		Status:     "waiting",
	}

	if err := m.DB.Create(room).Error; err != nil {
		return nil, err
	}

	ctx := context.Background()
	m.cacheRoom(ctx, room)

	return room, nil
}

// maxAllowedBlind caps blind sizes to a sane upper bound so a typo can't create
// a table with absurd stakes.
const maxAllowedBlind = int64(1_000_000)

// ErrInvalidRoomConfig wraps all room-validation failures so callers (e.g. the
// HTTP handler) can distinguish a client input error from an internal error
// and respond with the appropriate status code.
var ErrInvalidRoomConfig = errors.New("invalid room configuration")

// ValidateRoomConfig enforces sane room parameters at creation time so that
// invalid games (e.g. a single-seat table, non-positive blinds, or a big blind
// not larger than the small blind) can never be persisted. All failures wrap
// ErrInvalidRoomConfig.
func ValidateRoomConfig(cfg RoomConfig) error {
	switch {
	case strings.TrimSpace(cfg.Name) == "":
		return fmt.Errorf("%w: room name is required", ErrInvalidRoomConfig)
	case cfg.MaxPlayers < 2 || cfg.MaxPlayers > 9:
		return fmt.Errorf("%w: max_players must be between 2 and 9", ErrInvalidRoomConfig)
	case cfg.SmallBlind <= 0:
		return fmt.Errorf("%w: small_blind must be positive", ErrInvalidRoomConfig)
	case cfg.BigBlind <= cfg.SmallBlind:
		return fmt.Errorf("%w: big_blind must be greater than small_blind", ErrInvalidRoomConfig)
	case cfg.BigBlind > maxAllowedBlind:
		return fmt.Errorf("%w: big_blind must not exceed %d", ErrInvalidRoomConfig, maxAllowedBlind)
	}
	return nil
}

func (m *Manager) GetRoom(roomID uint) (*models.Room, error) {
	ctx := context.Background()
	key := roomCacheKey(roomID)

	val, err := m.Redis.Get(ctx, key).Result()
	if err == nil {
		var room models.Room
		if err := json.Unmarshal([]byte(val), &room); err == nil {
			return &room, nil
		}
	}

	var room models.Room
	if err := m.DB.First(&room, roomID).Error; err != nil {
		return nil, err
	}

	m.cacheRoom(ctx, &room)

	return &room, nil
}

func (m *Manager) ListRooms() ([]models.Room, error) {
	var rooms []models.Room
	if err := m.DB.Where("status != ?", "finished").Find(&rooms).Error; err != nil {
		return nil, err
	}
	return rooms, nil
}

func (m *Manager) JoinRoom(roomID, userID uint) error {
	ctx := context.Background()
	key := fmt.Sprintf("room:%d:players", roomID)

	room, err := m.GetRoom(roomID)
	if err != nil {
		return err
	}

	// Add the player first, then read the authoritative member count. Reading
	// the count BEFORE adding caused a time-of-check/time-of-use race: two
	// near-simultaneous joins both saw 0 and neither reached the 2-player
	// start threshold. SADD returns whether the member was newly added.
	added, err := m.Redis.SAdd(ctx, key, userID).Result()
	if err != nil {
		return err
	}

	newCount, err := m.Redis.SCard(ctx, key).Result()
	if err != nil {
		return err
	}

	// Enforce capacity using the post-add count. If this player pushed the
	// room over its max, roll back the add and reject.
	if int(newCount) > room.MaxPlayers {
		m.Redis.SRem(ctx, key, userID)
		return fmt.Errorf("room is full")
	}

	// Check if we have enough players to start the game
	if added == 1 {
		fmt.Printf("[ROOM %d] Player %d joined → %d players total\n", roomID, userID, newCount)
	}
	
	if newCount >= 2 && room.Status == "waiting" {
		// Check if game is not already starting
		startingKey := fmt.Sprintf("room:%d:starting", roomID)
		isStarting, _ := m.Redis.Exists(ctx, startingKey).Result()
		
		if isStarting == 0 {
			// Mark as starting to prevent duplicate starts (5 seconds)
			m.Redis.SetEx(ctx, startingKey, "1", 5*time.Second)
			
			fmt.Printf("[ROOM %d] %d players → starting game in 3s\n", roomID, newCount)
			
			// Start countdown in goroutine
			go m.StartGameCountdown(roomID)
		}
	}

	// Player membership changed; drop the cached snapshot so any derived
	// reads repopulate from the source of truth.
	m.invalidateRoomCache(ctx, roomID)

	return nil
}

func (m *Manager) LeaveRoom(roomID, userID uint) error {
	ctx := context.Background()
	key := fmt.Sprintf("room:%d:players", roomID)
	if err := m.Redis.SRem(ctx, key, userID).Err(); err != nil {
		return err
	}
	// Player membership changed; invalidate the cached room snapshot.
	m.invalidateRoomCache(ctx, roomID)
	return nil
}

func (m *Manager) GetRoomPlayers(roomID uint) ([]uint, error) {
	ctx := context.Background()
	key := fmt.Sprintf("room:%d:players", roomID)
	
	members, err := m.Redis.SMembers(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var playerIDs []uint
	for _, member := range members {
		var id uint
		fmt.Sscanf(member, "%d", &id)
		playerIDs = append(playerIDs, id)
	}

	return playerIDs, nil
}

func (m *Manager) StartGameCountdown(roomID uint) {
	ctx := context.Background()
	
	// Countdown: 3, 2, 1
	for i := 3; i > 0; i-- {
		countdownData, _ := json.Marshal(map[string]interface{}{
			"type":      "gameStarting",
			"room_id":   fmt.Sprintf("%d", roomID), // String for WebSocket hub
			"countdown": i,
			"message":   fmt.Sprintf("Game starting in %d...", i),
		})
		m.Redis.Publish(ctx, fmt.Sprintf("room:%d", roomID), countdownData)
		fmt.Printf("[ROOM %d] Countdown: %d\n", roomID, i)
		
		time.Sleep(1 * time.Second)
	}
	
	// Start the game
	fmt.Printf("[ROOM %d] Starting game NOW!\n", roomID)
	room, err := m.StartGame(roomID)
	if err != nil {
		fmt.Printf("[ROOM %d] Failed to start game: %v\n", roomID, err)
		// Broadcast error
		errorData, _ := json.Marshal(map[string]interface{}{
			"type":    "error",
			"room_id": fmt.Sprintf("%d", roomID),
			"message": fmt.Sprintf("Failed to start game: %v", err),
		})
		m.Redis.Publish(ctx, fmt.Sprintf("room:%d", roomID), errorData)
	} else {
		fmt.Printf("[ROOM %d] Game started successfully! Status: %s\n", roomID, room.Status)
	}
}

func (m *Manager) StartGame(roomID uint) (*models.Room, error) {
	room, err := m.GetRoom(roomID)
	if err != nil {
		return nil, err
	}

	playerIDs, err := m.GetRoomPlayers(roomID)
	if err != nil {
		return nil, err
	}

	if len(playerIDs) < 2 {
		return nil, fmt.Errorf("need at least 2 players to start game")
	}

	// Fetch player data
	var users []models.User
	if err := m.DB.Where("id IN ?", playerIDs).Find(&users).Error; err != nil {
		return nil, err
	}

	// Convert to poker players
	players := make([]*poker.Player, len(users))
	for i, user := range users {
		players[i] = &poker.Player{
			ID:       user.ID,
			Username: user.Username,
			Chips:    user.Chips,
			Position: i,
			Active:   true,
		}
	}

	game := poker.NewGame(roomID, players, room.SmallBlind, room.BigBlind)
	if err := game.Start(); err != nil {
		return nil, err
	}

	m.setGame(roomID, game)

	ctx := context.Background()

	// Update room status
	room.Status = "playing"
	m.DB.Model(room).Update("status", "playing")
	// Invalidate the cached snapshot so the next read reflects "playing".
	m.invalidateRoomCache(ctx, roomID)

	// Save game to database
	dbGame := &models.Game{
		RoomID: roomID,
		Status: "active",
		Pot:    game.Pots[0].Amount,
		Stage:  string(game.Phase),
	}
	if err := m.DB.Create(dbGame).Error; err != nil {
		return nil, err
	}
	game.ID = dbGame.ID

	roomIDStr := fmt.Sprintf("%d", roomID)
	
	// 1. Send hole cards to each player privately
	for _, player := range game.Players {
		dealData, _ := json.Marshal(map[string]interface{}{
			"type":            "deal",
			"room_id":         roomIDStr,
			"phase":           game.Phase,
			"community_cards": game.CommunityCards,
			"player_id":       player.ID,
			"hole_cards":      player.HoleCards,
		})
		m.Redis.Publish(ctx, fmt.Sprintf("room:%d", roomID), dealData)
		fmt.Printf("[ROOM %d] Sent deal to player %d with %d cards\n", roomID, player.ID, len(player.HoleCards))
	}

	// 2. Broadcast phase change
	phaseData, _ := json.Marshal(map[string]interface{}{
		"type":    "phaseChange",
		"room_id": roomIDStr,
		"phase":   game.Phase,
	})
	m.Redis.Publish(ctx, fmt.Sprintf("room:%d", roomID), phaseData)
	fmt.Printf("[ROOM %d] Broadcast: phaseChange (phase=%s)\n", roomID, game.Phase)

	// 3. Broadcast pot update
	potData, _ := json.Marshal(map[string]interface{}{
		"type":        "potUpdate",
		"room_id":     roomIDStr,
		"pot":         game.Pots[0].Amount,
		"current_bet": game.CurrentBet,
	})
	m.Redis.Publish(ctx, fmt.Sprintf("room:%d", roomID), potData)
	fmt.Printf("[ROOM %d] Broadcast: potUpdate (pot=%d, current_bet=%d)\n", roomID, game.Pots[0].Amount, game.CurrentBet)

	// 4. Broadcast player turn
	currentPlayer := game.Players[game.CurrentPosition]
	turnData, _ := json.Marshal(map[string]interface{}{
		"type":      "playerTurn",
		"room_id":   roomIDStr,
		"player_id": currentPlayer.ID,
		"username":  currentPlayer.Username,
		"position":  game.CurrentPosition,
	})
	m.Redis.Publish(ctx, fmt.Sprintf("room:%d", roomID), turnData)
	fmt.Printf("[ROOM %d] Broadcast: playerTurn (player=%s, position=%d)\n", roomID, currentPlayer.Username, game.CurrentPosition)

	// 5. Broadcast full game state
	gameStateData, _ := json.Marshal(map[string]interface{}{
		"type":    "gameState",
		"room_id": roomIDStr,
		"game":    m.serializeGameState(game, 0),
	})
	m.Redis.Publish(ctx, fmt.Sprintf("room:%d", roomID), gameStateData)
	fmt.Printf("[ROOM %d] Broadcast: gameState\n", roomID)

	return room, nil
}

func (m *Manager) GetGame(roomID uint) (interface{}, error) {
	game, ok := m.getGame(roomID)
	if !ok {
		return nil, fmt.Errorf("game not found for room %d", roomID)
	}
	return game, nil
}

func (m *Manager) ProcessAction(roomID, playerID uint, action string, amount int64) error {
	gameInterface, err := m.GetGame(roomID)
	if err != nil {
		return err
	}

	game, ok := gameInterface.(*poker.Game)
	if !ok {
		return fmt.Errorf("invalid game type")
	}

	if err := game.ProcessAction(playerID, poker.Action(action), amount); err != nil {
		return err
	}

	ctx := context.Background()
	roomIDStr := fmt.Sprintf("%d", roomID)

	// 1. Publish action event
	actionData, _ := json.Marshal(map[string]interface{}{
		"type":      "player_action",
		"room_id":   roomIDStr,
		"player_id": playerID,
		"action":    action,
		"amount":    amount,
		"game":      m.serializeGameState(game, 0),
	})
	m.Redis.Publish(ctx, fmt.Sprintf("room:%d", roomID), actionData)

	// 2. Check if only one player remains (everyone else folded)
	activePlayers := 0
	var lastActivePlayer *poker.Player
	for _, p := range game.Players {
		if !p.Folded {
			activePlayers++
			lastActivePlayer = p
		}
	}
	
	if activePlayers == 1 && lastActivePlayer != nil {
		// One player wins by fold. Delegate to the engine so the pot is
		// awarded and game.Winners is populated authoritatively (instead of
		// crediting chips here and recomputing winners later).
		if err := game.AwardToLastPlayer(); err != nil {
			return err
		}

		fmt.Printf("[ROOM %d] %s wins by fold!\n", roomID, lastActivePlayer.Username)
		
		// End the round
		m.EndRound(roomID)
		return nil
	}

	// 3. Check if game ended (showdown or finished)
	if game.Phase == poker.PhaseShowdown || game.Phase == poker.PhaseFinished {
		// Broadcast showdown with all players' cards
		showdownData, _ := json.Marshal(map[string]interface{}{
			"type":            "showdown",
			"room_id":         roomIDStr,
			"community_cards": game.CommunityCards,
			"players":         m.serializePlayersForShowdown(game),
		})
		m.Redis.Publish(ctx, fmt.Sprintf("room:%d", roomID), showdownData)
		fmt.Printf("[ROOM %d] Showdown! Broadcasting all cards\n", roomID)

		// End the round and broadcast winner
		m.EndRound(roomID)
		return nil
	}

	// 3. Broadcast whose turn it is next (if game is still active)
	currentPlayer := game.Players[game.CurrentPosition]
	turnData, _ := json.Marshal(map[string]interface{}{
		"type":      "playerTurn",
		"room_id":   roomIDStr,
		"player_id": currentPlayer.ID,
		"username":  currentPlayer.Username,
		"position":  game.CurrentPosition,
	})
	m.Redis.Publish(ctx, fmt.Sprintf("room:%d", roomID), turnData)
	fmt.Printf("[ROOM %d] Turn switched to player %s (ID: %d, position: %d)\n", roomID, currentPlayer.Username, currentPlayer.ID, game.CurrentPosition)

	return nil
}

func (m *Manager) DealCards(roomID uint) error {
	gameInterface, err := m.GetGame(roomID)
	if err != nil {
		return err
	}

	game, ok := gameInterface.(*poker.Game)
	if !ok {
		return fmt.Errorf("invalid game type")
	}

	ctx := context.Background()
	dealData, _ := json.Marshal(map[string]interface{}{
		"type":            "deal",
		"room_id":         fmt.Sprintf("%d", roomID),
		"phase":           game.Phase,
		"community_cards": game.CommunityCards,
	})
	m.Redis.Publish(ctx, fmt.Sprintf("room:%d", roomID), dealData)

	return nil
}

func (m *Manager) EndRound(roomID uint) error {
	gameInterface, err := m.GetGame(roomID)
	if err != nil {
		return err
	}

	game, ok := gameInterface.(*poker.Game)
	if !ok {
		return fmt.Errorf("invalid game type")
	}

	// Update player chips in database
	for _, player := range game.Players {
		m.DB.Model(&models.User{}).Where("id = ?", player.ID).Update("chips", player.Chips)
	}

	// Update game status
	m.DB.Model(&models.Game{}).Where("id = ?", game.ID).Updates(map[string]interface{}{
		"status": "finished",
		"stage":  string(poker.PhaseFinished),
	})

	// Mark the room as finished and invalidate its cached snapshot.
	ctx := context.Background()
	m.DB.Model(&models.Room{}).Where("id = ?", roomID).Update("status", "finished")
	m.invalidateRoomCache(ctx, roomID)

	// Publish game end event
	endData, _ := json.Marshal(map[string]interface{}{
		"type":    "game_end",
		"room_id": fmt.Sprintf("%d", roomID),
		"winners": m.getWinners(game),
	})
	m.Redis.Publish(ctx, fmt.Sprintf("room:%d", roomID), endData)

	m.deleteGame(roomID)
	return nil
}

func (m *Manager) serializeGameState(game *poker.Game, forPlayerID uint) map[string]interface{} {
	players := make([]map[string]interface{}, len(game.Players))
	for i, p := range game.Players {
		player := map[string]interface{}{
			"id":        p.ID,
			"username":  p.Username,
			"chips":     p.Chips,
			"bet":       p.Bet,
			"position":  p.Position,
			"folded":    p.Folded,
			"all_in":    p.AllIn,
		}
		// Only show hole cards to the player themselves
		if p.ID == forPlayerID || game.Phase == poker.PhaseShowdown {
			player["hole_cards"] = p.HoleCards
		}
		players[i] = player
	}

	return map[string]interface{}{
		"id":               game.ID,
		"room_id":          game.RoomID,
		"phase":            game.Phase,
		"community_cards":  game.CommunityCards,
		"players":          players,
		"pots":             game.Pots,
		"current_bet":      game.CurrentBet,
		"min_raise":        game.MinRaise,
		"current_position": game.CurrentPosition,
	}
}

func (m *Manager) serializePlayersForShowdown(game *poker.Game) []map[string]interface{} {
	players := make([]map[string]interface{}, len(game.Players))
	for i, p := range game.Players {
		player := map[string]interface{}{
			"id":         p.ID,
			"username":   p.Username,
			"chips":      p.Chips,
			"bet":        p.Bet,
			"position":   p.Position,
			"folded":     p.Folded,
			"all_in":     p.AllIn,
			"hole_cards": p.HoleCards, // Show all cards at showdown
		}
		
		// Add hand evaluation for non-folded players
		if !p.Folded {
			allCards := append(p.HoleCards, game.CommunityCards...)
			hand := poker.EvaluateHand(allCards)
			player["hand"] = hand.Rank.String()
			player["best_five"] = hand.BestFive
		}
		
		players[i] = player
	}
	return players
}

// getWinners returns the authoritative winner information recorded by the
// poker engine (Showdown / AwardToLastPlayer). It no longer recomputes winners
// or pot splits independently — doing so previously risked diverging from the
// chip amounts actually credited to players (e.g. with side pots or split
// pots). The engine is the single source of truth.
func (m *Manager) getWinners(game *poker.Game) []map[string]interface{} {
	winners := make([]map[string]interface{}, 0, len(game.Winners))

	for _, w := range game.Winners {
		entry := map[string]interface{}{
			"player_id": w.PlayerID,
			"username":  w.Username,
			"amount":    w.Amount,
		}
		// HandRank/BestFive are empty for a win by fold; include them only
		// when a showdown actually evaluated a hand.
		if w.HandRank != "" {
			entry["hand"] = w.HandRank
			entry["cards"] = w.BestFive
		}
		winners = append(winners, entry)
	}

	return winners
}

// ---------------------------------------------------------------------------
// Admin support methods
// ---------------------------------------------------------------------------

// LiveGameInfo is a lightweight snapshot of an in-progress game for admin views.
type LiveGameInfo struct {
	RoomID      uint   `json:"room_id"`
	GameID      uint   `json:"game_id"`
	Phase       string `json:"phase"`
	PlayerCount int    `json:"player_count"`
	Pot         int64  `json:"pot"`
	CurrentBet  int64  `json:"current_bet"`
}

// ListLiveGames returns a snapshot of all in-memory games currently in progress.
// It takes the read lock so it is safe to call concurrently with gameplay.
func (m *Manager) ListLiveGames() []LiveGameInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := make([]LiveGameInfo, 0, len(m.Games))
	for roomID, g := range m.Games {
		pot := int64(0)
		for _, p := range g.Pots {
			pot += p.Amount
		}
		out = append(out, LiveGameInfo{
			RoomID:      roomID,
			GameID:      g.ID,
			Phase:       string(g.Phase),
			PlayerCount: len(g.Players),
			Pot:         pot,
			CurrentBet:  g.CurrentBet,
		})
	}
	return out
}

// LiveGameCount returns the number of games currently in progress.
func (m *Manager) LiveGameCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.Games)
}

// GameState returns a sanitized full game-state snapshot for a room, or an
// error if no live game exists. Hole cards are hidden (forPlayerID = 0).
func (m *Manager) GameState(roomID uint) (map[string]interface{}, error) {
	game, ok := m.getGame(roomID)
	if !ok {
		return nil, fmt.Errorf("no active game for room %d", roomID)
	}
	return m.serializeGameState(game, 0), nil
}

// ForceEndGame ends the in-progress game in a room immediately (admin action).
// It reuses EndRound so chips and winners are settled consistently.
func (m *Manager) ForceEndGame(roomID uint) error {
	if _, ok := m.getGame(roomID); !ok {
		return fmt.Errorf("no active game for room %d", roomID)
	}
	return m.EndRound(roomID)
}

// KickPlayer removes a player from a room's membership set and broadcasts a
// kick event. It does not forcibly end an in-progress hand; the player simply
// won't be seated for the next one.
func (m *Manager) KickPlayer(roomID, userID uint) error {
	ctx := context.Background()
	key := fmt.Sprintf("room:%d:players", roomID)
	removed, err := m.Redis.SRem(ctx, key, userID).Result()
	if err != nil {
		return err
	}
	if removed == 0 {
		return fmt.Errorf("player %d is not in room %d", userID, roomID)
	}
	m.invalidateRoomCache(ctx, roomID)

	kickData, _ := json.Marshal(map[string]interface{}{
		"type":      "player_kicked",
		"room_id":   fmt.Sprintf("%d", roomID),
		"player_id": userID,
	})
	m.Redis.Publish(ctx, fmt.Sprintf("room:%d", roomID), kickData)
	return nil
}

// CloseRoom force-closes a room: ends any live game, marks the room finished,
// clears its membership set, and clears caches. Soft-deletes the DB row.
func (m *Manager) CloseRoom(roomID uint) error {
	ctx := context.Background()

	// End any in-progress game first so chips settle.
	if _, ok := m.getGame(roomID); ok {
		_ = m.EndRound(roomID)
	}

	// Mark finished and clear membership + caches.
	if err := m.DB.Model(&models.Room{}).Where("id = ?", roomID).Update("status", "finished").Error; err != nil {
		return err
	}
	m.Redis.Del(ctx, fmt.Sprintf("room:%d:players", roomID))
	m.Redis.Del(ctx, fmt.Sprintf("room:%d:starting", roomID))
	m.invalidateRoomCache(ctx, roomID)

	closeData, _ := json.Marshal(map[string]interface{}{
		"type":    "room_closed",
		"room_id": fmt.Sprintf("%d", roomID),
	})
	m.Redis.Publish(ctx, fmt.Sprintf("room:%d", roomID), closeData)

	// Soft-delete the room row.
	return m.DB.Delete(&models.Room{}, roomID).Error
}
