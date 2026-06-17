package history

import (
	"encoding/json"
	"fmt"
	"time"
	"go-poker-arena/internal/models"
	"go-poker-arena/internal/poker"
	"gorm.io/gorm"
)

type Service struct {
	DB *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{DB: db}
}

// SaveGame saves completed game to history
func (s *Service) SaveGame(game *poker.Game, winnerID uint, duration int) error {
	// Serialize players
	playersJSON, err := json.Marshal(game.Players)
	if err != nil {
		return err
	}

	// Get final hands
	finalHands := make(map[uint]string)
	for _, player := range game.Players {
		if !player.Folded {
			allCards := append(player.HoleCards, game.CommunityCards...)
			hand := poker.EvaluateHand(allCards)
			finalHands[player.ID] = hand.Rank.String()
		}
	}
	finalHandsJSON, _ := json.Marshal(finalHands)

	history := &models.GameHistory{
		GameID:     game.ID,
		RoomID:     game.RoomID,
		WinnerID:   winnerID,
		Pot:        game.Pots[0].Amount,
		Players:    string(playersJSON),
		Duration:   duration,
		FinalHands: string(finalHandsJSON),
	}

	return s.DB.Create(history).Error
}

// SaveAction saves player action
func (s *Service) SaveAction(gameID, userID uint, action string, amount int64, phase string, latency int) error {
	playerAction := &models.PlayerAction{
		GameID:    gameID,
		UserID:    userID,
		Action:    action,
		Amount:    amount,
		Phase:     phase,
		Timestamp: time.Now(),
		Latency:   latency,
	}

	return s.DB.Create(playerAction).Error
}

// GetUserHistory retrieves user's game history
func (s *Service) GetUserHistory(userID uint, limit int) ([]models.GameHistory, error) {
	var history []models.GameHistory
	
	err := s.DB.Raw(`
		SELECT gh.* FROM game_histories gh
		WHERE gh.players::jsonb @> ?
		ORDER BY gh.created_at DESC
		LIMIT ?
	`, json.RawMessage(fmt.Sprintf(`[{"id":%d}]`, userID)), limit).Scan(&history).Error
	
	if err != nil {
		return nil, err
	}

	return history, nil
}

// GetGameDetails retrieves detailed game information
func (s *Service) GetGameDetails(gameID uint) (*models.GameHistory, []models.PlayerAction, error) {
	var history models.GameHistory
	if err := s.DB.First(&history, gameID).Error; err != nil {
		return nil, nil, err
	}

	var actions []models.PlayerAction
	if err := s.DB.Where("game_id = ?", gameID).Order("created_at ASC").Find(&actions).Error; err != nil {
		return nil, nil, err
	}

	return &history, actions, nil
}

// GetPlayerStats retrieves player statistics
func (s *Service) GetPlayerStats(userID uint) (map[string]interface{}, error) {
	var user models.User
	if err := s.DB.First(&user, userID).Error; err != nil {
		return nil, err
	}

	var totalGames int64
	s.DB.Model(&models.GameHistory{}).
		Where("players::jsonb @> ?", json.RawMessage(fmt.Sprintf(`[{"id":%d}]`, userID))).
		Count(&totalGames)

	var totalActions int64
	s.DB.Model(&models.PlayerAction{}).Where("user_id = ?", userID).Count(&totalActions)

	winRate := float64(0)
	if totalGames > 0 {
		winRate = float64(user.Wins) / float64(totalGames) * 100
	}

	return map[string]interface{}{
		"user_id":      user.ID,
		"username":     user.Username,
		"chips":        user.Chips,
		"wins":         user.Wins,
		"losses":       user.Losses,
		"total_games":  totalGames,
		"win_rate":     winRate,
		"total_actions": totalActions,
	}, nil
}
