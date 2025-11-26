package rooms

import (
	"context"
	"encoding/json"
	"fmt"
	"go-poker-arena/internal/models"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Manager struct {
	DB    *gorm.DB
	Redis *redis.Client
}

func NewManager(db *gorm.DB, redisClient *redis.Client) *Manager {
	return &Manager{
		DB:    db,
		Redis: redisClient,
	}
}

func (m *Manager) CreateRoom(name string, maxPlayers int, smallBlind, bigBlind int64) (*models.Room, error) {
	room := &models.Room{
		Name:       name,
		MaxPlayers: maxPlayers,
		SmallBlind: smallBlind,
		BigBlind:   bigBlind,
		Status:     "waiting",
	}

	if err := m.DB.Create(room).Error; err != nil {
		return nil, err
	}

	ctx := context.Background()
	roomData, _ := json.Marshal(room)
	m.Redis.Set(ctx, fmt.Sprintf("room:%d", room.ID), roomData, 0)

	return room, nil
}

func (m *Manager) GetRoom(roomID uint) (*models.Room, error) {
	ctx := context.Background()
	key := fmt.Sprintf("room:%d", roomID)

	val, err := m.Redis.Get(ctx, key).Result()
	if err == nil {
		var room models.Room
		if err := json.Unmarshal([]byte(val), &room); err == nil {
			return &room, nil
		}
	}

	var room models.Room
	if err := m.DB.First(&room, roomID).Error; err != nil {
		return nil, err
	}

	roomData, _ := json.Marshal(room)
	m.Redis.Set(ctx, key, roomData, 0)

	return &room, nil
}

func (m *Manager) ListRooms() ([]models.Room, error) {
	var rooms []models.Room
	if err := m.DB.Where("status != ?", "finished").Find(&rooms).Error; err != nil {
		return nil, err
	}
	return rooms, nil
}

func (m *Manager) JoinRoom(roomID, userID uint) error {
	ctx := context.Background()
	key := fmt.Sprintf("room:%d:players", roomID)
	
	count, err := m.Redis.SCard(ctx, key).Result()
	if err != nil {
		return err
	}

	room, err := m.GetRoom(roomID)
	if err != nil {
		return err
	}

	if int(count) >= room.MaxPlayers {
		return fmt.Errorf("room is full")
	}

	return m.Redis.SAdd(ctx, key, userID).Err()
}

func (m *Manager) LeaveRoom(roomID, userID uint) error {
	ctx := context.Background()
	key := fmt.Sprintf("room:%d:players", roomID)
	return m.Redis.SRem(ctx, key, userID).Err()
}

func (m *Manager) GetRoomPlayers(roomID uint) ([]uint, error) {
	ctx := context.Background()
	key := fmt.Sprintf("room:%d:players", roomID)
	
	members, err := m.Redis.SMembers(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var playerIDs []uint
	for _, member := range members {
		var id uint
		fmt.Sscanf(member, "%d", &id)
		playerIDs = append(playerIDs, id)
	}

	return playerIDs, nil
}
