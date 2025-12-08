package rooms

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"go-poker-arena/internal/models"
	"go-poker-arena/internal/poker"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Manager struct {
	DB    *gorm.DB
	Redis *redis.Client
	Games map[uint]*poker.Game
}

func NewManager(db *gorm.DB, redisClient *redis.Client) *Manager {
	return &Manager{
		DB:    db,
		Redis: redisClient,
		Games: make(map[uint]*poker.Game),
	}
}

func (m *Manager) CreateRoom(name string, maxPlayers int, smallBlind, bigBlind int64) (*models.Room, error) {
	room := &models.Room{
		Name:       name,
		MaxPlayers: maxPlayers,
		SmallBlind: smallBlind,
		BigBlind:   bigBlind,
		Status:     "waiting",
	}

	if err := m.DB.Create(room).Error; err != nil {
		return nil, err
	}

	ctx := context.Background()
	roomData, _ := json.Marshal(room)
	m.Redis.Set(ctx, fmt.Sprintf("room:%d", room.ID), roomData, 0)

	return room, nil
}

func (m *Manager) GetRoom(roomID uint) (*models.Room, error) {
	ctx := context.Background()
	key := fmt.Sprintf("room:%d", roomID)

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

	roomData, _ := json.Marshal(room)
	m.Redis.Set(ctx, key, roomData, 0)

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
	
	count, err := m.Redis.SCard(ctx, key).Result()
	if err != nil {
		return err
	}

	room, err := m.GetRoom(roomID)
	if err != nil {
		return err
	}

	if int(count) >= room.MaxPlayers {
		return fmt.Errorf("room is full")
	}

	if err := m.Redis.SAdd(ctx, key, userID).Err(); err != nil {
		return err
	}

	// Check if we have enough players to start the game
	newCount := count + 1
	fmt.Printf("[ROOM %d] Player %d joined → %d players total\n", roomID, userID, newCount)
	
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

	return nil
}

func (m *Manager) LeaveRoom(roomID, userID uint) error {
	ctx := context.Background()
	key := fmt.Sprintf("room:%d:players", roomID)
	return m.Redis.SRem(ctx, key, userID).Err()
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

	m.Games[roomID] = game

	// Update room status
	room.Status = "playing"
	m.DB.Model(room).Update("status", "playing")

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

	ctx := context.Background()
	
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
	game, ok := m.Games[roomID]
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

	// Publish action event
	ctx := context.Background()
	actionData, _ := json.Marshal(map[string]interface{}{
		"type":      "player_action",
		"room_id":   fmt.Sprintf("%d", roomID),
		"player_id": playerID,
		"action":    action,
		"amount":    amount,
		"game":      m.serializeGameState(game, 0),
	})
	m.Redis.Publish(ctx, fmt.Sprintf("room:%d", roomID), actionData)

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

	// Publish game end event
	ctx := context.Background()
	endData, _ := json.Marshal(map[string]interface{}{
		"type":    "game_end",
		"room_id": fmt.Sprintf("%d", roomID),
		"winners": m.getWinners(game),
	})
	m.Redis.Publish(ctx, fmt.Sprintf("room:%d", roomID), endData)

	delete(m.Games, roomID)
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

func (m *Manager) getWinners(game *poker.Game) []map[string]interface{} {
	winners := make([]map[string]interface{}, 0)
	for _, player := range game.Players {
		if !player.Folded {
			allCards := append(player.HoleCards, game.CommunityCards...)
			hand := poker.EvaluateHand(allCards)
			winners = append(winners, map[string]interface{}{
				"player_id": player.ID,
				"username":  player.Username,
				"hand":      hand.Rank.String(),
				"cards":     hand.BestFive,
			})
		}
	}
	return winners
}
