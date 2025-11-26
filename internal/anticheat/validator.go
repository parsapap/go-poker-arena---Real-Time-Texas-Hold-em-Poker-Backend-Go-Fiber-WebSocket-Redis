package anticheat

import (
	"errors"
	"time"
	"go-poker-arena/internal/poker"
)

const (
	MaxLatency        = 5000  // 5 seconds max latency
	MinActionInterval = 100   // 100ms minimum between actions
	MaxActionsPerMin  = 60    // Max 60 actions per minute
)

type Validator struct {
	playerActions map[uint][]time.Time
	lastAction    map[uint]time.Time
}

func NewValidator() *Validator {
	return &Validator{
		playerActions: make(map[uint][]time.Time),
		lastAction:    make(map[uint]time.Time),
	}
}

// ValidateAction checks if action is valid and not suspicious
func (v *Validator) ValidateAction(playerID uint, action poker.Action, amount int64, game *poker.Game) error {
	now := time.Now()

	// Check action interval
	if last, ok := v.lastAction[playerID]; ok {
		if now.Sub(last) < MinActionInterval*time.Millisecond {
			return errors.New("actions too fast - possible bot")
		}
	}
	v.lastAction[playerID] = now

	// Track actions per minute
	v.playerActions[playerID] = append(v.playerActions[playerID], now)
	
	// Clean old actions (older than 1 minute)
	cutoff := now.Add(-1 * time.Minute)
	filtered := make([]time.Time, 0)
	for _, t := range v.playerActions[playerID] {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}
	v.playerActions[playerID] = filtered

	// Check action rate
	if len(v.playerActions[playerID]) > MaxActionsPerMin {
		return errors.New("too many actions - possible bot")
	}

	// Validate action against game state
	player := game.GetPlayer(playerID)
	if player == nil {
		return errors.New("player not found")
	}

	if player.Folded {
		return errors.New("player already folded")
	}

	if player.AllIn {
		return errors.New("player is all-in")
	}

	// Validate bet amounts
	switch action {
	case poker.ActionRaise:
		if amount < game.MinRaise {
			return errors.New("raise amount too small")
		}
		if amount > player.Chips {
			return errors.New("insufficient chips")
		}
	case poker.ActionCall:
		callAmount := game.CurrentBet - player.Bet
		if callAmount > player.Chips {
			return errors.New("insufficient chips to call")
		}
	case poker.ActionCheck:
		if player.Bet < game.CurrentBet {
			return errors.New("cannot check - must call or raise")
		}
	}

	return nil
}

// CheckLatency validates action latency
func (v *Validator) CheckLatency(latency int) error {
	if latency > MaxLatency {
		return errors.New("latency too high - possible network manipulation")
	}
	if latency < 0 {
		return errors.New("invalid latency value")
	}
	return nil
}

// GetPartialGameState returns game state with hidden opponent cards
func GetPartialGameState(game *poker.Game, forPlayerID uint) map[string]interface{} {
	players := make([]map[string]interface{}, len(game.Players))
	
	for i, p := range game.Players {
		player := map[string]interface{}{
			"id":       p.ID,
			"username": p.Username,
			"chips":    p.Chips,
			"bet":      p.Bet,
			"position": p.Position,
			"folded":   p.Folded,
			"all_in":   p.AllIn,
		}
		
		// Only show hole cards to the player themselves or during showdown
		if p.ID == forPlayerID || game.Phase == poker.PhaseShowdown {
			player["hole_cards"] = p.HoleCards
		} else {
			// Show card count but not actual cards
			player["card_count"] = len(p.HoleCards)
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

// DetectCollusion checks for suspicious patterns
func (v *Validator) DetectCollusion(playerID1, playerID2 uint) bool {
	// Check if players always fold when facing each other
	// This is a simplified check - real implementation would be more sophisticated
	actions1 := v.playerActions[playerID1]
	actions2 := v.playerActions[playerID2]
	
	// If both players have very similar action patterns, flag as suspicious
	if len(actions1) > 10 && len(actions2) > 10 {
		// More sophisticated pattern matching would go here
		return false
	}
	
	return false
}
