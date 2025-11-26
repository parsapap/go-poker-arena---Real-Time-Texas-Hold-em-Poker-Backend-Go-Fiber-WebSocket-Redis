package matchmaking

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"github.com/redis/go-redis/v9"
	"go-poker-arena/internal/models"
)

type Queue struct {
	Redis *redis.Client
}

func NewQueue(redisClient *redis.Client) *Queue {
	return &Queue{Redis: redisClient}
}

type QueueEntry struct {
	UserID    uint      `json:"user_id"`
	Username  string    `json:"username"`
	Chips     int64     `json:"chips"`
	SkillRank int64     `json:"skill_rank"`
	JoinedAt  time.Time `json:"joined_at"`
}

// JoinQueue adds player to matchmaking queue
func (q *Queue) JoinQueue(userID uint, username string, chips, skillRank int64) error {
	ctx := context.Background()
	
	entry := QueueEntry{
		UserID:    userID,
		Username:  username,
		Chips:     chips,
		SkillRank: skillRank,
		JoinedAt:  time.Now(),
	}
	
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	
	// Add to sorted set by skill rank
	return q.Redis.ZAdd(ctx, "matchmaking:queue", redis.Z{
		Score:  float64(skillRank),
		Member: string(data),
	}).Err()
}

// LeaveQueue removes player from queue
func (q *Queue) LeaveQueue(userID uint) error {
	ctx := context.Background()
	
	// Find and remove player's entry
	members, err := q.Redis.ZRange(ctx, "matchmaking:queue", 0, -1).Result()
	if err != nil {
		return err
	}
	
	for _, member := range members {
		var entry QueueEntry
		if err := json.Unmarshal([]byte(member), &entry); err != nil {
			continue
		}
		
		if entry.UserID == userID {
			return q.Redis.ZRem(ctx, "matchmaking:queue", member).Err()
		}
	}
	
	return nil
}

// FindMatch finds suitable players for a game
func (q *Queue) FindMatch(minPlayers, maxPlayers int) ([]QueueEntry, error) {
	ctx := context.Background()
	
	// Get players from queue
	members, err := q.Redis.ZRange(ctx, "matchmaking:queue", 0, int64(maxPlayers-1)).Result()
	if err != nil {
		return nil, err
	}
	
	if len(members) < minPlayers {
		return nil, fmt.Errorf("not enough players in queue")
	}
	
	entries := make([]QueueEntry, 0, len(members))
	for _, member := range members {
		var entry QueueEntry
		if err := json.Unmarshal([]byte(member), &entry); err != nil {
			continue
		}
		entries = append(entries, entry)
	}
	
	// Remove matched players from queue
	if len(entries) >= minPlayers {
		for _, member := range members[:len(entries)] {
			q.Redis.ZRem(ctx, "matchmaking:queue", member)
		}
	}
	
	return entries, nil
}

// GetQueueSize returns number of players in queue
func (q *Queue) GetQueueSize() (int64, error) {
	ctx := context.Background()
	return q.Redis.ZCard(ctx, "matchmaking:queue").Result()
}

// GetQueuePosition returns player's position in queue
func (q *Queue) GetQueuePosition(userID uint) (int64, error) {
	ctx := context.Background()
	
	members, err := q.Redis.ZRange(ctx, "matchmaking:queue", 0, -1).Result()
	if err != nil {
		return 0, err
	}
	
	for i, member := range members {
		var entry QueueEntry
		if err := json.Unmarshal([]byte(member), &entry); err != nil {
			continue
		}
		
		if entry.UserID == userID {
			return int64(i + 1), nil
		}
	}
	
	return 0, fmt.Errorf("player not in queue")
}

// AutoMatchWorker continuously looks for matches
func (q *Queue) AutoMatchWorker(interval time.Duration, roomManager interface {
	CreateRoom(name string, maxPlayers int, smallBlind, bigBlind int64) (*models.Room, error)
	JoinRoom(roomID, userID uint) error
	StartGame(roomID uint) (interface{}, error)
}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	
	for range ticker.C {
		entries, err := q.FindMatch(2, 9)
		if err != nil {
			continue
		}
		
		// Create room for matched players
		room, err := roomManager.CreateRoom(
			fmt.Sprintf("Auto-Match-%d", time.Now().Unix()),
			9,
			10,
			20,
		)
		if err != nil {
			continue
		}
		
		// Add players to room
		for _, entry := range entries {
			roomManager.JoinRoom(room.ID, entry.UserID)
		}
		
		// Start game if enough players
		if len(entries) >= 2 {
			roomManager.StartGame(room.ID)
		}
	}
}
