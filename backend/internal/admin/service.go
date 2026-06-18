// Package admin provides the business logic and HTTP handlers for the platform
// administration panel: user/room/game management, monitoring, and system
// controls. Handlers (handlers.go) stay thin and delegate to Service.
package admin

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go-poker-arena/internal/models"
	"go-poker-arena/internal/rooms"
	"gorm.io/gorm"

	"github.com/redis/go-redis/v9"
)

// ErrNotFound is returned when a requested entity does not exist.
var ErrNotFound = errors.New("not found")

// Service holds the dependencies needed by admin operations.
type Service struct {
	DB    *gorm.DB
	Redis *redis.Client
	Rooms *rooms.Manager
}

// NewService constructs an admin Service.
func NewService(db *gorm.DB, redisClient *redis.Client, roomManager *rooms.Manager) *Service {
	return &Service{DB: db, Redis: redisClient, Rooms: roomManager}
}

// Pagination is a normalized limit/offset pair derived from query params.
type Pagination struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// NewPagination clamps user-supplied paging params to safe bounds.
func NewPagination(limit, offset int) Pagination {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return Pagination{Limit: limit, Offset: offset}
}

// UserListResult is a paginated list of users.
type UserListResult struct {
	Users  []models.User `json:"users"`
	Total  int64         `json:"total"`
	Limit  int           `json:"limit"`
	Offset int           `json:"offset"`
}

// UserFilter describes the optional filters for ListUsers.
type UserFilter struct {
	Search string // matches username or email (case-insensitive, partial)
	Banned *bool  // nil = any, true = only banned, false = only active
}

// ListUsers returns a paginated, filterable list of users.
func (s *Service) ListUsers(f UserFilter, p Pagination) (*UserListResult, error) {
	q := s.DB.Model(&models.User{})

	if search := strings.TrimSpace(f.Search); search != "" {
		like := "%" + strings.ToLower(search) + "%"
		q = q.Where("LOWER(username) LIKE ? OR LOWER(email) LIKE ?", like, like)
	}
	if f.Banned != nil {
		q = q.Where("is_banned = ?", *f.Banned)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}

	var users []models.User
	if err := q.Order("id DESC").Limit(p.Limit).Offset(p.Offset).Find(&users).Error; err != nil {
		return nil, err
	}

	return &UserListResult{Users: users, Total: total, Limit: p.Limit, Offset: p.Offset}, nil
}

// UserDetail bundles a user with their aggregate stats and recent games.
type UserDetail struct {
	User        models.User           `json:"user"`
	TotalGames  int64                 `json:"total_games"`
	WinRate     float64               `json:"win_rate"`
	RecentGames []models.GameHistory  `json:"recent_games"`
	ActiveBans  []models.BanRecord    `json:"active_bans"`
}

// GetUserDetail returns a user with stats and recent game history.
func (s *Service) GetUserDetail(userID uint) (*UserDetail, error) {
	var user models.User
	if err := s.DB.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	recent, err := s.GetUserHistory(userID, 10)
	if err != nil {
		return nil, err
	}

	var totalGames int64
	s.DB.Model(&models.GameHistory{}).
		Where("players::jsonb @> ?", fmt.Sprintf(`[{"id":%d}]`, userID)).
		Count(&totalGames)

	winRate := 0.0
	if totalGames > 0 {
		winRate = float64(user.Wins) / float64(totalGames) * 100
	}

	var bans []models.BanRecord
	s.DB.Where("user_id = ?", userID).Order("created_at DESC").Limit(10).Find(&bans)

	return &UserDetail{
		User:        user,
		TotalGames:  totalGames,
		WinRate:     winRate,
		RecentGames: recent,
		ActiveBans:  bans,
	}, nil
}

