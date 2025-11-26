package anticheat

import (
	"go-poker-arena/internal/poker"
	"testing"
	"time"
)

func TestNewValidator(t *testing.T) {
	v := NewValidator()
	
	if v == nil {
		t.Error("NewValidator should not return nil")
	}
	
	if v.playerActions == nil {
		t.Error("playerActions map should be initialized")
	}
	
	if v.lastAction == nil {
		t.Error("lastAction map should be initialized")
	}
}

func TestCheckLatency(t *testing.T) {
	v := NewValidator()
	
	// Test valid latency
	err := v.CheckLatency(100)
	if err != nil {
		t.Errorf("CheckLatency should pass for valid latency: %v", err)
	}
	
	// Test max latency
	err = v.CheckLatency(MaxLatency)
	if err != nil {
		t.Errorf("CheckLatency should pass for max latency: %v", err)
	}
	
	// Test too high latency
	err = v.CheckLatency(MaxLatency + 1)
	if err == nil {
		t.Error("CheckLatency should fail for latency > max")
	}
	
	// Test negative latency
	err = v.CheckLatency(-1)
	if err == nil {
		t.Error("CheckLatency should fail for negative latency")
	}
}

func TestValidateActionInterval(t *testing.T) {
	v := NewValidator()
	game := createTestGame()
	
	playerID := uint(1)
	
	// First action should pass
	err := v.ValidateAction(playerID, poker.ActionCall, 0, game)
	if err != nil {
		t.Errorf("First action should pass: %v", err)
	}
	
	// Immediate second action should fail
	err = v.ValidateAction(playerID, poker.ActionCall, 0, game)
	if err == nil {
		t.Error("Immediate second action should fail")
	}
	
	// Wait and try again
	time.Sleep(MinActionInterval * time.Millisecond)
	err = v.ValidateAction(playerID, poker.ActionCall, 0, game)
	if err != nil {
		t.Errorf("Action after interval should pass: %v", err)
	}
}

func TestValidateActionRate(t *testing.T) {
	v := NewValidator()
	game := createTestGame()
	
	playerID := uint(1)
	
	// Simulate many actions
	for i := 0; i < MaxActionsPerMin; i++ {
		time.Sleep(MinActionInterval * time.Millisecond)
		v.ValidateAction(playerID, poker.ActionCheck, 0, game)
	}
	
	// Next action should fail
	time.Sleep(MinActionInterval * time.Millisecond)
	err := v.ValidateAction(playerID, poker.ActionCheck, 0, game)
	if err == nil {
		t.Error("Should fail after max actions per minute")
	}
}

func TestValidateActionPlayerNotFound(t *testing.T) {
	v := NewValidator()
	game := createTestGame()
	
	err := v.ValidateAction(999, poker.ActionCall, 0, game)
	if err == nil {
		t.Error("Should fail for non-existent player")
	}
}

func TestValidateActionFoldedPlayer(t *testing.T) {
	v := NewValidator()
	game := createTestGame()
	
	// Fold player
	game.Players[0].Folded = true
	
	err := v.ValidateAction(1, poker.ActionCall, 0, game)
	if err == nil {
		t.Error("Should fail for folded player")
	}
}

func TestValidateActionAllInPlayer(t *testing.T) {
	v := NewValidator()
	game := createTestGame()
	
	// Set player all-in
	game.Players[0].AllIn = true
	
	err := v.ValidateAction(1, poker.ActionCall, 0, game)
	if err == nil {
		t.Error("Should fail for all-in player")
	}
}

func TestValidateRaiseAmount(t *testing.T) {
	v := NewValidator()
	game := createTestGame()
	game.MinRaise = 100
	game.Players[0].Chips = 1000
	
	// Test raise too small
	err := v.ValidateAction(1, poker.ActionRaise, 50, game)
	if err == nil {
		t.Error("Should fail for raise amount < min raise")
	}
	
	// Test raise too large
	err = v.ValidateAction(1, poker.ActionRaise, 2000, game)
	if err == nil {
		t.Error("Should fail for raise amount > chips")
	}
	
	// Test valid raise
	time.Sleep(MinActionInterval * time.Millisecond)
	err = v.ValidateAction(1, poker.ActionRaise, 100, game)
	if err != nil {
		t.Errorf("Should pass for valid raise: %v", err)
	}
}

