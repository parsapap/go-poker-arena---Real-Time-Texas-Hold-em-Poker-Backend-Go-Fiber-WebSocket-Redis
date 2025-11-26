package history

import (
	"go-poker-arena/internal/models"
	"go-poker-arena/internal/poker"
	"testing"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Migrate tables
	err = db.AutoMigrate(&models.User{}, &models.GameHistory{}, &models.PlayerAction{})
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	return db
}

func TestNewService(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)

	if service == nil {
		t.Error("NewService should not return nil")
	}

	if service.DB == nil {
		t.Error("Service DB should not be nil")
	}
}

func TestSaveGame(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)

	// Create test game
	players := []*poker.Player{
		{ID: 1, Username: "player1", Chips: 1000},
		{ID: 2, Username: "player2", Chips: 2000},
	}
	game := poker.NewGame(1, players, 10, 20)
	game.ID = 1
	game.CommunityCards = []poker.Card{
		{Suit: poker.Hearts, Rank: poker.Ace},
		{Suit: poker.Diamonds, Rank: poker.King},
	}
	game.Pots = []poker.Pot{{Amount: 500, Players: []uint{1, 2}}}

	// Save game
	err := service.SaveGame(game, 1, 300)
	if err != nil {
		t.Errorf("SaveGame failed: %v", err)
	}

	// Verify game was saved
	var history models.GameHistory
	db.First(&history)

	if history.GameID != 1 {
		t.Errorf("Expected GameID 1, got %d", history.GameID)
	}

	if history.WinnerID != 1 {
		t.Errorf("Expected WinnerID 1, got %d", history.WinnerID)
	}

	if history.Pot != 500 {
		t.Errorf("Expected Pot 500, got %d", history.Pot)
	}

	if history.Duration != 300 {
		t.Errorf("Expected Duration 300, got %d", history.Duration)
	}
}

func TestSaveAction(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)

	// Save action
	err := service.SaveAction(1, 1, "raise", 100, "preflop", 150)
	if err != nil {
		t.Errorf("SaveAction failed: %v", err)
	}

	// Verify action was saved
	var action models.PlayerAction
	db.First(&action)

	if action.GameID != 1 {
		t.Errorf("Expected GameID 1, got %d", action.GameID)
	}

	if action.UserID != 1 {
		t.Errorf("Expected UserID 1, got %d", action.UserID)
	}

	if action.Action != "raise" {
		t.Errorf("Expected action 'raise', got '%s'", action.Action)
	}

	if action.Amount != 100 {
		t.Errorf("Expected amount 100, got %d", action.Amount)
	}

	if action.Phase != "preflop" {
		t.Errorf("Expected phase 'preflop', got '%s'", action.Phase)
	}

	if action.Latency != 150 {
		t.Errorf("Expected latency 150, got %d", action.Latency)
	}

	if action.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}
}

func TestSaveMultipleActions(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)

	// Save multiple actions
	service.SaveAction(1, 1, "call", 20, "preflop", 100)
	service.SaveAction(1, 2, "raise", 40, "preflop", 120)
	service.SaveAction(1, 1, "call", 20, "preflop", 110)

	// Verify all actions were saved
	var count int64
	db.Model(&models.PlayerAction{}).Count(&count)

	if count != 3 {
		t.Errorf("Expected 3 actions, got %d", count)
	}
}

func TestGetGameDetails(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)

	// Create test game and actions
	players := []*poker.Player{
		{ID: 1, Username: "player1", Chips: 1000},
	}
	game := poker.NewGame(1, players, 10, 20)
	game.ID = 1
	game.Pots = []poker.Pot{{Amount: 100, Players: []uint{1}}}

	service.SaveGame(game, 1, 100)
	service.SaveAction(1, 1, "call", 20, "preflop", 100)
	service.SaveAction(1, 1, "raise", 40, "flop", 120)

	// Get game details
	history, actions, err := service.GetGameDetails(1)
	if err != nil {
		t.Errorf("GetGameDetails failed: %v", err)
	}

	if history == nil {
		t.Fatal("History should not be nil")
	}

	if history.GameID != 1 {
		t.Errorf("Expected GameID 1, got %d", history.GameID)
	}

	if len(actions) != 2 {
		t.Errorf("Expected 2 actions, got %d", len(actions))
	}

	// Check actions are ordered by timestamp
	if actions[0].Phase != "preflop" {
		t.Error("First action should be preflop")
	}

	if actions[1].Phase != "flop" {
		t.Error("Second action should be flop")
	}
}

func TestGetGameDetailsNotFound(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)

	// Try to get non-existent game
	_, _, err := service.GetGameDetails(999)
	if err == nil {
		t.Error("GetGameDetails should fail for non-existent game")
	}
}

func TestGetPlayerStats(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)

	// Create test user
	user := &models.User{
		Username: "player1",
		Chips:    5000,
		Wins:     10,
		Losses:   5,
	}
	db.Create(user)

	// Save some actions
	service.SaveAction(1, user.ID, "call", 20, "preflop", 100)
	service.SaveAction(2, user.ID, "raise", 40, "flop", 120)

	// Get stats
	stats, err := service.GetPlayerStats(user.ID)
	if err != nil {
		t.Errorf("GetPlayerStats failed: %v", err)
	}

	if stats == nil {
		t.Fatal("Stats should not be nil")
	}

	if stats["user_id"] != user.ID {
		t.Errorf("Expected user_id %d, got %v", user.ID, stats["user_id"])
	}

	if stats["username"] != "player1" {
		t.Errorf("Expected username 'player1', got %v", stats["username"])
	}

	if stats["chips"] != int64(5000) {
		t.Errorf("Expected 5000 chips, got %v", stats["chips"])
	}

	if stats["wins"] != int64(10) {
		t.Errorf("Expected 10 wins, got %v", stats["wins"])
	}

	if stats["losses"] != int64(5) {
		t.Errorf("Expected 5 losses, got %v", stats["losses"])
	}

	if stats["total_actions"] != int64(2) {
		t.Errorf("Expected 2 total actions, got %v", stats["total_actions"])
	}
}

func TestGetPlayerStatsWinRate(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)

	// Create user with wins and losses
	user := &models.User{
		Username: "player1",
		Chips:    1000,
		Wins:     6,
		Losses:   4,
	}
	db.Create(user)

	// Get stats
	stats, _ := service.GetPlayerStats(user.ID)

	// Win rate should be 0 when no games in history
	if stats["win_rate"].(float64) != 0 {
		t.Errorf("Expected win rate 0 with no games, got %v", stats["win_rate"])
	}
}

func TestGetPlayerStatsNotFound(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)

	// Try to get stats for non-existent user
	_, err := service.GetPlayerStats(999)
	if err == nil {
		t.Error("GetPlayerStats should fail for non-existent user")
	}
}
