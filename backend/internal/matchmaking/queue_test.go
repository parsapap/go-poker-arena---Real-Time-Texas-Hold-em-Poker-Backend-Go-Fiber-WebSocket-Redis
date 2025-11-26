package matchmaking

import (
	"testing"
	"time"
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

func TestNewQueue(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	q := NewQueue(client)
	if q == nil {
		t.Error("NewQueue should not return nil")
	}

	if q.Redis == nil {
		t.Error("Redis client should not be nil")
	}
}

func TestJoinQueue(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	q := NewQueue(client)

	err := q.JoinQueue(1, "player1", 1000, 1500)
	if err != nil {
		t.Errorf("JoinQueue failed: %v", err)
	}

	// Verify player was added
	size, _ := q.GetQueueSize()
	if size != 1 {
		t.Errorf("Expected queue size 1, got %d", size)
	}
}

func TestJoinQueueMultiplePlayers(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	q := NewQueue(client)

	q.JoinQueue(1, "player1", 1000, 1500)
	q.JoinQueue(2, "player2", 2000, 1600)
	q.JoinQueue(3, "player3", 1500, 1400)

	size, _ := q.GetQueueSize()
	if size != 3 {
		t.Errorf("Expected queue size 3, got %d", size)
	}
}

func TestLeaveQueue(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	q := NewQueue(client)

	// Join and then leave
	q.JoinQueue(1, "player1", 1000, 1500)
	q.JoinQueue(2, "player2", 2000, 1600)

	err := q.LeaveQueue(1)
	if err != nil {
		t.Errorf("LeaveQueue failed: %v", err)
	}

	size, _ := q.GetQueueSize()
	if size != 1 {
		t.Errorf("Expected queue size 1 after leave, got %d", size)
	}
}

func TestLeaveQueueNotInQueue(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	q := NewQueue(client)

	// Try to leave when not in queue
	err := q.LeaveQueue(999)
	if err != nil {
		t.Errorf("LeaveQueue should not error for non-existent player: %v", err)
	}
}

func TestGetQueueSize(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	q := NewQueue(client)

	// Empty queue
	size, err := q.GetQueueSize()
	if err != nil {
		t.Errorf("GetQueueSize failed: %v", err)
	}

	if size != 0 {
		t.Errorf("Expected queue size 0, got %d", size)
	}

	// Add players
	q.JoinQueue(1, "player1", 1000, 1500)
	q.JoinQueue(2, "player2", 2000, 1600)

	size, _ = q.GetQueueSize()
	if size != 2 {
		t.Errorf("Expected queue size 2, got %d", size)
	}
}

func TestGetQueuePosition(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	q := NewQueue(client)

	// Add players with different skill ranks
	q.JoinQueue(1, "player1", 1000, 1400) // Lowest skill
	q.JoinQueue(2, "player2", 2000, 1600) // Highest skill
	q.JoinQueue(3, "player3", 1500, 1500) // Middle skill

	// Check positions (sorted by skill rank)
	pos, err := q.GetQueuePosition(1)
	if err != nil {
		t.Errorf("GetQueuePosition failed: %v", err)
	}

	if pos != 1 {
		t.Errorf("Expected position 1 for player1, got %d", pos)
	}

	pos, _ = q.GetQueuePosition(3)
	if pos != 2 {
		t.Errorf("Expected position 2 for player3, got %d", pos)
	}

	pos, _ = q.GetQueuePosition(2)
	if pos != 3 {
		t.Errorf("Expected position 3 for player2, got %d", pos)
	}
}

func TestGetQueuePositionNotInQueue(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	q := NewQueue(client)

	_, err := q.GetQueuePosition(999)
	if err == nil {
		t.Error("GetQueuePosition should error for non-existent player")
	}
}

func TestFindMatch(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	q := NewQueue(client)

	// Add enough players for a match
	q.JoinQueue(1, "player1", 1000, 1500)
	q.JoinQueue(2, "player2", 2000, 1600)
	q.JoinQueue(3, "player3", 1500, 1400)

	// Find match for 2-4 players
	entries, err := q.FindMatch(2, 4)
	if err != nil {
		t.Errorf("FindMatch failed: %v", err)
	}

	if len(entries) < 2 {
		t.Errorf("Expected at least 2 players, got %d", len(entries))
	}

	// Queue should be empty after match
	size, _ := q.GetQueueSize()
	if size != 0 {
		t.Errorf("Expected empty queue after match, got size %d", size)
	}
}

func TestFindMatchNotEnoughPlayers(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	q := NewQueue(client)

	// Add only 1 player
	q.JoinQueue(1, "player1", 1000, 1500)

	// Try to find match requiring 2 players
	_, err := q.FindMatch(2, 4)
	if err == nil {
		t.Error("FindMatch should fail when not enough players")
	}
}

func TestFindMatchEmptyQueue(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	q := NewQueue(client)

	// Try to find match in empty queue
	_, err := q.FindMatch(2, 4)
	if err == nil {
		t.Error("FindMatch should fail for empty queue")
	}
}

func TestQueueEntryFields(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	q := NewQueue(client)

	// Join queue
	q.JoinQueue(1, "player1", 1000, 1500)

	// Find match to get entry
	entries, _ := q.FindMatch(1, 2)

	if len(entries) != 1 {
		t.Fatal("Expected 1 entry")
	}

	entry := entries[0]

	if entry.UserID != 1 {
		t.Errorf("Expected UserID 1, got %d", entry.UserID)
	}

	if entry.Username != "player1" {
		t.Errorf("Expected username 'player1', got '%s'", entry.Username)
	}

	if entry.Chips != 1000 {
		t.Errorf("Expected 1000 chips, got %d", entry.Chips)
	}

	if entry.SkillRank != 1500 {
		t.Errorf("Expected skill rank 1500, got %d", entry.SkillRank)
	}

	if entry.JoinedAt.IsZero() {
		t.Error("JoinedAt should not be zero")
	}
}

func TestFindMatchMaxPlayers(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	q := NewQueue(client)

	// Add more players than max
	for i := 1; i <= 10; i++ {
		q.JoinQueue(uint(i), "player", 1000, 1500)
	}

	// Find match with max 5 players
	entries, err := q.FindMatch(2, 5)
	if err != nil {
		t.Errorf("FindMatch failed: %v", err)
	}

	if len(entries) > 5 {
		t.Errorf("Expected max 5 players, got %d", len(entries))
	}

	// Should still have players in queue
	size, _ := q.GetQueueSize()
	if size == 0 {
		t.Error("Queue should not be empty")
	}
}
