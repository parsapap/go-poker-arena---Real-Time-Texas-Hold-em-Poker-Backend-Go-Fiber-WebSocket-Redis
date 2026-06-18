package admin

import (
	"context"
	"testing"
	"time"

	"go-poker-arena/internal/models"
	"go-poker-arena/internal/rooms"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupService(t *testing.T) (*Service, *gorm.DB, *miniredis.Miniredis) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Room{}, &models.BanRecord{}, &models.GameHistory{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	rc := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	rm := rooms.NewManager(db, rc)
	return NewService(db, rc, rm), db, mr
}

func seedUsers(t *testing.T, db *gorm.DB) {
	t.Helper()
	users := []models.User{
		{Username: "alice", Email: "alice@example.com", Password: "x"},
		{Username: "bob", Email: "bob@example.com", Password: "x", IsBanned: true},
		{Username: "carol", Email: "carol@example.com", Password: "x"},
		{Username: "root", Email: "root@example.com", Password: "x", IsAdmin: true},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("seed users: %v", err)
	}
}

func TestListUsers_Pagination(t *testing.T) {
	svc, db, mr := setupService(t)
	defer mr.Close()
	seedUsers(t, db)

	res, err := svc.ListUsers(UserFilter{}, NewPagination(2, 0))
	if err != nil {
		t.Fatalf("ListUsers: %v", err)
	}
	if res.Total != 4 {
		t.Errorf("total = %d, want 4", res.Total)
	}
	if len(res.Users) != 2 {
		t.Errorf("page size = %d, want 2", len(res.Users))
	}
}

func TestListUsers_SearchAndBannedFilter(t *testing.T) {
	svc, db, mr := setupService(t)
	defer mr.Close()
	seedUsers(t, db)

	// Search
	res, err := svc.ListUsers(UserFilter{Search: "ali"}, NewPagination(20, 0))
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if res.Total != 1 || res.Users[0].Username != "alice" {
		t.Errorf("search 'ali' => %d results, want 1 (alice)", res.Total)
	}

	// Banned filter
	banned := true
	res, err = svc.ListUsers(UserFilter{Banned: &banned}, NewPagination(20, 0))
	if err != nil {
		t.Fatalf("banned filter: %v", err)
	}
	if res.Total != 1 || res.Users[0].Username != "bob" {
		t.Errorf("banned filter => %d results, want 1 (bob)", res.Total)
	}
}

func TestBanAndUnbanUser(t *testing.T) {
	svc, db, mr := setupService(t)
	defer mr.Close()
	seedUsers(t, db)

	var carol models.User
	db.Where("username = ?", "carol").First(&carol)

	// Temporary ban
	if err := svc.BanUser(carol.ID, 1, "spamming", 30*time.Minute); err != nil {
		t.Fatalf("BanUser: %v", err)
	}
	var reloaded models.User
	db.First(&reloaded, carol.ID)
	if !reloaded.IsBanned {
		t.Error("user should be banned")
	}
	var ban models.BanRecord
	if err := db.Where("user_id = ?", carol.ID).First(&ban).Error; err != nil {
		t.Fatalf("ban record not created: %v", err)
	}
	if ban.Permanent {
		t.Error("ban should be temporary")
	}
	if ban.ExpiresAt == nil {
		t.Error("temporary ban should have an expiry")
	}

	// Unban
	if err := svc.UnbanUser(carol.ID); err != nil {
		t.Fatalf("UnbanUser: %v", err)
	}
	db.First(&reloaded, carol.ID)
	if reloaded.IsBanned {
		t.Error("user should be unbanned")
	}
}

func TestBanUser_CannotBanAdmin(t *testing.T) {
	svc, db, mr := setupService(t)
	defer mr.Close()
	seedUsers(t, db)

	var root models.User
	db.Where("username = ?", "root").First(&root)
	if err := svc.BanUser(root.ID, 1, "nope", 0); err == nil {
		t.Error("expected error banning an admin")
	}
}

func TestBanUser_NotFound(t *testing.T) {
	svc, _, mr := setupService(t)
	defer mr.Close()
	if err := svc.BanUser(9999, 1, "x", 0); err == nil {
		t.Error("expected ErrNotFound for missing user")
	}
}

func TestListRooms_StatusFilter(t *testing.T) {
	svc, db, mr := setupService(t)
	defer mr.Close()
	db.Create(&[]models.Room{
		{Name: "A", Status: "waiting", MaxPlayers: 6, SmallBlind: 10, BigBlind: 20},
		{Name: "B", Status: "finished", MaxPlayers: 6, SmallBlind: 50, BigBlind: 100},
		{Name: "C", Status: "playing", MaxPlayers: 6, SmallBlind: 10, BigBlind: 20},
	})

	res, err := svc.ListRooms(RoomFilter{Status: "finished"}, NewPagination(20, 0))
	if err != nil {
		t.Fatalf("ListRooms: %v", err)
	}
	if res.Total != 1 || res.Rooms[0].Name != "B" {
		t.Errorf("status filter => %d, want 1 (B)", res.Total)
	}

	// Min big blind filter
	res, err = svc.ListRooms(RoomFilter{MinBigBlind: 100}, NewPagination(20, 0))
	if err != nil {
		t.Fatalf("ListRooms minbb: %v", err)
	}
	if res.Total != 1 {
		t.Errorf("min_big_blind filter => %d, want 1", res.Total)
	}
}

func TestMaintenanceAndHealth(t *testing.T) {
	svc, _, mr := setupService(t)
	defer mr.Close()

	if svc.maintenanceEnabled() {
		t.Error("maintenance should default off")
	}
	svc.Redis.Set(context.Background(), "system:maintenance", "1", 0)
	if !svc.maintenanceEnabled() {
		t.Error("maintenance should be on")
	}

	h := svc.GetSystemHealth()
	if !h.Redis {
		t.Error("redis should be healthy (miniredis)")
	}
	if !h.MaintenanceMode {
		t.Error("health should reflect maintenance mode")
	}
}

func TestNewPaginationClamps(t *testing.T) {
	if p := NewPagination(0, -5); p.Limit != 20 || p.Offset != 0 {
		t.Errorf("defaults wrong: %+v", p)
	}
	if p := NewPagination(500, 10); p.Limit != 100 {
		t.Errorf("limit should clamp to 100, got %d", p.Limit)
	}
}