// GetUserHistory returns a user's most recent games.
func (s *Service) GetUserHistory(userID uint, limit int) ([]models.GameHistory, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	var hist []models.GameHistory
	err := s.DB.
		Where("players::jsonb @> ?", fmt.Sprintf(`[{"id":%d}]`, userID)).
		Order("created_at DESC").
		Limit(limit).
		Find(&hist).Error
	return hist, err
}

// BanUser bans a user, recording the reason, the acting admin, and an optional
// expiry (temporary ban). A non-positive duration means a permanent ban.
func (s *Service) BanUser(userID, adminID uint, reason string, duration time.Duration) error {
	var user models.User
	if err := s.DB.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	if user.IsAdmin {
		return errors.New("cannot ban an admin user")
	}

	permanent := duration <= 0
	var expiresAt *time.Time
	if !permanent {
		t := time.Now().Add(duration)
		expiresAt = &t
	}

	return s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.User{}).Where("id = ?", userID).Update("is_banned", true).Error; err != nil {
			return err
		}
		return tx.Create(&models.BanRecord{
			UserID:    userID,
			AdminID:   adminID,
			Reason:    reason,
			Permanent: permanent,
			ExpiresAt: expiresAt,
		}).Error
	})
}

// UnbanUser clears a user's banned flag.
func (s *Service) UnbanUser(userID uint) error {
	res := s.DB.Model(&models.User{}).Where("id = ?", userID).Update("is_banned", false)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// RoomFilter describes optional filters for ListRooms.
type RoomFilter struct {
	Status      string // waiting, playing, finished, or "" for any
	MinBigBlind int64
}

// RoomInfo augments a room with its current player count.
type RoomInfo struct {
	models.Room
	PlayerCount int  `json:"player_count"`
	HasLiveGame bool `json:"has_live_game"`
}

// RoomListResult is a paginated list of rooms.
type RoomListResult struct {
	Rooms  []RoomInfo `json:"rooms"`
	Total  int64      `json:"total"`
	Limit  int        `json:"limit"`
	Offset int        `json:"offset"`
}

// ListRooms returns a paginated, filterable list of all rooms (including
// finished ones, unlike the public endpoint).
func (s *Service) ListRooms(f RoomFilter, p Pagination) (*RoomListResult, error) {
	q := s.DB.Model(&models.Room{})
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.MinBigBlind > 0 {
		q = q.Where("big_blind >= ?", f.MinBigBlind)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}

	var roomRows []models.Room
	if err := q.Order("id DESC").Limit(p.Limit).Offset(p.Offset).Find(&roomRows).Error; err != nil {
		return nil, err
	}

	infos := make([]RoomInfo, 0, len(roomRows))
	for _, r := range roomRows {
		players, _ := s.Rooms.GetRoomPlayers(r.ID)
		_, hasGame := s.Rooms.GetGame(r.ID)
		infos = append(infos, RoomInfo{
			Room:        r,
			PlayerCount: len(players),
			HasLiveGame: hasGame == nil,
		})
	}

	return &RoomListResult{Rooms: infos, Total: total, Limit: p.Limit, Offset: p.Offset}, nil
}

// RoomDetail bundles a room with player IDs and (optionally) live game state.
type RoomDetail struct {
	Room      models.Room            `json:"room"`
	PlayerIDs []uint                 `json:"player_ids"`
	GameState map[string]interface{} `json:"game_state,omitempty"`
}

// GetRoomDetail returns a room plus its membership and live game state.
func (s *Service) GetRoomDetail(roomID uint) (*RoomDetail, error) {
	room, err := s.Rooms.GetRoom(roomID)
	if err != nil {
		return nil, ErrNotFound
	}
	players, _ := s.Rooms.GetRoomPlayers(roomID)
	detail := &RoomDetail{Room: *room, PlayerIDs: players}
	if state, err := s.Rooms.GameState(roomID); err == nil {
		detail.GameState = state
	}
	return detail, nil
}

// Dashboard is the aggregated overview shown on the admin landing page.
type Dashboard struct {
	TotalUsers       int64 `json:"total_users"`
	BannedUsers      int64 `json:"banned_users"`
	ActiveRooms      int64 `json:"active_rooms"`
	GamesInProgress  int   `json:"games_in_progress"`
	HandsToday       int64 `json:"hands_today"`
	GamesToday       int64 `json:"games_today"`
	OnlinePlayers    int64 `json:"online_players"`
	MaintenanceMode  bool  `json:"maintenance_mode"`
}

// GetDashboard computes the overview statistics.
func (s *Service) GetDashboard() (*Dashboard, error) {
	d := &Dashboard{}

	s.DB.Model(&models.User{}).Count(&d.TotalUsers)
	s.DB.Model(&models.User{}).Where("is_banned = ?", true).Count(&d.BannedUsers)
	s.DB.Model(&models.Room{}).Where("status IN ?", []string{"waiting", "playing"}).Count(&d.ActiveRooms)

	startOfDay := time.Now().Truncate(24 * time.Hour)
	s.DB.Model(&models.GameHistory{}).Where("created_at >= ?", startOfDay).Count(&d.GamesToday)
	// One game == one hand in this engine.
	d.HandsToday = d.GamesToday

	d.GamesInProgress = s.Rooms.LiveGameCount()

	// Online players: distinct active WebSocket connection counters in Redis.
	d.OnlinePlayers = s.countOnlinePlayers()
	d.MaintenanceMode = s.maintenanceEnabled()

	return d, nil
}

// LiveGames returns lightweight info on all in-progress games.
func (s *Service) LiveGames() []rooms.LiveGameInfo {
	return s.Rooms.ListLiveGames()
}

// ForceEndGame ends a room's current game (admin action).
func (s *Service) ForceEndGame(roomID uint) error {
	return s.Rooms.ForceEndGame(roomID)
}

// KickPlayer removes a player from a room.
func (s *Service) KickPlayer(roomID, userID uint) error {
	return s.Rooms.KickPlayer(roomID, userID)
}

// CloseRoom force-closes a room.
func (s *Service) CloseRoom(roomID uint) error {
	return s.Rooms.CloseRoom(roomID)
}

// SystemHealth reports detailed dependency health.
type SystemHealth struct {
	Status          string `json:"status"`
	Database        bool   `json:"database"`
	Redis           bool   `json:"redis"`
	MaintenanceMode bool   `json:"maintenance_mode"`
	GamesInProgress int    `json:"games_in_progress"`
}

// GetSystemHealth probes DB and Redis and returns a detailed status.
func (s *Service) GetSystemHealth() *SystemHealth {
	h := &SystemHealth{Status: "ok", MaintenanceMode: s.maintenanceEnabled()}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if sqlDB, err := s.DB.DB(); err == nil && sqlDB.PingContext(ctx) == nil {
		h.Database = true
	}
	if s.Redis != nil && s.Redis.Ping(ctx).Err() == nil {
		h.Redis = true
	}
	h.GamesInProgress = s.Rooms.LiveGameCount()

	if !h.Database || !h.Redis {
		h.Status = "degraded"
	}
	return h
}

func (s *Service) maintenanceEnabled() bool {
	if s.Redis == nil {
		return false
	}
	val, err := s.Redis.Get(context.Background(), "system:maintenance").Result()
	return err == nil && val == "1"
}

// countOnlinePlayers counts active per-user WebSocket connection counters.
func (s *Service) countOnlinePlayers() int64 {
	if s.Redis == nil {
		return 0
	}
	ctx := context.Background()
	var count int64
	var cursor uint64
	for {
		keys, next, err := s.Redis.Scan(ctx, cursor, "ws:connections:u:*", 100).Result()
		if err != nil {
			break
		}
		for _, k := range keys {
			if v, err := s.Redis.Get(ctx, k).Int(); err == nil && v > 0 {
				count++
			}
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return count
}
