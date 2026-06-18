// Command createadmin creates (or promotes) an admin user.
//
// Usage:
//
//	go run ./cmd/createadmin -username admin -email admin@example.com -password 'StrongPass123!'
//
// If a user with the given username already exists, they are promoted to admin
// (password is left unchanged). Database connection settings are read from the
// same environment variables as the server (POSTGRES_*).
package main

import (
	"flag"
	"fmt"
	"os"

	"go-poker-arena/internal/auth"
	"go-poker-arena/internal/config"
	"go-poker-arena/internal/database"
	"go-poker-arena/internal/models"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	username := flag.String("username", "", "admin username (required)")
	email := flag.String("email", "", "admin email (required for new users)")
	password := flag.String("password", "", "admin password (required for new users)")
	flag.Parse()

	if *username == "" {
		fmt.Println("error: -username is required")
		flag.Usage()
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("config error: %v\n", err)
		os.Exit(1)
	}

	db, err := database.Connect(cfg.DB.DSN(), database.PoolConfig{
		MaxOpenConns: cfg.DB.MaxOpenConns,
		MaxIdleConns: cfg.DB.MaxIdleConns,
		ConnMaxLife:  cfg.DB.ConnMaxLife,
	})
	if err != nil {
		fmt.Printf("database error: %v\n", err)
		os.Exit(1)
	}

	var user models.User
	err = db.Where("username = ?", *username).First(&user).Error
	if err == nil {
		// Existing user: promote to admin.
		if err := db.Model(&user).Update("is_admin", true).Error; err != nil {
			fmt.Printf("failed to promote user: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ user %q (id=%d) promoted to admin\n", user.Username, user.ID)
		return
	}

	// New user: requires email + password.
	if *email == "" || *password == "" {
		fmt.Println("error: -email and -password are required to create a new admin")
		os.Exit(1)
	}

	hashed, err := auth.HashPassword(*password)
	if err != nil {
		fmt.Printf("failed to hash password: %v\n", err)
		os.Exit(1)
	}

	admin := models.User{
		Username: *username,
		Email:    *email,
		Password: hashed,
		Chips:    1000,
		IsAdmin:  true,
	}
	if err := db.Create(&admin).Error; err != nil {
		fmt.Printf("failed to create admin: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ admin user %q (id=%d) created\n", admin.Username, admin.ID)
}
