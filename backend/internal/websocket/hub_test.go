package websocket

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

func TestNewHub(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	hub := NewHub(client)

	if hub == nil {
		t.Error("NewHub should not return nil")
	}

	if hub.Clients == nil {
		t.Error("Clients map should be initialized")
	}

	if hub.Rooms == nil {
		t.Error("Rooms map should be initialized")
	}

	if hub.Broadcast == nil {
		t.Error("Broadcast channel should be initialized")
	}

	if hub.Register == nil {
		t.Error("Register channel should be initialized")
	}

	if hub.Unregister == nil {
		t.Error("Unregister channel should be initialized")
	}

	if hub.Redis == nil {
		t.Error("Redis client should not be nil")
	}
}

func TestMessage(t *testing.T) {
	msg := &Message{
		Type:     "test",
		RoomID:   "room1",
		UserID:   1,
		Username: "player1",
		Payload:  map[string]interface{}{"key": "value"},
	}

	if msg.Type != "test" {
		t.Errorf("Expected type 'test', got '%s'", msg.Type)
	}

	if msg.RoomID != "room1" {
		t.Errorf("Expected RoomID 'room1', got '%s'", msg.RoomID)
	}

	if msg.UserID != 1 {
		t.Errorf("Expected UserID 1, got %d", msg.UserID)
	}

	if msg.Username != "player1" {
		t.Errorf("Expected Username 'player1', got '%s'", msg.Username)
	}
}
