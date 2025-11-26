package rooms

import (
	"context"
	"encoding/json"
	"fmt"
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

	return m.Redis.SAdd(ctx, key, userID).Err()
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

	// Save game to database
	dbGame := &models.Game{
		RoomID: roomID,
		Status: "active",
		Pot:    0,
		Stage:  string(game.Phase),
	}
	if err := m.DB.Create(dbGame).Error; err != nil {
		return nil, err
	}
	game.ID = dbGame.ID

	// Publish game start event
	ctx := context.Background()
	gameData, _ := json.Marshal(map[string]interface{}{
		"type":    "game_start",
		"room_id": roomID,
		"game":    game,
	})
	m.Redis.Publish(ctx, fmt.Sprintf("room:%d", roomID), gameData)

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
		"room_id":   roomID,
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
		"room_id":         roomID,
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
		"room_id": roomID,
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