func TestValidateCallAmount(t *testing.T) {
	v := NewValidator()
	game := createTestGame()
	game.CurrentBet = 100
	game.Players[0].Bet = 0
	game.Players[0].Chips = 50
	
	// Test insufficient chips to call
	err := v.ValidateAction(1, poker.ActionCall, 0, game)
	if err == nil {
		t.Error("Should fail for insufficient chips to call")
	}
}

func TestValidateCheck(t *testing.T) {
	v := NewValidator()
	game := createTestGame()
	game.CurrentBet = 100
	game.Players[0].Bet = 0
	
	// Test check when bet required
	err := v.ValidateAction(1, poker.ActionCheck, 0, game)
	if err == nil {
		t.Error("Should fail for check when bet required")
	}
	
	// Test valid check
	game.Players[0].Bet = 100
	time.Sleep(MinActionInterval * time.Millisecond)
	err = v.ValidateAction(1, poker.ActionCheck, 0, game)
	if err != nil {
		t.Errorf("Should pass for valid check: %v", err)
	}
}

func TestGetPartialGameState(t *testing.T) {
	game := createTestGame()
	game.Phase = poker.PhaseFlop
	
	// Add hole cards
	game.Players[0].HoleCards = []poker.Card{
		{Suit: poker.Hearts, Rank: poker.Ace},
		{Suit: poker.Diamonds, Rank: poker.King},
	}
	game.Players[1].HoleCards = []poker.Card{
		{Suit: poker.Clubs, Rank: poker.Queen},
		{Suit: poker.Spades, Rank: poker.Jack},
	}
	
	// Get partial state for player 1
	state := GetPartialGameState(game, 1)
	
	if state == nil {
		t.Fatal("State should not be nil")
	}
	
	// Check basic fields
	if state["id"] != game.ID {
		t.Error("ID should match")
	}
	
	if state["phase"] != game.Phase {
		t.Error("Phase should match")
	}
	
	// Check players
	players := state["players"].([]map[string]interface{})
	if len(players) != 2 {
		t.Errorf("Expected 2 players, got %d", len(players))
	}
	
	// Player 1 should see their own cards
	if players[0]["hole_cards"] == nil {
		t.Error("Player should see their own hole cards")
	}
	
	// Player 1 should not see player 2's cards
	if players[1]["hole_cards"] != nil {
		t.Error("Player should not see opponent's hole cards")
	}
	
	// Player 1 should see card count for opponent
	if players[1]["card_count"] == nil {
		t.Error("Player should see opponent's card count")
	}
}

func TestGetPartialGameStateShowdown(t *testing.T) {
	game := createTestGame()
	game.Phase = poker.PhaseShowdown
	
	game.Players[0].HoleCards = []poker.Card{
		{Suit: poker.Hearts, Rank: poker.Ace},
	}
	game.Players[1].HoleCards = []poker.Card{
		{Suit: poker.Clubs, Rank: poker.King},
	}
	
	// During showdown, all cards should be visible
	state := GetPartialGameState(game, 1)
	players := state["players"].([]map[string]interface{})
	
	// Both players' cards should be visible
	if players[0]["hole_cards"] == nil {
		t.Error("Should see own cards during showdown")
	}
	
	if players[1]["hole_cards"] == nil {
		t.Error("Should see opponent's cards during showdown")
	}
}

func TestDetectCollusion(t *testing.T) {
	v := NewValidator()
	
	// Test with no actions
	result := v.DetectCollusion(1, 2)
	if result {
		t.Error("Should not detect collusion with no actions")
	}
	
	// Test with few actions
	for i := 0; i < 5; i++ {
		v.playerActions[1] = append(v.playerActions[1], time.Now())
		v.playerActions[2] = append(v.playerActions[2], time.Now())
	}
	
	result = v.DetectCollusion(1, 2)
	if result {
		t.Error("Should not detect collusion with few actions")
	}
}

// Helper function to create test game
func createTestGame() *poker.Game {
	players := []*poker.Player{
		{
			ID:       1,
			Username: "player1",
			Chips:    1000,
			Bet:      0,
			Position: 0,
			Folded:   false,
			AllIn:    false,
		},
		{
			ID:       2,
			Username: "player2",
			Chips:    1000,
			Bet:      0,
			Position: 1,
			Folded:   false,
			AllIn:    false,
		},
	}
	
	game := poker.NewGame(1, players, 10, 20)
	game.ID = 1
	game.Phase = poker.PhasePreFlop
	game.CurrentBet = 0
	game.MinRaise = 20
	
	return game
}
