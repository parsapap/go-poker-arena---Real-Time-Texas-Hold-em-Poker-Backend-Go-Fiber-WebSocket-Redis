package leaderboard

import (
	"context"
	"fmt"
	"strconv"
	"github.com/redis/go-redis/v9"
)

type Leaderboard struct {
	Redis *redis.Client
}

func NewLeaderboard(redisClient *redis.Client) *Leaderboard {
	return &Leaderboard{Redis: redisClient}
}

type PlayerStats struct {
	UserID   uint    `json:"user_id"`
	Username string  `json:"username"`
	Wins     int64   `json:"wins"`
	Chips    int64   `json:"chips"`
	Score    float64 `json:"score"`
	Rank     int64   `json:"rank"`
}

// UpdateWins increments player's win count
func (l *Leaderboard) UpdateWins(userID uint, username string, wins int64) error {
	ctx := context.Background()
	key := "leaderboard:wins"
	
	// Store username mapping
	l.Redis.HSet(ctx, "leaderboard:usernames", userID, username)
	
	return l.Redis.ZIncrBy(ctx, key, float64(wins), fmt.Sprintf("%d", userID)).Err()
}

// UpdateChips updates player's chip count
func (l *Leaderboard) UpdateChips(userID uint, chips int64) error {
	ctx := context.Background()
	key := "leaderboard:chips"
	return l.Redis.ZAdd(ctx, key, redis.Z{
		Score:  float64(chips),
		Member: fmt.Sprintf("%d", userID),
	}).Err()
}

// GetTopByWins returns top N players by wins
func (l *Leaderboard) GetTopByWins(limit int64) ([]PlayerStats, error) {
	ctx := context.Background()
	key := "leaderboard:wins"
	
	results, err := l.Redis.ZRevRangeWithScores(ctx, key, 0, limit-1).Result()
	if err != nil {
		return nil, err
	}
	
	return l.parseResults(results)
}

// GetTopByChips returns top N players by chips
func (l *Leaderboard) GetTopByChips(limit int64) ([]PlayerStats, error) {
	ctx := context.Background()
	key := "leaderboard:chips"
	
	results, err := l.Redis.ZRevRangeWithScores(ctx, key, 0, limit-1).Result()
	if err != nil {
		return nil, err
	}
	
	return l.parseResults(results)
}

// GetPlayerRank returns player's rank by wins
func (l *Leaderboard) GetPlayerRank(userID uint) (int64, error) {
	ctx := context.Background()
	key := "leaderboard:wins"
	
	rank, err := l.Redis.ZRevRank(ctx, key, fmt.Sprintf("%d", userID)).Result()
	if err != nil {
		return 0, err
	}
	
	return rank + 1, nil
}

// GetPlayerStats returns complete player statistics
func (l *Leaderboard) GetPlayerStats(userID uint) (*PlayerStats, error) {
	ctx := context.Background()
	
	wins, err := l.Redis.ZScore(ctx, "leaderboard:wins", fmt.Sprintf("%d", userID)).Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}
	
	chips, err := l.Redis.ZScore(ctx, "leaderboard:chips", fmt.Sprintf("%d", userID)).Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}
	
	rank, _ := l.GetPlayerRank(userID)
	
	username, _ := l.Redis.HGet(ctx, "leaderboard:usernames", fmt.Sprintf("%d", userID)).Result()
	
	return &PlayerStats{
		UserID:   userID,
		Username: username,
		Wins:     int64(wins),
		Chips:    int64(chips),
		Score:    wins,
		Rank:     rank,
	}, nil
}

func (l *Leaderboard) parseResults(results []redis.Z) ([]PlayerStats, error) {
	ctx := context.Background()
	stats := make([]PlayerStats, 0, len(results))
	
	for i, result := range results {
		userID, _ := strconv.ParseUint(result.Member.(string), 10, 64)
		username, _ := l.Redis.HGet(ctx, "leaderboard:usernames", result.Member.(string)).Result()
		
		stats = append(stats, PlayerStats{
			UserID:   uint(userID),
			Username: username,
			Score:    result.Score,
			Rank:     int64(i + 1),
		})
	}
	
	return stats, nil
}
