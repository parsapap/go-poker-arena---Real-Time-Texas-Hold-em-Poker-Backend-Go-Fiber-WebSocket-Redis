package rooms

import (
	"go-poker-arena/internal/models"
	"testing"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTest(t *testing.T) (*Manager, *miniredis.Miniredis) {
	// Setup database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	err = db.AutoMigrate(&models.Room{}, &models.User{})
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// Setup Redis
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	manager := NewManager(db, client)
	return manager, mr
}

func TestNewManager(t *testing.T) {
	manager, mr := setupTest(t)
	defer mr.Close()

	if manager == nil {
		t.Error("NewManager should not return nil")
	}

	if manager.DB == nil {
		t.Error("DB should not be nil")
	}

	if manager.Redis == nil {
		t.Error("Redis should not be nil")
	}

	if manager.Games == nil {
		t.Error("Games map should be initialized")
	}
}

func TestCreateRoom(t *testing.T) {
	manager, mr := setupTest(t)
	defer mr.Close()

	room, err := manager.CreateRoom("Test Room", 6, 10, 20)
	if err != nil {
		t.Errorf("CreateRoom failed: %v", err)
	}

	if room == nil {
		t.Fatal("Room should not be nil")
	}

	if room.Name != "Test Room" {
		t.Errorf("Expected name 'Test Room', got '%s'", room.Name)
	}

	if room.MaxPlayers != 6 {
		t.Errorf("Expected max players 6, got %d", room.MaxPlayers)
	}

	if room.SmallBlind != 10 {
		t.Errorf("Expected small blind 10, got %d", room.SmallBlind)
	}

	if room.BigBlind != 20 {
		t.Errorf("Expected big blind 20, got %d", room.BigBlind)
	}

	if room.Status != "waiting" {
		t.Errorf("Expected status 'waiting', got '%s'", room.Status)
	}
}

func TestGetRoom(t *testing.T) {
	manager, mr := setupTest(t)
	defer mr.Close()

	// Create room
	created, _ := manager.CreateRoom("Test Room", 6, 10, 20)

	// Get room
	room, err := manager.GetRoom(created.ID)
	if err != nil {
		t.Errorf("GetRoom failed: %v", err)
	}

	if room.ID != created.ID {
		t.Errorf("Expected ID %d, got %d", created.ID, room.ID)
	}

	if room.Name != "Test Room" {
		t.Errorf("Expected name 'Test Room', got '%s'", room.Name)
	}
}

func TestGetRoomNotFound(t *testing.T) {
	manager, mr := setupTest(t)
	defer mr.Close()

	// Try to get non-existent room
	_, err := manager.GetRoom(999)
	if err == nil {
		t.Error("GetRoom should fail for non-existent room")
	}
}

func TestListRooms(t *testing.T) {
	manager, mr := setupTest(t)
	defer mr.Close()

	// Create multiple rooms
	manager.CreateRoom("Room 1", 6, 10, 20)
	manager.CreateRoom("Room 2", 8, 20, 40)
	manager.CreateRoom("Room 3", 4, 5, 10)

	// List rooms
	rooms, err := manager.ListRooms()
	if err != nil {
		t.Errorf("ListRooms failed: %v", err)
	}

	if len(rooms) != 3 {
		t.Errorf("Expected 3 rooms, got %d", len(rooms))
	}
}

func TestListRoomsEmpty(t *testing.T) {
	manager, mr := setupTest(t)
	defer mr.Close()

	// List rooms when none exist
	rooms, err := manager.ListRooms()
	if err != nil {
		t.Errorf("ListRooms failed: %v", err)
	}

	if len(rooms) != 0 {
		t.Errorf("Expected 0 rooms, got %d", len(rooms))
	}
}

func TestJoinRoom(t *testing.T) {
	manager, mr := setupTest(t)
	defer mr.Close()

	// Create room
	room, _ := manager.CreateRoom("Test Room", 6, 10, 20)

	// Join room
	err := manager.JoinRoom(room.ID, 1)
	if err != nil {
		t.Errorf("JoinRoom failed: %v", err)
	}

	// Verify player was added
	// (In real implementation, would check Redis set)
}

func TestJoinRoomFull(t *testing.T) {
	manager, mr := setupTest(t)
	defer mr.Close()

	// Create room with max 2 players
	room, _ := manager.CreateRoom("Small Room", 2, 10, 20)

	// Add 2 players
	manager.JoinRoom(room.ID, 1)
	manager.JoinRoom(room.ID, 2)

	// Try to add 3rd player
	err := manager.JoinRoom(room.ID, 3)
	if err == nil {
		t.Error("JoinRoom should fail when room is full")
	}
}

func TestJoinRoomNotFound(t *testing.T) {
	manager, mr := setupTest(t)
	defer mr.Close()

	// Try to join non-existent room
	err := manager.JoinRoom(999, 1)
	if err == nil {
		t.Error("JoinRoom should fail for non-existent room")
	}
}

func TestCreateMultipleRooms(t *testing.T) {
	manager, mr := setupTest(t)
	defer mr.Close()

	// Create rooms with different configurations
	room1, _ := manager.CreateRoom("Low Stakes", 9, 5, 10)
	room2, _ := manager.CreateRoom("High Stakes", 6, 100, 200)

	if room1.ID == room2.ID {
		t.Error("Rooms should have different IDs")
	}

	if room1.SmallBlind == room2.SmallBlind {
		t.Error("Rooms should have different blinds")
	}
}

func TestGetRoomFromCache(t *testing.T) {
	manager, mr := setupTest(t)
	defer mr.Close()

	// Create room (will be cached)
	created, _ := manager.CreateRoom("Test Room", 6, 10, 20)

	// Get room first time (from cache)
	room1, _ := manager.GetRoom(created.ID)

	// Get room second time (should also be from cache)
	room2, _ := manager.GetRoom(created.ID)

	if room1.ID != room2.ID {
		t.Error("Should get same room from cache")
	}
}
