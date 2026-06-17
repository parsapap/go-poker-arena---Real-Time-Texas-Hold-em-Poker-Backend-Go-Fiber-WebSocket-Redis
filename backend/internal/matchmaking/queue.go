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

const (
	queueKey   = "matchmaking:queue"
	membersKey = "matchmaking:members" // hash: userID -> queue member JSON
)

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

	// Add to sorted set by skill rank and record a userID -> member mapping so
	// the entry can later be removed in O(1) without scanning the whole set.
	pipe := q.Redis.Pipeline()
	pipe.ZAdd(ctx, queueKey, redis.Z{
		Score:  float64(skillRank),
		Member: string(data),
	})
	pipe.HSet(ctx, membersKey, fmt.Sprintf("%d", userID), string(data))
	_, err = pipe.Exec(ctx)
	return err
}

// LeaveQueue removes player from queue in O(log N) using the userID -> member
// mapping instead of scanning every entry.
func (q *Queue) LeaveQueue(userID uint) error {
	ctx := context.Background()
	field := fmt.Sprintf("%d", userID)

	member, err := q.Redis.HGet(ctx, membersKey, field).Result()
	if err == redis.Nil {
		return nil // not in queue
	}
	if err != nil {
		return err
	}

	pipe := q.Redis.Pipeline()
	pipe.ZRem(ctx, queueKey, member)
	pipe.HDel(ctx, membersKey, field)
	_, err = pipe.Exec(ctx)
	return err
}

// FindMatch finds suitable players for a game
func (q *Queue) FindMatch(minPlayers, maxPlayers int) ([]QueueEntry, error) {
	ctx := context.Background()

	// Get players from queue
	members, err := q.Redis.ZRange(ctx, queueKey, 0, int64(maxPlayers-1)).Result()
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

	// Remove matched players from both the sorted set and the member mapping.
	if len(entries) >= minPlayers {
		pipe := q.Redis.Pipeline()
		for i, member := range members {
			pipe.ZRem(ctx, queueKey, member)
			pipe.HDel(ctx, membersKey, fmt.Sprintf("%d", entries[i].UserID))
		}
		pipe.Exec(ctx)
	}

	return entries, nil
}

// GetQueueSize returns number of players in queue
func (q *Queue) GetQueueSize() (int64, error) {
	ctx := context.Background()
	return q.Redis.ZCard(ctx, queueKey).Result()
}

// GetQueuePosition returns player's position in queue
func (q *Queue) GetQueuePosition(userID uint) (int64, error) {
	ctx := context.Background()

	member, err := q.Redis.HGet(ctx, membersKey, fmt.Sprintf("%d", userID)).Result()
	if err == redis.Nil {
		return 0, fmt.Errorf("player not in queue")
	}
	if err != nil {
		return 0, err
	}

	rank, err := q.Redis.ZRank(ctx, queueKey, member).Result()
	if err != nil {
		return 0, fmt.Errorf("player not in queue")
	}
	return rank + 1, nil
}

// AutoMatchWorker continuously looks for matches until ctx is cancelled.
func (q *Queue) AutoMatchWorker(ctx context.Context, interval time.Duration, roomManager interface {
	CreateRoom(name string, maxPlayers int, smallBlind, bigBlind int64) (*models.Room, error)
	JoinRoom(roomID, userID uint) error
	StartGame(roomID uint) (*models.Room, error)
}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

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

		// Add players to the room. JoinRoom itself triggers the start
		// countdown once enough players have joined, so we must NOT also call
		// StartGame here — doing both previously caused the game to start
		// twice (a duplicate-start race).
		for _, entry := range entries {
			roomManager.JoinRoom(room.ID, entry.UserID)
		}
	}
}
