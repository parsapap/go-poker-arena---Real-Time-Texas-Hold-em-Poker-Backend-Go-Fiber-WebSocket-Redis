package middleware

import (
	"go-poker-arena/internal/models"
	"testing"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	err = db.AutoMigrate(&models.User{})
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	return db
}

func TestNewAdminMiddleware(t *testing.T) {
	db := setupTestDB(t)
	am := NewAdminMiddleware(db)

	if am == nil {
		t.Error("NewAdminMiddleware should not return nil")
	}

	if am.DB == nil {
		t.Error("DB should not be nil")
	}
}

func TestRequireAdminMiddleware(t *testing.T) {
	db := setupTestDB(t)
	am := NewAdminMiddleware(db)

	// Create admin user
	admin := &models.User{
		Username: "admin",
		IsAdmin:  true,
	}
	db.Create(admin)

	// Create regular user
	user := &models.User{
		Username: "user",
		IsAdmin:  false,
	}
	db.Create(user)

	// Test that middleware is created
	handler := am.RequireAdmin()
	if handler == nil {
		t.Error("RequireAdmin should return a handler")
	}
}

func TestCheckBannedMiddleware(t *testing.T) {
	db := setupTestDB(t)
	am := NewAdminMiddleware(db)

	// Create banned user
	banned := &models.User{
		Username: "banned",
		IsBanned: true,
	}
	db.Create(banned)

	// Create regular user
	user := &models.User{
		Username: "user",
		IsBanned: false,
	}
	db.Create(user)

	// Test that middleware is created
	handler := am.CheckBanned()
	if handler == nil {
		t.Error("CheckBanned should return a handler")
	}
}
