package leaderboard

import (
	"context"
	"testing"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func setupTestRedis(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return client, mr
}

func TestNewLeaderboard(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	lb := NewLeaderboard(client)
	if lb == nil {
		t.Error("NewLeaderboard should not return nil")
	}

	if lb.Redis == nil {
		t.Error("Redis client should not be nil")
	}
}

func TestUpdateWins(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	lb := NewLeaderboard(client)

	err := lb.UpdateWins(1, "player1", 1)
	if err != nil {
		t.Errorf("UpdateWins failed: %v", err)
	}

	// Verify wins were updated
	ctx := context.Background()
	score, err := client.ZScore(ctx, "leaderboard:wins", "1").Result()
	if err != nil {
		t.Errorf("Failed to get score: %v", err)
	}

	if score != 1 {
		t.Errorf("Expected score 1, got %f", score)
	}

	// Verify username was stored
	username, err := client.HGet(ctx, "leaderboard:usernames", "1").Result()
	if err != nil {
		t.Errorf("Failed to get username: %v", err)
	}

	if username != "player1" {
		t.Errorf("Expected username 'player1', got '%s'", username)
	}
}

func TestUpdateWinsIncrement(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	lb := NewLeaderboard(client)

	// Add initial wins
	lb.UpdateWins(1, "player1", 5)

	// Increment wins
	lb.UpdateWins(1, "player1", 3)

	ctx := context.Background()
	score, _ := client.ZScore(ctx, "leaderboard:wins", "1").Result()

	if score != 8 {
		t.Errorf("Expected score 8, got %f", score)
	}
}

func TestUpdateChips(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	lb := NewLeaderboard(client)

	err := lb.UpdateChips(1, 5000)
	if err != nil {
		t.Errorf("UpdateChips failed: %v", err)
	}

	ctx := context.Background()
	score, err := client.ZScore(ctx, "leaderboard:chips", "1").Result()
	if err != nil {
		t.Errorf("Failed to get score: %v", err)
	}

	if score != 5000 {
		t.Errorf("Expected score 5000, got %f", score)
	}
}

func TestGetTopByWins(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	lb := NewLeaderboard(client)

	// Add multiple players
	lb.UpdateWins(1, "player1", 10)
	lb.UpdateWins(2, "player2", 20)
	lb.UpdateWins(3, "player3", 15)

	// Get top 2
	stats, err := lb.GetTopByWins(2)
	if err != nil {
		t.Errorf("GetTopByWins failed: %v", err)
	}

	if len(stats) != 2 {
		t.Errorf("Expected 2 players, got %d", len(stats))
	}

	// Check order (highest first)
	if stats[0].UserID != 2 {
		t.Errorf("Expected player 2 first, got %d", stats[0].UserID)
	}

	if stats[0].Score != 20 {
		t.Errorf("Expected score 20, got %f", stats[0].Score)
	}

	if stats[0].Rank != 1 {
		t.Errorf("Expected rank 1, got %d", stats[0].Rank)
	}

	if stats[1].UserID != 3 {
		t.Errorf("Expected player 3 second, got %d", stats[1].UserID)
	}
}

func TestGetTopByChips(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	lb := NewLeaderboard(client)

	// Add multiple players
	lb.UpdateChips(1, 1000)
	lb.UpdateChips(2, 5000)
	lb.UpdateChips(3, 3000)

	// Get top 3
	stats, err := lb.GetTopByChips(3)
	if err != nil {
		t.Errorf("GetTopByChips failed: %v", err)
	}

	if len(stats) != 3 {
		t.Errorf("Expected 3 players, got %d", len(stats))
	}

	// Check order
	if stats[0].UserID != 2 {
		t.Errorf("Expected player 2 first, got %d", stats[0].UserID)
	}

	if stats[1].UserID != 3 {
		t.Errorf("Expected player 3 second, got %d", stats[1].UserID)
	}

	if stats[2].UserID != 1 {
		t.Errorf("Expected player 1 third, got %d", stats[2].UserID)
	}
}

func TestGetPlayerRank(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	lb := NewLeaderboard(client)

	// Add players
	lb.UpdateWins(1, "player1", 10)
	lb.UpdateWins(2, "player2", 20)
	lb.UpdateWins(3, "player3", 15)

	// Get rank for player 2 (should be 1st)
	rank, err := lb.GetPlayerRank(2)
	if err != nil {
		t.Errorf("GetPlayerRank failed: %v", err)
	}

	if rank != 1 {
		t.Errorf("Expected rank 1, got %d", rank)
	}

	// Get rank for player 3 (should be 2nd)
	rank, _ = lb.GetPlayerRank(3)
	if rank != 2 {
		t.Errorf("Expected rank 2, got %d", rank)
	}

	// Get rank for player 1 (should be 3rd)
	rank, _ = lb.GetPlayerRank(1)
	if rank != 3 {
		t.Errorf("Expected rank 3, got %d", rank)
	}
}

func TestGetPlayerStats(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	lb := NewLeaderboard(client)

	// Add player data
	lb.UpdateWins(1, "player1", 10)
	lb.UpdateChips(1, 5000)

	// Get stats
	stats, err := lb.GetPlayerStats(1)
	if err != nil {
		t.Errorf("GetPlayerStats failed: %v", err)
	}

	if stats.UserID != 1 {
		t.Errorf("Expected UserID 1, got %d", stats.UserID)
	}

	if stats.Username != "player1" {
		t.Errorf("Expected username 'player1', got '%s'", stats.Username)
	}

	if stats.Wins != 10 {
		t.Errorf("Expected 10 wins, got %d", stats.Wins)
	}

	if stats.Chips != 5000 {
		t.Errorf("Expected 5000 chips, got %d", stats.Chips)
	}

	if stats.Rank != 1 {
		t.Errorf("Expected rank 1, got %d", stats.Rank)
	}
}

func TestGetPlayerStatsNotFound(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	lb := NewLeaderboard(client)

	// Get stats for non-existent player
	stats, err := lb.GetPlayerStats(999)
	if err != nil {
		t.Errorf("GetPlayerStats should not error for non-existent player: %v", err)
	}

	if stats.Wins != 0 {
		t.Errorf("Expected 0 wins for non-existent player, got %d", stats.Wins)
	}
}
