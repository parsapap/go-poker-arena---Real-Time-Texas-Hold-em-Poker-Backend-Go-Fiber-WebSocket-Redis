package auth

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

	// Migrate tables
	err = db.AutoMigrate(&models.User{}, &models.BanRecord{})
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	return db
}

func TestHashPassword(t *testing.T) {
	password := "testpassword123"
	hash, err := HashPassword(password)
	
	if err != nil {
		t.Errorf("HashPassword failed: %v", err)
	}
	
	if hash == "" {
		t.Error("Hash should not be empty")
	}
	
	if hash == password {
		t.Error("Hash should not equal plain password")
	}
}

func TestCheckPassword(t *testing.T) {
	password := "testpassword123"
	hash, _ := HashPassword(password)
	
	// Test correct password
	if !CheckPassword(password, hash) {
		t.Error("CheckPassword should return true for correct password")
	}
	
	// Test incorrect password
	if CheckPassword("wrongpassword", hash) {
		t.Error("CheckPassword should return false for incorrect password")
	}
}

func TestSignup(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)
	
	// Test successful signup
	user, err := service.Signup("testuser", "test@example.com", "password123")
	if err != nil {
		t.Errorf("Signup failed: %v", err)
	}
	
	if user == nil {
		t.Fatal("User should not be nil")
	}
	
	if user.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", user.Username)
	}
	
	if user.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", user.Email)
	}
	
	if user.Chips != 1000 {
		t.Errorf("Expected 1000 chips, got %d", user.Chips)
	}
	
	if user.IsAdmin {
		t.Error("New user should not be admin")
	}
	
	if user.IsBanned {
		t.Error("New user should not be banned")
	}
	
	// Test duplicate username
	_, err = service.Signup("testuser", "other@example.com", "password123")
	if err == nil {
		t.Error("Signup should fail for duplicate username")
	}
	
	// Test duplicate email
	_, err = service.Signup("otheruser", "test@example.com", "password123")
	if err == nil {
		t.Error("Signup should fail for duplicate email")
	}
}

func TestLogin(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)
	
	// Create test user
	service.Signup("testuser", "test@example.com", "password123")
	
	// Test successful login
	user, err := service.Login("testuser", "password123")
	if err != nil {
		t.Errorf("Login failed: %v", err)
	}
	
	if user == nil {
		t.Fatal("User should not be nil")
	}
	
	if user.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", user.Username)
	}
	
	// Test wrong password
	_, err = service.Login("testuser", "wrongpassword")
	if err == nil {
		t.Error("Login should fail with wrong password")
	}
	
	// Test non-existent user
	_, err = service.Login("nonexistent", "password123")
	if err == nil {
		t.Error("Login should fail for non-existent user")
	}
}

func TestLoginBannedUser(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)
	
	// Create and ban user
	user, _ := service.Signup("banneduser", "banned@example.com", "password123")
	service.BanUser(user.ID, 1, "test ban", true)
	
	// Test login with banned user
	_, err := service.Login("banneduser", "password123")
	if err == nil {
		t.Error("Login should fail for banned user")
	}
	
	if err.Error() != "user is banned" {
		t.Errorf("Expected 'user is banned' error, got '%v'", err)
	}
}

func TestGetUser(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)
	
	// Create test user
	created, _ := service.Signup("testuser", "test@example.com", "password123")
	
	// Test get existing user
	user, err := service.GetUser(created.ID)
	if err != nil {
		t.Errorf("GetUser failed: %v", err)
	}
	
	if user.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", user.Username)
	}
	
	// Test get non-existent user
	_, err = service.GetUser(9999)
	if err == nil {
		t.Error("GetUser should fail for non-existent user")
	}
}

func TestUpdateUser(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)
	
	// Create test user
	user, _ := service.Signup("testuser", "test@example.com", "password123")
	
	// Update user
	user.Chips = 2000
	user.Wins = 5
	err := service.UpdateUser(user)
	if err != nil {
		t.Errorf("UpdateUser failed: %v", err)
	}
	
	// Verify update
	updated, _ := service.GetUser(user.ID)
	if updated.Chips != 2000 {
		t.Errorf("Expected 2000 chips, got %d", updated.Chips)
	}
	
	if updated.Wins != 5 {
		t.Errorf("Expected 5 wins, got %d", updated.Wins)
	}
}

func TestBanUser(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)
	
	// Create test users
	user, _ := service.Signup("testuser", "test@example.com", "password123")
	admin, _ := service.Signup("admin", "admin@example.com", "password123")
	
	// Ban user
	err := service.BanUser(user.ID, admin.ID, "test reason", true)
	if err != nil {
		t.Errorf("BanUser failed: %v", err)
	}
	
	// Verify ban
	banned, _ := service.GetUser(user.ID)
	if !banned.IsBanned {
		t.Error("User should be banned")
	}
	
	// Verify ban record
	var banRecord models.BanRecord
	db.Where("user_id = ?", user.ID).First(&banRecord)
	if banRecord.Reason != "test reason" {
		t.Errorf("Expected reason 'test reason', got '%s'", banRecord.Reason)
	}
	
	if !banRecord.Permanent {
		t.Error("Ban should be permanent")
	}
}

func TestUnbanUser(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)
	
	// Create and ban user
	user, _ := service.Signup("testuser", "test@example.com", "password123")
	service.BanUser(user.ID, 1, "test ban", true)
	
	// Unban user
	err := service.UnbanUser(user.ID)
	if err != nil {
		t.Errorf("UnbanUser failed: %v", err)
	}
	
	// Verify unban
	unbanned, _ := service.GetUser(user.ID)
	if unbanned.IsBanned {
		t.Error("User should not be banned")
	}
}

func TestNewService(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)
	
	if service == nil {
		t.Error("NewService should not return nil")
	}
	
	if service.DB == nil {
		t.Error("Service DB should not be nil")
	}
}
