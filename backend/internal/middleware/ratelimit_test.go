package middleware

import (
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

func TestNewRateLimiter(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	rl := NewRateLimiter(client)

	if rl == nil {
		t.Error("NewRateLimiter should not return nil")
	}

	if rl.Redis == nil {
		t.Error("Redis client should not be nil")
	}
}

func TestDecrementWSConnection(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	rl := NewRateLimiter(client)

	// This should not panic
	rl.DecrementWSConnection("user1")
}
